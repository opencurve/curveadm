/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/pkg/log"
)

const (
	POLICY_ALWAYS_RESTART = "always"
	POLICY_NEVER_RESTART  = "no"
	CLEANED_CONTAINER_ID  = "-"
)

type step2GetService struct {
	serviceId   string
	containerId *string
	storage     *storage.Storage
}

type step2InsertService struct {
	clusterId      int
	serviceId      string
	containerId    *string
	oldContainerId *string
	storage        *storage.Storage
}

func (s *step2GetService) Execute(ctx *context.Context) error {
	containerId, err := s.storage.GetContainerId(s.serviceId)

	if err != nil {
		return err
	} else if containerId == CLEANED_CONTAINER_ID { // "-" means container removed
		// do nothing
	} else if len(containerId) > 0 { // service already exist
		return task.ERR_BREAK_TASK
	}

	*s.containerId = containerId
	return nil
}

func (s *step2InsertService) Execute(ctx *context.Context) error {
	var err error
	serviceId := s.serviceId
	clusterId := s.clusterId
	oldContainerId := *s.oldContainerId
	containerId := *s.containerId
	if oldContainerId == CLEANED_CONTAINER_ID { // container cleaned
		err = s.storage.SetContainId(serviceId, containerId)
	} else {
		err = s.storage.InsertService(clusterId, serviceId, containerId)
	}

	log.SwitchLevel(err)("InsertService",
		log.Field("serviceId", serviceId),
		log.Field("containerId", containerId))
	return err

}

func getArguments(dc *topology.DeployConfig) string {
	role := dc.GetRole()
	if role != topology.ROLE_CHUNKSERVER {
		return ""
	}

	// only chunkserver need so many arguments, but who cares
	layout := dc.GetProjectLayout()
	dataDir := layout.ServiceDataDir
	chunkserverArguments := map[string]interface{}{
		// chunkserver
		"conf":                  layout.ServiceConfPath,
		"chunkServerIp":         dc.GetListenIp(),
		"enableExternalServer":  true,
		"chunkServerExternalIp": dc.GetListenExternalIp(),
		"chunkServerPort":       dc.GetListenPort(),
		"chunkFilePoolDir":      dataDir,
		"chunkFilePoolMetaPath": fmt.Sprintf("%s/chunkfilepool.meta", dataDir),
		"walFilePoolDir":        dataDir,
		"walFilePoolMetaPath":   fmt.Sprintf("%s/walfilepool.meta", dataDir),
		"copySetUri":            fmt.Sprintf("local://%s/copysets", dataDir),
		"recycleUri":            fmt.Sprintf("local://%s/recycler", dataDir),
		"raftLogUri":            fmt.Sprintf("curve://%s/copysets", dataDir),
		"raftSnapshotUri":       fmt.Sprintf("curve://%s/copysets", dataDir),
		"chunkServerStoreUri":   fmt.Sprintf("local://%s", dataDir),
		"chunkServerMetaUri":    fmt.Sprintf("local://%s/chunkserver.dat", dataDir),
		// brpc
		"bthread_concurrency":      18,
		"graceful_quit_on_sigterm": true,
		// raft
		"raft_sync":                            true,
		"raft_sync_meta":                       true,
		"raft_sync_segments":                   true,
		"raft_max_segment_size":                8388608,
		"raft_max_install_snapshot_tasks_num":  1,
		"raft_use_fsync_rather_than_fdatasync": false,
	}

	arguments := []string{}
	for k, v := range chunkserverArguments {
		arguments = append(arguments, fmt.Sprintf("-%s=%v", k, v))
	}
	return strings.Join(arguments, " ")
}

func getMountVolumes(dc *topology.DeployConfig) []step.Volume {
	volumes := []step.Volume{}
	layout := dc.GetProjectLayout()
	logDir := dc.GetLogDir()
	dataDir := dc.GetDataDir()
	coreDir := dc.GetCoreDir()

	if len(logDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      logDir,
			ContainerPath: layout.ServiceLogDir,
		})
	}

	if len(dataDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      dataDir,
			ContainerPath: layout.ServiceDataDir,
		})
	}

	if len(coreDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      coreDir,
			ContainerPath: layout.CoreSystemDir,
		})
	}

	return volumes
}

func getRestartPolicy(dc *topology.DeployConfig) string {
	switch dc.GetRole() {
	case topology.ROLE_ETCD:
		return POLICY_ALWAYS_RESTART
	case topology.ROLE_MDS:
		return POLICY_ALWAYS_RESTART
	}
	return POLICY_NEVER_RESTART
}

func NewCreateContainerTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Create Container", subname, dc.GetSSHConfig())

	// add step
	var oldContainerId, containerId string
	clusterId := curveadm.ClusterId()
	dcId := dc.GetId()
	serviceId := curveadm.GetServiceId(dcId)
	t.AddStep(&step2GetService{ // if service exist, break task
		serviceId:   serviceId,
		containerId: &oldContainerId,
		storage:     curveadm.Storage(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:        []string{dc.GetLogDir(), dc.GetDataDir()},
		ExecWithSudo: false,
		ExecInLocal:  false,
	})
	t.AddStep(&step.CreateContainer{
		Image:        dc.GetContainerImage(),
		Command:      fmt.Sprintf("--role %s --args='%s'", dc.GetRole(), getArguments(dc)),
		Envs:         []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Hostname:     fmt.Sprintf("%s-%s", dc.GetKind(), dc.GetRole()),
		Privileged:   true,
		Restart:      getRestartPolicy(dc),
		Ulimits:      []string{"core=-1"},
		Volumes:      getMountVolumes(dc),
		Out:          &containerId,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step2InsertService{
		clusterId:      clusterId,
		serviceId:      serviceId,
		containerId:    &containerId,
		oldContainerId: &oldContainerId,
		storage:        curveadm.Storage(),
	})

	return t, nil
}

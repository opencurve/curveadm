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

// __SIGN_BY_WINE93__

package common

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/disks"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	log "github.com/opencurve/curveadm/pkg/log/glg"
)

const (
	POLICY_ALWAYS_RESTART = "always"
	POLICY_NEVER_RESTART  = "no"
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
		return errno.ERR_GET_SERVICE_CONTAINER_ID_FAILED.E(err)
	} else if containerId == comm.CLEANED_CONTAINER_ID { // "-" means container removed
		// do nothing
	} else if len(containerId) > 0 {
		return task.ERR_SKIP_TASK
	}

	*s.containerId = containerId
	return nil
}

func (s *step2InsertService) E(e error, ec *errno.ErrorCode) error {
	if e == nil {
		return nil
	}
	return ec.E(e)
}

func (s *step2InsertService) Execute(ctx *context.Context) error {
	var err error
	serviceId := s.serviceId
	clusterId := s.clusterId
	oldContainerId := *s.oldContainerId
	containerId := *s.containerId
	if oldContainerId == comm.CLEANED_CONTAINER_ID { // container cleaned
		err = s.storage.SetContainId(serviceId, containerId)
		err = s.E(err, errno.ERR_SET_SERVICE_CONTAINER_ID_FAILED)
	} else {
		err = s.storage.InsertService(clusterId, serviceId, containerId)
		err = s.E(err, errno.ERR_INSERT_SERVICE_CONTAINER_ID_FAILED)
	}

	log.SwitchLevel(err)("Insert service",
		log.Field("ServiceId", serviceId),
		log.Field("ContainerId", containerId))

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
		"enableExternalServer":  dc.GetEnableExternalServer(),
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

func getEnvironments(dc *topology.DeployConfig) []string {
	preloads := []string{"/usr/local/lib/libjemalloc.so"}
	if dc.GetEnableRDMA() {
		preloads = append(preloads, "/usr/local/lib/libsmc-preload.so")
	}

	return []string{
		fmt.Sprintf("'LD_PRELOAD=%s'", strings.Join(preloads, " ")),
	}
}

func getMountVolumes(dc *topology.DeployConfig, serviceMountDevice bool) []step.Volume {
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

	// add volume binds if not directly mount disk device in container
	if len(dataDir) > 0 && !serviceMountDevice {
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

func getRestartPolicy(dc *topology.DeployConfig, serviceMountDevice bool) string {
	switch dc.GetRole() {
	case topology.ROLE_ETCD:
		return POLICY_ALWAYS_RESTART
	case topology.ROLE_MDS:
		return POLICY_ALWAYS_RESTART
	case topology.ROLE_CHUNKSERVER:
		if serviceMountDevice {
			return POLICY_ALWAYS_RESTART
		}
	}
	return POLICY_NEVER_RESTART
}

func trimContainerId(containerId *string) step.LambdaType {
	return func(ctx *context.Context) error {
		items := strings.Split(*containerId, "\n")
		*containerId = items[len(items)-1]
		return nil
	}
}

func NewCreateContainerTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Create Container", subname, hc.GetSSHConfig())

	// add step to task
	var oldContainerId, containerId string
	clusterId := curveadm.ClusterId()
	dcId := dc.GetId()
	serviceId := curveadm.GetServiceId(dcId)
	kind := dc.GetKind()
	role := dc.GetRole()
	hostname := fmt.Sprintf("%s-%s-%s", kind, role, serviceId)
	options := curveadm.ExecOptions()
	options.ExecWithSudo = false
	host := dc.GetHost()
	dataDir := dc.GetDataDir()

	device := ""
	extraParam := ""
	serviceMountDevice := false
	// update disk service(chunkserver) ID and get disk UUID for service device direct mounting
	if role == topology.ROLE_CHUNKSERVER && len(curveadm.DiskRecords()) > 0 {
		if err := curveadm.Storage().UpdateDiskChunkServerID(host, dataDir, serviceId); err != nil {
			return t, err
		}
		disk, err := curveadm.Storage().GetDiskByMountPoint(host, dataDir)
		if err != nil {
			return t, err
		}

		device = disk.Device
		serviceMountDevice = disk.ServiceMountDevice != 0
		if serviceMountDevice {
			diskId, diskUriProto, err := disks.GetDiskId(disk)
			if err != nil {
				return t, err
			}
			if diskUriProto == disks.DISK_URI_PROTO_FS_UUID {
				extraParam = fmt.Sprintf("--disk UUID=%s", diskId)
			}
		}
	}

	t.AddStep(&step2GetService{ // if service exist, break task
		serviceId:   serviceId,
		containerId: &oldContainerId,
		storage:     curveadm.Storage(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{dc.GetLogDir(), dc.GetDataDir()},
		ExecOptions: options,
	})
	t.AddStep(&step.CreateContainer{
		Image:       dc.GetContainerImage(),
		Command:     fmt.Sprintf("--role %s --args='%s' %s", role, getArguments(dc), extraParam),
		AddHost:     []string{fmt.Sprintf("%s:127.0.0.1", hostname)},
		Envs:        getEnvironments(dc),
		Hostname:    hostname,
		Init:        true,
		Name:        hostname,
		Privileged:  true,
		Restart:     getRestartPolicy(dc, serviceMountDevice),
		Ulimits:     []string{"core=-1"},
		Volumes:     getMountVolumes(dc, serviceMountDevice),
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: trimContainerId(&containerId),
	})
	t.AddStep(&step2InsertService{
		clusterId:      clusterId,
		serviceId:      serviceId,
		containerId:    &containerId,
		oldContainerId: &oldContainerId,
		storage:        curveadm.Storage(),
	})
	if serviceMountDevice && device != "" {
		t.AddStep(&step.UmountFilesystem{
			Directorys:     []string{device},
			IgnoreUmounted: false, // should not start service if failed to unmount disk from host
			IgnoreNotFound: false, // should not start service if disk was not found
			ExecOptions:    curveadm.ExecOptions(),
		})
	}

	return t, nil
}

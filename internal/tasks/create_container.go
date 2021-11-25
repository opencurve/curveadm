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

package tasks

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tasks/task"
)

const (
	POLICY_ALWAYS_RESTART = "always"
	POLICY_NEVER_RESTART  = "no"
)

type (
	step2CreateContainer struct {
		clusterId int
		storage   *storage.Storage
	}
)

func (s *step2CreateContainer) createDirectory(ctx *task.Context) error {
	config := ctx.Config()
	logDir := config.GetLogDir()
	dataDir := config.GetDataDir()
	if err := ctx.Module().SshCreateDir(logDir); err != nil {
		return err
	} else if err := ctx.Module().SshCreateDir(dataDir); err != nil {
		return err
	}
	return nil
}

// -v hostPath1:conatinerPath1 -v hostPath2:conatinerPath2
func (s *step2CreateContainer) volumeArgs(ctx *task.Context) string {
	volumes := []string{}
	config := ctx.Config()
	prefix := config.GetProjectPrefix()
	logDir := config.GetLogDir()
	dataDir := config.GetDataDir()

	if logDir != "" {
		hostPath := logDir
		containerPath := prefix + "/logs"
		volumes = append(volumes, fmt.Sprintf("-v %s:%s", hostPath, containerPath))
	}

	if dataDir != "" {
		hostPath := dataDir
		containerPath := prefix + "/data"
		volumes = append(volumes, fmt.Sprintf("-v %s:%s", hostPath, containerPath))
	}

	return strings.Join(volumes, " ")
}

// --restart always
func (s *step2CreateContainer) restartArgs(ctx *task.Context) string {
	config := ctx.Config()
	restartPolicy := POLICY_ALWAYS_RESTART
	if config.GetRole() == configure.ROLE_METASERVER {
		restartPolicy = POLICY_NEVER_RESTART
	}
	return fmt.Sprintf("--restart %s", restartPolicy)
}

func (s *step2CreateContainer) formatCommand(ctx *task.Context) string {
	return fmt.Sprintf("sudo docker create --network=host %s %s %s --role=%s 2>/dev/null",
		s.restartArgs(ctx),
		s.volumeArgs(ctx),
		ctx.Config().GetContainerImage(),
		ctx.Config().GetRole())
}

func (s *step2CreateContainer) Execute(ctx *task.Context) error {
	clusterId := s.clusterId
	config := ctx.Config()
	dcId := config.GetId()
	serviceId := configure.ServiceId(clusterId, dcId)
	oldContainerId, err := s.storage.GetContainerId(serviceId)
	if err != nil {
		return err
	} else if oldContainerId != "" && oldContainerId != "-" { // service already exist, "-" means container removed
		return nil
	}

	// container not exist
	if err := s.createDirectory(ctx); err != nil { // create directory
		return err
	} else if out, err := ctx.Module().SshShell(s.formatCommand(ctx)); err != nil { // create container
		return err
	} else { // insert container to sqlite
		containerId := utils.TrimNewline(out)
		if oldContainerId == "" { // not exist
			err = s.storage.InsertService(clusterId, serviceId, containerId)
		} else { // removed
			err = s.storage.SetContainId(serviceId, containerId)
		}

		log.SwitchLevel(err)("InsertService",
			log.Field("clusterId", clusterId),
			log.Field("dcId", dcId),
			log.Field("containerId", containerId))
		return err
	}

	return nil
}

func (s *step2CreateContainer) Rollback(ctx *task.Context) {
}

func NewCreateContainerTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Create Container", subname, dc)
	t.AddStep(&step2CreateContainer{
		clusterId: curveadm.ClusterId(),
		storage:   curveadm.Storage(),
	})
	return t, nil
}

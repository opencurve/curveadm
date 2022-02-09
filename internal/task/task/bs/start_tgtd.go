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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	DEFAULT_TGTD_CONTAINER_NAME = "curvebs-target-daemon"
)

type (
	step2CheckTargetDaemon struct{ containerId *string }
)

func (s *step2CheckTargetDaemon) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return task.ERR_SKIP_TASK
	}
	return nil
}

func NewStartTargetDaemonTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	subname := fmt.Sprintf("hostname=%s image=%s", cc.GetHost(), cc.GetContainerImage())
	t := task.NewTask("Start Target Daemon", subname, cc.GetSSHConfig())

	// add step
	var containerId string
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2CheckTargetDaemon{ // skip if target-daemon exist
		containerId: &containerId,
	})
	t.AddStep(&step.PullImage{
		Image:         cc.GetContainerImage(),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.CreateContainer{
		Image:         cc.GetContainerImage(),
		Command:       "-f",
		Entrypoint:    "tgtd",
		Envs:          []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Init:          true,
		Name:          DEFAULT_TGTD_CONTAINER_NAME,
		Pid:           "host",
		Privileged:    true,
		Volumes:       getVolumes(cc),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.SyncFile{ // sync nebd-client config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  "/curvebs/conf/nebd-client.conf",
		ContainerDestId:   &containerId,
		ContainerDestPath: "/etc/nebd/nebd-client.conf",
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.StartContainer{
		ContainerId:   &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})

	return t, nil
}

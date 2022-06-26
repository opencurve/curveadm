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
 * Created Date: 2022-06-26
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/playground"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/context"
)

type step2CheckContainer struct {
	containerId *string
}

func (s *step2CheckContainer) Execute(ctx *context.Context) error {
	if len(*s.containerId) == 0 {
		return task.ERR_SKIP_TASK
	}
	return nil
}

func NewRemovePlaygroundTask(curveadm *cli.CurveAdm, pc *playground.PlaygroundConfig) (*task.Task, error) {
	name := pc.GetName()
	mountPoint := pc.GetMointpoint()
	subname := fmt.Sprintf("name=%s mountPoint=%s", name, mountPoint)
	t := task.NewTask("Remove Playground", subname, nil)

	var containerId string
	containerName := name
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", containerName),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   true,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2CheckContainer{
		containerId: &containerId,
	})
	t.AddStep(&step.StopContainer{
		ContainerId:   containerName,
		ExecWithSudo:  true,
		ExecInLocal:   true,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.RemoveContainer{
		ContainerId:   containerName,
		ExecWithSudo:  true,
		ExecInLocal:   true,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.UmountFilesystem{
		Directorys:     []string{mountPoint},
		IgnoreNotFound: true,
		IgnoreUmounted: true,
		ExecWithSudo:   true,
		ExecInLocal:    true,
		ExecSudoAlias:  curveadm.SudoAlias(),
	})

	return t, nil
}

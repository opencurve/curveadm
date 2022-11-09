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
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

type (
	step2RemoveContainer struct {
		containerId *string
		plaground   storage.Playground
		curveadm    *cli.CurveAdm
	}

	step2DeletePlayground struct {
		plaground storage.Playground
		curveadm  *cli.CurveAdm
	}
)

func (s *step2RemoveContainer) Execute(ctx *context.Context) error {
	containerId := *s.containerId
	playground := s.plaground
	if len(containerId) == 0 {
		return nil
	}

	steps := []task.Step{}
	steps = append(steps, &step.StopContainer{
		ContainerId: playground.Name,
		ExecOptions: execOptions(s.curveadm),
	})
	steps = append(steps, &step.RemoveContainer{
		ContainerId: playground.Name,
		ExecOptions: execOptions(s.curveadm),
	})
	/*
		mountPoint := playground.MountPoint
		if len(playground.MountPoint) > 0 {
			steps = append(steps, &step.UmountFilesystem{
				Directorys:     []string{mountPoint},
				IgnoreNotFound: true,
				IgnoreUmounted: true,
				ExecOptions:    execOptions(s.curveadm),
			})
		}
	*/

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *step2DeletePlayground) Execute(ctx *context.Context) error {
	err := s.curveadm.Storage().DeletePlayground(s.plaground.Name)
	if err != nil {
		return errno.ERR_DELETE_PLAYGROUND_FAILED.E(err)
	}
	return nil
}

func NewRemovePlaygroundTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	// new task
	playground := v.(storage.Playground)
	subname := fmt.Sprintf("name=%s", playground.Name)
	t := task.NewTask("Remove Playground", subname, nil)

	// add step to task
	var containerId string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}}'",
		Quiet:       true,
		Filter:      fmt.Sprintf("name=%s", playground.Name),
		Out:         &containerId,
		ExecOptions: execOptions(curveadm),
	})
	t.AddStep(&step2RemoveContainer{
		containerId: &containerId,
		plaground:   playground,
		curveadm:    curveadm,
	})
	t.AddStep(&step2DeletePlayground{
		plaground: playground,
		curveadm:  curveadm,
	})

	return t, nil
}

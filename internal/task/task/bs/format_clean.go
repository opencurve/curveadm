/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2023-10-05
 * Author: junfan song (Sonjf-ttk)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

type step2FormatClean struct {
	containerId *string
	fc          *configure.FormatConfig
	curveadm    *cli.CurveAdm
}

func (s *step2FormatClean) Execute(ctx *context.Context) error {
	if len(*s.containerId) == 0 {
		return nil
	}

	var success bool
	steps := []task.Step{}
	steps = append(steps, &step.StopContainer{
		ContainerId: *s.containerId,
		Time:        1,
		ExecOptions: s.curveadm.ExecOptions(),
	})
	steps = append(steps, &step.RemoveContainer{
		Success:     &success,
		ContainerId: *s.containerId,
		ExecOptions: s.curveadm.ExecOptions(),
	})

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewCleanFormatTask(curveadm *cli.CurveAdm, fc *configure.FormatConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(fc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	device := fc.GetDevice()
	mountPoint := fc.GetMountPoint()
	containerName := device2ContainerName(device)
	subname := fmt.Sprintf("host=%s device=%s mountPoint=%s containerName=%s",
		fc.GetHost(), device, mountPoint, containerName)
	t := task.NewTask("Clean Format Container", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2FormatClean{
		containerId: &out,
		fc:          fc,
		curveadm:    curveadm,
	})

	return t, nil
}

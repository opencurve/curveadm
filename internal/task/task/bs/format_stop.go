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
 * Created Date: 2022-11-16
 * Author: guiming liang (demoliang)
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

func skipStopFormat(containerId *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*containerId) < 0 {
			return task.ERR_SKIP_TASK
		}
		return nil
	}
}

type stopContainer struct {
	containerId *string
	curveadm    *cli.CurveAdm
}

func (s *stopContainer) Execute(ctx *context.Context) error {
	if len(*s.containerId) == 0 {
		return nil
	}

	steps := []task.Step{}
	steps = append(steps, &step.StopContainer{
		ContainerId: *s.containerId,
		Time:        1,
		ExecOptions: s.curveadm.ExecOptions(),
	})
	for _, s := range steps {
		err := s.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewStopFormatTask(curveadm *cli.CurveAdm, fc *configure.FormatConfig) (*task.Task, error) {
	host := fc.GetHost()
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	device := fc.GetDevice()
	mountPoint := fc.GetMountPoint()
	containerName := device2ContainerName(device)
	subname := fmt.Sprintf("host=%s device=%s mountPoint=%s containerName=%s",
		fc.GetHost(), device, mountPoint, containerName)
	t := task.NewTask("Stop Format Chunkfile Pool", subname, hc.GetSSHConfig())

	var oldContainerId string
	var oldUuid string

	// 1: list block device and edit fstab delete record
	t.AddStep(&step.BlockId{
		Device:      device,
		Format:      "value",
		MatchTag:    "UUID",
		Out:         &oldUuid,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2EditFSTab{
		host:       host,
		device:     device,
		oldUuid:    &oldUuid,
		mountPoint: mountPoint,
		curveadm:   curveadm,
		skipAdd:    true,
	})

	// 2: list container id and add step to task
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}}'",
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &oldContainerId,
		ExecOptions: curveadm.ExecOptions(),
	})

	t.AddStep(&step.Lambda{
		Lambda: skipStopFormat(&oldContainerId),
	})

	// 3: stop container
	t.AddStep(&stopContainer{
		containerId: &oldContainerId,
		curveadm:    curveadm,
	})

	// 4. umount filesystem
	t.AddStep(&step.UmountFilesystem{
		Directorys:     []string{device},
		IgnoreUmounted: true,
		IgnoreNotFound: true,
		ExecOptions:    curveadm.ExecOptions(),
	})

	return t, nil
}

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

package fs

import (
	"fmt"
	"path"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/fs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

type (
	step2CheckMountPoint struct {
		mountPoint string
	}

	step2CheckMountStatus struct {
		output *string
	}
)

func (s *step2CheckMountPoint) Execute(ctx *context.Context) error {
	if !path.IsAbs(s.mountPoint) {
		return fmt.Errorf("%s: is not an absolute path", s.mountPoint)
	}
	return nil
}

func (s *step2CheckMountStatus) Execute(ctx *context.Context) error {
	if len(*s.output) == 0 { // not mounted
		return task.ERR_SKIP_TASK
	}
	return nil
}

func NewUmountFSTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	mountPoint := curveadm.MemStorage().Get(KEY_MOUNT_POINT).(string)
	subname := fmt.Sprintf("mountPoint=%s", mountPoint)
	t := task.NewTask("Umount FileSystem", subname, nil)

	// add step
	var output string
	containerName := mountPoint2ContainerName(mountPoint)
	t.AddStep(&step2CheckMountPoint{
		mountPoint: mountPoint,
	})
	t.AddStep(&step.UmountFilesystem{
		Directorys:     []string{mountPoint},
		IgnoreUmounted: true,
		ExecWithSudo:   false,
		ExecInLocal:    true,
	})
	t.AddStep(&step.ListContainers{
		ShowAll:      true,
		Format:       "'{{.Status}}'",
		Quiet:        true,
		Filter:       fmt.Sprintf("name=%s", containerName),
		Out:          &output,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	t.AddStep(&step2CheckMountStatus{
		output: &output,
	})
	t.AddStep(&step.WaitContainer{
		ContainerId:  containerName,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	t.AddStep(&step.RemoveContainer{
		ContainerId:  containerName,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	return t, nil
}

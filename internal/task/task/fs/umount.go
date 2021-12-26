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
	"strings"

	"github.com/opencurve/curveadm/internal/configure/client"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	DEFAULT_WAIT_STOP_SECONDS = 24 * 3600 // 1 day
	ERR_NOT_MOUNTED           = "not mounted"
)

type step2UmountFilesystem struct {
	mountPoint string
	curveadm   *cli.CurveAdm
}

func (s *step2UmountFilesystem) Execute(ctx *context.Context) error {
	var out string
	step := step.UmountFilesystem{
		Directory:    s.mountPoint,
		Out:          &out,
		ExecWithSudo: false,
		ExecInLocal:  true,
	}
	err := step.Execute(ctx)
	if err != nil && !strings.Contains(out, ERR_NOT_MOUNTED) {
		return err
	}
	return nil
}

func NewUmountFSTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	mountPoint := curveadm.MemStorage().Get(KEY_MOUNT_POINT).(string)
	subname := fmt.Sprintf("mountPoint=%s", mountPoint)
	t := task.NewTask("Umount FileSystem", subname, nil)

	// add step
	containerId := mountPoint2ContainerName(mountPoint)
	t.AddStep(&step2UmountFilesystem{
		mountPoint: mountPoint,
		curveadm:   curveadm,
	})
	t.AddStep(&step.StopContainer{
		ContainerId:  containerId,
		Time:         DEFAULT_WAIT_STOP_SECONDS,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	t.AddStep(&step.RemoveContainer{
		ContainerId:  containerId,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	return t, nil
}

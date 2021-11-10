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

package client

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tasks/task"
)

type (
	step2UmountFS struct{ mountPoint string }
)

func (s *step2UmountFS) Execute(ctx *task.Context) error {
	status, err := getMountStatus(ctx, s.mountPoint)
	if err != nil {
		return err
	} else if status.Status == STATUS_UNMOUNTED {
		return nil
	} else if status.Status == STATUS_NORMAL { // stop container
		_, err = ctx.Module().LocalShell("sudo docker stop %s", status.ContainerId)
		if err != nil {
			return err
		}
	}

	// remove container
	_, err = ctx.Module().LocalShell("sudo docker rm %s", status.ContainerId)
	return err
}

func (s *step2UmountFS) Rollback(ctx *task.Context) {
}

func NewUmountFSTask(curvradm *cli.CurveAdm, mountPoint string) (*task.Task, error) {
	t := task.NewTask("Umount FileSystem", "", nil)
	t.AddStep(&step2UmountFS{mountPoint: mountPoint})
	return t, nil
}

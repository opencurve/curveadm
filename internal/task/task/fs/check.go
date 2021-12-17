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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/client"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	STATUS_UNMOUNTED = "unmount"
	STATUS_NORMAL    = "normal"
	STATUS_ABNORMAL  = "abnormal"

	KEY_MOUNT_STATUS = "MOUNT_STATUS"
)

type (
	step2FormatMountStatus struct {
		output        *string
		mountPoint    string
		containerName string
		memStorage    *utils.SafeMap
	}

	MountStatus struct {
		MountPoint    string
		ContainerId   string
		ContainerName string
		Status        string
	}
)

func (s *step2FormatMountStatus) Execute(ctx *context.Context) error {
	status, containerId := func(output string) (status, containerId string) {
		if len(output) == 0 {
			status = STATUS_UNMOUNTED
			containerId = "-"
		} else {
			items := strings.Split(output, " ")
			containerId = items[0]
			if strings.HasPrefix(items[1], "Up") {
				status = STATUS_NORMAL
			} else {
				status = STATUS_ABNORMAL
			}
		}
		return
	}(*s.output)

	s.memStorage.Set(KEY_MOUNT_STATUS, MountStatus{
		MountPoint:    s.mountPoint,
		ContainerId:   tui.TrimContainerId(containerId),
		ContainerName: s.containerName,
		Status:        status,
	})
	return nil
}

func mountPoint2ContainerName(mountPoint string) string {
	return utils.MD5Sum(mountPoint)
}

func NewGetMountStatusTask(curvradm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	mountPoint := curvradm.MemStorage().Get(KEY_MOUNT_POINT).(string)
	subname := fmt.Sprintf("mountPoint=%s", mountPoint)
	t := task.NewTask("Check Mount Point", subname, nil)

	var output string
	containerName := mountPoint2ContainerName(mountPoint)
	t.AddStep(&step.ListContainers{
		ShowAll:      true,
		Format:       "'{{.ID}} {{.Status}}'",
		Quiet:        true,
		Filter:       fmt.Sprintf("name=%s", containerName),
		Out:          &output,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	t.AddStep(&step2FormatMountStatus{
		output:     &output,
		mountPoint: mountPoint,
		memStorage: curvradm.MemStorage(),
	})

	return t, nil
}

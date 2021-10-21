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
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	STATUS_UNMOUNTED = "unmount"
	STATUS_NORMAL    = "normal"
	STATUS_ABNORMAL  = "abnormal"

	KEY_MOUNT_STATUS = "MOUNT_STATUS"
)

var (
	GET_CONTAINER_STATUS_ARGS = []string{
		"--quiet",
		"--all",
		"--format '{{.Status}} {{.ID}}'",
		"--filter name=%s",
	}
)

type (
	step2CheckMount struct {
		mountPoint string
		memStorage *utils.SafeMap
	}
)

type MountStatus struct {
	MountPoint    string
	ContainerId   string
	ContainerName string
	Status        string
}

func extractContainerId(out string) string {
	if out == "" {
		return "-"
	}
	items := strings.Split(out, " ")
	return tui.TrimContainerId(items[len(items)-1])
}

func getMountStatus(ctx *task.Context, mountPoint string) (*MountStatus, error) {
	containerName := utils.MD5Sum(mountPoint)
	format := strings.Join(GET_CONTAINER_STATUS_ARGS, " ")
	args := fmt.Sprintf(format, containerName)
	out, err := ctx.Module().LocalShell("sudo docker ps %s", args)

	status := STATUS_ABNORMAL
	if err != nil {
		return nil, err
	} else if len(out) == 0 {
		status = STATUS_UNMOUNTED
	} else if strings.HasPrefix(out, "Up") {
		status = STATUS_NORMAL
	}

	return &MountStatus{
		MountPoint:    mountPoint,
		ContainerId:   extractContainerId(out),
		ContainerName: containerName,
		Status:        status,
	}, nil
}

func (s *step2CheckMount) Execute(ctx *task.Context) error {
	status, err := getMountStatus(ctx, s.mountPoint)
	if err != nil {
		return err
	}
	s.memStorage.Set(KEY_MOUNT_STATUS, *status)
	return nil
}

func (s *step2CheckMount) Rollback(ctx *task.Context) {
}

func NewGetMountStatusTask(curvradm *cli.CurveAdm, mountPoint string) (*task.Task, error) {
	t := task.NewTask("Check Mount Point", "", nil)
	t.AddStep(&step2CheckMount{mountPoint: mountPoint, memStorage: curvradm.MemStorage()})
	return t, nil
}

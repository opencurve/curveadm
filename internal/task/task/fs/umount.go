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
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	SIGNATURE_NOT_MOUNTED = "not mounted"
)

type (
	step2UmountFS struct {
		containerId string
		status      *string
		mountPoint  string
		curveadm    *cli.CurveAdm
	}

	step2RemoveContainer struct {
		status      *string
		containerId string
		curveadm    *cli.CurveAdm
	}

	step2DeleteClient struct {
		fsId     string
		curveadm *cli.CurveAdm
	}
)

func checkContainerId(containerId string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(containerId) == 0 {
			return task.ERR_SKIP_TASK
		}
		return nil
	}
}

func (s *step2UmountFS) Execute(ctx *context.Context) error {
	if len(*s.status) == 0 {
		return nil
	} else if !strings.HasPrefix(*s.status, "Up") {
		return nil
	}

	command := fmt.Sprintf("umount %s", configure.GetFSClientMountPath(s.mountPoint))
	dockerCli := ctx.Module().DockerCli().ContainerExec(s.containerId, command)
	out, err := dockerCli.Execute(s.curveadm.ExecOptions())
	if strings.Contains(out, SIGNATURE_NOT_MOUNTED) {
		return nil
	} else if err == nil {
		return nil
	}
	return errno.ERR_UMOUNT_FILESYSTEM_FAILED.S(out)
}

func (s *step2DeleteClient) Execute(ctx *context.Context) error {
	err := s.curveadm.Storage().DeleteClient(s.fsId)
	if err != nil {
		return errno.ERR_DELETE_CLIENT_FAILED.E(err)
	}
	return nil
}

func (s *step2RemoveContainer) Execute(ctx *context.Context) error {
	if len(*s.status) == 0 {
		return nil
	}

	steps := []task.Step{}
	if strings.HasPrefix(*s.status, "Up") {
		steps = append(steps, &step.WaitContainer{
			ContainerId: s.containerId,
			ExecOptions: s.curveadm.ExecOptions(),
		})
	}
	steps = append(steps, &step.RemoveContainer{
		ContainerId: s.containerId,
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

func NewUmountFSTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_MOUNT_OPTIONS).(MountOptions)
	fsId := curveadm.GetFilesystemId(options.Host, options.MountPoint)
	containerId, err := curveadm.Storage().GetClientContainerId(fsId)
	if err != nil {
		return nil, errno.ERR_GET_CLIENT_CONTAINER_ID_FAILED.E(err)
	}
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	// new task
	mountPoint := options.MountPoint
	subname := fmt.Sprintf("host=%s mountPoint=%s", options.Host, mountPoint)
	t := task.NewTask("Umount FileSystem", subname, hc.GetSSHConfig())

	// add step to task
	var status string
	t.AddStep(&step.Lambda{
		Lambda: checkContainerId(containerId),
	})
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Quiet:       true,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2UmountFS{
		containerId: containerId,
		status:      &status,
		mountPoint:  options.MountPoint,
		curveadm:    curveadm,
	})
	t.AddStep(&step2RemoveContainer{
		status:      &status,
		containerId: containerId,
		curveadm:    curveadm,
	})
	t.AddStep(&step2DeleteClient{
		curveadm: curveadm,
		fsId:     fsId,
	})

	return t, nil
}

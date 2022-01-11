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
 * Created Date: 2021-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/pkg/module"
)

type (
	step2UnmapImage struct {
		output *string
		user   string
		volume string
	}
)

func (s *step2UnmapImage) Execute(ctx *context.Context) error {
	output := *s.output
	if len(output) == 0 {
		return nil
	}

	items := strings.Split(output, " ")
	containerId := items[0]
	status := items[1]
	if !strings.HasPrefix(status, "Up") {
		return nil
	}

	command := fmt.Sprintf("curve-nbd unmap cbd:pool/%s_%s_", s.volume, s.user)
	dockerCli := ctx.Module().DockerCli().ContainerExec(containerId, command)
	_, err := dockerCli.Execute(module.ExecOption{
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	return err
}

func NewUnmapTask(curvradm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	option := curvradm.MemStorage().Get(KEY_MAP_OPTION).(MapOption)
	user, volume := option.User, option.Volume
	subname := fmt.Sprintf("hostname=%s volume=%s", cc.GetHost(), volume)
	t := task.NewTask("Unmap Image", subname, cc.GetSSHConfig())

	// add step
	var output string
	containerName := volume2ContainerName(user, volume)
	t.AddStep(&step.ListContainers{
		ShowAll:      true,
		Format:       "'{{.ID}} {{.Status}}'",
		Quiet:        true,
		Filter:       fmt.Sprintf("name=%s", containerName),
		Out:          &output,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step2UnmapImage{
		output: &output,
		user:   user,
		volume: volume,
	})
	t.AddStep(&step.StopContainer{
		ContainerId:  containerName,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step.RemoveContainer{
		ContainerId:  containerName,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})

	return t, nil
}

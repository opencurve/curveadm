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
 * Created Date: 2022-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	SIGNATURE_NOT_MAPPED = "is not mapped"
)

type (
	step2UnmapImage struct {
		output      *string
		user        string
		volume      string
		execOptions module.ExecOptions
	}

	step2RemoveContainer struct {
		curveadm    *cli.CurveAdm
		status      *string
		containerId string
	}

	step2DeleteClient struct {
		curveadm *cli.CurveAdm
		volumeId string
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
	out, err := dockerCli.Execute(s.execOptions)
	if err == nil {
		return nil
	} else if strings.Contains(out, SIGNATURE_NOT_MAPPED) {
		return nil
	}
	return errno.ERR_UNMAP_VOLUME_FAILED.S(out)
}

func (s *step2RemoveContainer) Execute(ctx *context.Context) error {
	if len(*s.status) == 0 {
		return nil
	}

	steps := []task.Step{}
	steps = append(steps, &step.StopContainer{
		ContainerId: s.containerId,
		ExecOptions: s.curveadm.ExecOptions(),
	})
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

func (s *step2DeleteClient) Execute(ctx *context.Context) error {
	err := s.curveadm.Storage().DeleteClient(s.volumeId)
	if err != nil {
		return errno.ERR_DELETE_CLIENT_FAILED.E(err)
	}

	err = s.curveadm.Storage().DeleteClientConfig(s.volumeId)
	if err != nil {
		return errno.ERR_DELETE_CLIENT_CONFIG_FAILED.E(err)
	}

	return nil
}

func NewUnmapTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_MAP_OPTIONS).(MapOptions)
	volumeId := curveadm.GetVolumeId(options.Host, options.User, options.Volume)
	containerId, err := curveadm.Storage().GetClientContainerId(volumeId)
	if err != nil {
		return nil, errno.ERR_GET_CLIENT_CONTAINER_ID_FAILED.E(err)
	}
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("hostname=%s volume=%s:%s containerId=%s",
		hc.GetHostname(), options.User, options.Volume, tui.TrimContainerId(containerId))
	t := task.NewTask("Unmap Volume", subname, hc.GetConnectConfig())

	// add step
	var output string
	t.AddStep(&step.Lambda{
		Lambda: checkContainerId(containerId),
	})
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}} {{.Status}}'",
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2UnmapImage{
		output:      &output,
		user:        options.User,
		volume:      options.Volume,
		execOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2RemoveContainer{
		curveadm:    curveadm,
		status:      &output,
		containerId: containerId,
	})
	t.AddStep(&step2DeleteClient{
		curveadm: curveadm,
		volumeId: volumeId,
	})

	return t, nil
}

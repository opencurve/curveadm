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

// __SIGN_BY_WINE93__

package common

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	CMD_ADD_CONTABLE = "bash -c '[[ ! -z $(which crontab) ]] && crontab %s'"
)

type Step2CheckPostStart struct {
	Host        string
	Role        string
	ContainerId string
	Success     *bool
	Out         *string
	ExecOptions module.ExecOptions
}

func (s *Step2CheckPostStart) Execute(ctx *context.Context) error {
	if *s.Success {
		return nil
	}

	var status string
	step := &step.InspectContainer{
		ContainerId: s.ContainerId,
		Format:      "'{{.State.Status}}'",
		Out:         &status,
		ExecOptions: s.ExecOptions,
	}
	err := step.Execute(ctx)
	if err != nil {
		return errno.ERR_START_CRONTAB_IN_CONTAINER_FAILED.S(*s.Out)
	} else if status != "running" {
		return errno.ERR_CONTAINER_IS_ABNORMAL.
			F("host=%s role=%s containerId=%s",
				s.Host, s.Role, tui.TrimContainerId(s.ContainerId))
	}
	return nil
}

func NewStartServiceTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if curveadm.IsSkip(dc) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Start Service", subname, hc.GetSSHConfig(), hc.GetHttpConfig())

	// add step to task
	var out string
	var success bool
	host, role := dc.GetHost(), dc.GetRole()
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: CheckContainerExist(host, role, containerId, &out),
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: WaitContainerStart(3),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     fmt.Sprintf(CMD_ADD_CONTABLE, CURVE_CRONTAB_FILE),
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&Step2CheckPostStart{
		Host:        dc.GetHost(),
		Role:        dc.GetRole(),
		ContainerId: containerId,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

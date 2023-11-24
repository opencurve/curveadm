/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-04-24
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func IsSkip(mc *configure.MonitorConfig, roles []string) bool {
	role := mc.GetRole()
	for _, r := range roles {
		if role == r {
			return true
		}
	}
	return false
}

func NewStartServiceTask(curveadm *cli.CurveAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(cfg.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if IsSkip(cfg, []string{ROLE_MONITOR_CONF}) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := curveadm.GetHost(cfg.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Start Service", subname, hc.GetSSHConfig(), hc.GetHttpConfig())

	// add step to task
	var out string
	var success bool
	role, host := cfg.GetRole(), cfg.GetHost()
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.CheckContainerExist(host, role, containerId, &out),
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.WaitContainerStart(3),
	})
	t.AddStep(&common.Step2CheckPostStart{
		Host:        cfg.GetHost(),
		Role:        cfg.GetRole(),
		ContainerId: containerId,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

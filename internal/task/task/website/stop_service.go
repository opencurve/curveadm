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
* Created Date: 2023-05-08
* Author: wanghai (SeanHai)
 */

package website

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func NewStopServiceTask(curveadm *cli.CurveAdm, cfg *configure.WebsiteConfig) (*task.Task, error) {
	serviceId := curveadm.GetWebsiteServiceId(cfg.GetId())
	containerId, _ := curveadm.GetContainerId(serviceId)

	hc, err := curveadm.GetHost(cfg.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Stop Website", subname, hc.GetSSHConfig())

	// add step to task
	var out string
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
	t.AddStep(&step.StopContainer{
		ContainerId: containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	return t, nil
}

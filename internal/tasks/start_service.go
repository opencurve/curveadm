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

package tasks

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

type (
	step2RunContainer struct{ containerId string }
	step2PostStart    struct{ containerId string }
)

func (s *step2RunContainer) Execute(ctx *task.Context) error {
	_, err := ctx.Module().SshShell("sudo docker start %s", s.containerId)
	return err
}

func (s *step2RunContainer) Rollback(ctx *task.Context) {
}

func (s *step2PostStart) Execute(ctx *task.Context) error {
	cmd := fmt.Sprintf("[[ ! -z $(which crontab) ]] && crontab /var/spool/cron/crontabs/root")
	_, err := ctx.Module().SshShell("sudo docker exec %s /bin/bash -c '%s'", s.containerId, cmd)
	return err
}

func (s *step2PostStart) Rollback(ctx *task.Context) {
}

func NewStartServiceTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Start Service", subname, dc)
	t.AddStep(&step2RunContainer{containerId: containerId})
	t.AddStep(&step2PostStart{containerId: containerId})
	return t, nil
}

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
 * Project: CurveAdm
 * Created Date: 2023-11-30
 * Author: Xianfei Cao (caoxianfei1)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func CheckContainerExist(host, role, containerId string, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*out) == 0 {
			return errno.ERR_CONTAINER_ALREADT_REMOVED.
				F("host=%s role=%s containerId=%s",
					host, role, tui.TrimContainerId(containerId))
		}
		return nil
	}
}

func checkMarkStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_MARK_CHUNKSERVER_PENDDING.S(*out)
		}
		return nil
	}
}

func NewMarkServerPendding(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Mark Chunkserver Pendding", subname, hc.GetSSHConfig())

	var out string
	var success bool
	host, role := dc.GetHost(), dc.GetRole()
	layout := dc.GetProjectLayout()
	markCSPenddingScript := scripts.MARK_SERVER_PENDDING
	scriptPath := layout.ToolsBinDir + "/mark_server_pendding.sh"

	migrates := []*configure.MigrateServer{}
	if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil {
		migrates = curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
	}

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
	t.AddStep(&step.InstallFile{ // install /curvebs/tools/sbin/mark_chunkserver_pendding.sh
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		Content:           &markCSPenddingScript,
		ExecOptions:       curveadm.ExecOptions(),
	})
	for _, migrate := range migrates {
		hostip := migrate.From.GetListenIp()
		hostport := migrate.From.GetListenPort()
		t.AddStep(&step.ContainerExec{
			ContainerId: &containerId,
			Command:     fmt.Sprintf("/bin/bash %s %s %d", scriptPath, hostip, hostport),
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step.Lambda{
			Lambda: checkMarkStatus(&success, &out),
		})
	}

	return t, nil
}

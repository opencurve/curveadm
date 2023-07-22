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
 * Created Date: 2023-07-07
 * Author: caoxianfei1
 */

package common

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	ROLE_USER_SERVICE = "service"
	ROLE_USER_CLIENT  = "client"
)

func checkDistSericeKeySuccess(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_DIST_SERVICE_KEY_FAILED
		}
		return nil
	}
}

func NewDiskAuthKeyTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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

	var success bool
	var out string
	host, role := dc.GetHost(), dc.GetRole()
	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Distribute Auth Key", subname, hc.GetSSHConfig())

	script := scripts.SCRIPT_DIST_AUTH_KEY
	layout := dc.GetProjectLayout()
	scriptPath := fmt.Sprintf("%s/dist_auth_key.sh", layout.ToolsBinDir)

	authServerKey := curveadm.MemStorage().Get(comm.AUTH_SERVER_KEY).(string)
	rolesAuthInfo := curveadm.MemStorage().Get(comm.ROLES_AUTH_INFO).(map[string]comm.RoleAuthInfo)

	waitScript := scripts.SCRIPT_WAIT
	waitScriptPath := fmt.Sprintf("%s/wait.sh", layout.ToolsBinDir)

	// cluster MDS address
	clusterMDSAddrs, err := dc.GetVariables().Get("cluster_mds_addr")
	clusterMDSAddrs = strings.Replace(clusterMDSAddrs, ",", " ", -1)
	if err != nil {
		return nil, err
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
	t.AddStep(&step.InstallFile{ // install /curvebs/tools/sbin/dist_auth_key.sh
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		Content:           &script,
		ExecOptions:       curveadm.ExecOptions(),
	})

	t.AddStep(&step.InstallFile{ // install wait script
		ContainerId:       &containerId,
		ContainerDestPath: waitScriptPath,
		Content:           &waitScript,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{ // wait mds leader election success
		ContainerId: &containerId,
		Command:     fmt.Sprintf("bash %s %s", waitScriptPath, clusterMDSAddrs),
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkWaitMDSElectionSuccess(&success, &out),
	})

	for role, authInfo := range rolesAuthInfo { // register mds, chunkserver, snapshotserver key and relevant tools key
		command := fmt.Sprintf("/bin/bash %s %s %s %s %s %s %s %s",
			scriptPath,
			role,
			ROLE_USER_SERVICE,
			authInfo.AuthKeyCurrent,
			authServerKey,
			authInfo.AuthClientId,
			ROLE_USER_CLIENT,
			authInfo.AuthClientKey,
		)

		t.AddStep(&step.ContainerExec{
			ContainerId: &containerId,
			Success:     &success,
			Out:         &out,
			Command:     command,
			ExecOptions: curveadm.ExecOptions(),
		})

		t.AddStep(&step.Lambda{
			Lambda: checkDistSericeKeySuccess(&success, &out),
		})
	}

	return t, nil
}

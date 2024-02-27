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
* Created Date: 2023-08-02
* Author: wanghai (SeanHai)
 */

package common

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func checkEnableEtcdAuthStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_ENABLE_ETCD_AUTH_FAILED.S(*out)
		}
		return nil
	}
}

func NewEnableEtcdAuthTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Enable Etcd Auth", subname, hc.GetConnectConfig())

	script := scripts.ENABLE_ETCD_AUTH
	layout := dc.GetProjectLayout()
	scriptPath := fmt.Sprintf("%s/enable_auth.sh", layout.ServiceBinDir)

	etcdEndPoints, err := dc.GetVariables().Get("cluster_etcd_addr")
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
	t.AddStep(&step.InstallFile{ // install /curvebs(fs)/etcd/sbin/enable_auth.sh
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		Content:           &script,
		ExecOptions:       curveadm.ExecOptions(),
	})
	command := fmt.Sprintf("/bin/bash %s %s %s %s", scriptPath, etcdEndPoints, dc.GetEtcdAuthUsername(),
		dc.GetEtcdAuthPassword())
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Success:     &success,
		Out:         &out,
		Command:     command,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkEnableEtcdAuthStatus(&success, &out),
	})
	return t, nil
}

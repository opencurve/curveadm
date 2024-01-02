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
 * Created Date: 2023-12-20
 * Author: Caoxianfei
 */

package common

import (
	"fmt"
	"strconv"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func checkRemoveEtcdMemberStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_REMOVE_ETCD_MEMBER.S(*out)
		}
		if *out == "NOTEXIST" {
			return task.ERR_SKIP_TASK
		}
		return nil
	}
}

func NewRemoveEtcdMemberTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Remove Old Etcd Member", subname, hc.GetSSHConfig())

	host, role := dc.GetHost(), dc.GetRole()
	script := scripts.REMOVE_ETCD
	layout := dc.GetProjectLayout()
	scriptPath := fmt.Sprintf("%s/remove_etcd.sh", layout.ServiceBinDir)
	etcdctlPath := layout.ServiceBinDir + "/etcdctl"
	endpoints, err := dc.GetVariables().Get("cluster_etcd_addr")
	if err != nil {
		return nil, errno.ERR_GET_CLUSTER_ETCD_ADDR
	}
	oldName := fmt.Sprint("etcd", strconv.Itoa(dc.GetHostSequence()), strconv.Itoa(dc.GetInstancesSequence()))
	removeEtcdCmd := fmt.Sprintf("/bin/bash %s %s %s %s", scriptPath, etcdctlPath, endpoints, oldName)

	var success bool
	var out string
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
	t.AddStep(&step.InstallFile{
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		Content:           &script,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Success:     &success,
		Out:         &out,
		Command:     removeEtcdCmd,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkRemoveEtcdMemberStatus(&success, &out),
	})

	return t, nil
}

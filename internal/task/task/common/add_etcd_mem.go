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

func checkAddEtcdMemberStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_ADD_ETCD_MEMEBER.S(*out)
		}
		if *out == "EXIST" {
			return task.ERR_SKIP_TASK
		}
		return nil
	}
}

func NewAddEtcdMemberTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Add Etcd Member", subname, hc.GetSSHConfig())

	host, role := dc.GetHost(), dc.GetRole()
	script := scripts.ADD_ETCD
	layout := dc.GetProjectLayout()
	scriptPath := fmt.Sprintf("%s/add_etcd.sh", layout.ServiceBinDir)
	etcdctlPath := layout.ServiceBinDir + "/etcdctl"
	endpoints, err := dc.GetVariables().Get("cluster_etcd_addr")
	if err != nil {
		return nil, errno.ERR_GET_CLUSTER_ETCD_ADDR
	}
	oldName := fmt.Sprint("etcd", strconv.Itoa(dc.GetHostSequence()), strconv.Itoa(dc.GetInstancesSequence()))
	newName := fmt.Sprint("etcd", strconv.Itoa(dc.GetHostSequence()+3), strconv.Itoa(dc.GetInstancesSequence()))
	migrates := []*configure.MigrateServer{}
	if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil {
		migrates = curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
	}
	toService := migrates[0].To
	peerUrl := fmt.Sprint("http://", toService.GetListenIp(), ":", strconv.Itoa(toService.GetListenPort()))
	addEtcdCmd := fmt.Sprintf("/bin/bash %s %s %s %s %s %s", scriptPath, etcdctlPath, endpoints, oldName, newName, peerUrl)

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
		Command:     addEtcdCmd,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkAddEtcdMemberStatus(&success, &out),
	})

	return t, nil
}

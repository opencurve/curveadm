/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-05-21
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func genBackupCommand(dc *topology.DeployConfig) string {
	layout := dc.GetProjectLayout()
	binaryPath := fmt.Sprintf("%s/etcdctl", layout.ServiceBinDir)
	endpoint := fmt.Sprintf("%s:%d", dc.GetListenIp(), dc.GetListenPort())
	savePath := fmt.Sprintf("%s/snapshot.%s.db", layout.ServiceDataDir, time.Now().Format("2006-01-02-15:04:05"))
	command := fmt.Sprintf("%s --endpoints %s snapshot save %s", binaryPath, endpoint, savePath)
	return command
}

func NewBackupEtcdDataTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Backup Etcd Data", subname, hc.GetConnectConfig())

	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     genBackupCommand(dc),
		ExecOptions: curveadm.ExecOptions(),
	})
	return t, nil
}

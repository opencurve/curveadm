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
)

func getMountVolumes(cfg *configure.WebsiteConfig) []step.Volume {
	volumes := []step.Volume{}
	volumes = append(volumes, step.Volume{
		HostPath:      cfg.GetDataDir(),
		ContainerPath: "/curve-manager/db",
	},
		step.Volume{
			HostPath:      cfg.GetLogDir(),
			ContainerPath: "/curve-manager/logs",
		})
	return volumes
}

func NewCreateContainerTask(curveadm *cli.CurveAdm, cfg *configure.WebsiteConfig) (*task.Task, error) {
	host := cfg.GetHost()
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s", host, cfg.GetRole())
	t := task.NewTask("Create Container", subname, hc.GetSSHConfig())

	// add step to task
	var oldContainerId, containerId string
	clusterId := curveadm.ClusterId()
	wcId := cfg.GetId()
	serviceId := curveadm.GetWebsiteServiceId(wcId)
	kind := cfg.GetKind()
	role := cfg.GetRole()
	hostname := fmt.Sprintf("%s-%s-%s", kind, role, serviceId)
	options := curveadm.ExecOptions()
	options.ExecWithSudo = false

	t.AddStep(&common.Step2GetService{ // if service exist, break task
		ServiceId:   serviceId,
		ContainerId: &oldContainerId,
		Storage:     curveadm.Storage(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{cfg.GetDataDir(), cfg.GetLogDir()},
		ExecOptions: options,
	})
	t.AddStep(&step.CreateContainer{
		Image:       cfg.GetImage(),
		AddHost:     []string{fmt.Sprintf("%s:127.0.0.1", hostname)},
		Hostname:    hostname,
		Init:        true,
		Name:        hostname,
		Privileged:  true,
		User:        "0:0",
		Pid:         "host",
		Network:     "bridge",
		Publish:     fmt.Sprintf("%d:443", cfg.GetListenPort()),
		Restart:     common.POLICY_NEVER_RESTART,
		Ulimits:     []string{"core=-1"},
		Volumes:     getMountVolumes(cfg),
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.TrimContainerId(&containerId),
	})
	t.AddStep(&common.Step2InsertService{
		ClusterId:      clusterId,
		ServiceId:      serviceId,
		ContainerId:    &containerId,
		OldContainerId: &oldContainerId,
		Storage:        curveadm.Storage(),
	})
	return t, nil
}

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
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	ETCD_ENDPOINT = "etcd.endpoint"
)

func mutateMDSConf(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}
		if key == ETCD_ENDPOINT {
			value = options[ETCD_ENDPOINT].(string)
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func NewAmendMdsConfigTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Override Mds configure", subname, hc.GetSSHConfig())

	layout := dc.GetProjectLayout()
	endpoints, err := dc.GetVariables().Get("cluster_etcd_addr")
	if err != nil {
		return nil, err
	}
	migrates := []*configure.MigrateServer{}
	if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil {
		migrates = curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
	}
	toService := migrates[0].To
	toSeriveEndpint := fmt.Sprint(toService.GetListenIp(), ":", strconv.Itoa(toService.GetListenClientPort()))
	endpoints = fmt.Sprint(endpoints, ",", toSeriveEndpint)
	options[ETCD_ENDPOINT] = endpoints

	t.AddStep(&step.SyncFile{ // sync mds.conf config again
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  layout.ServiceConfPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: layout.ServiceConfPath,
		KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
		Mutate:            mutateMDSConf(dc, DEFAULT_CONFIG_DELIMITER, false),
		ExecOptions:       curveadm.ExecOptions(),
	})

	return t, nil
}

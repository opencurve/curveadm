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
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	AMEND_NAME      = "name"
	AMEND_ENDPOINTS = "initial-cluster"
	AMEND_STATE     = "initial-cluster-state"
)

var options = make(map[string]interface{})

func mutateEtcdConf(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}
		if key == AMEND_NAME {
			value = options[AMEND_NAME].(string)
		} else if key == AMEND_ENDPOINTS {
			value = options[AMEND_ENDPOINTS].(string)
		} else if key == AMEND_STATE {
			value = "existing"
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func NewAmendEtcdConfigTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Override Etcd configure", subname, hc.GetSSHConfig())

	layout := dc.GetProjectLayout()
	migrates := []*configure.MigrateServer{}
	if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil {
		migrates = curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
	}
	dcs := []*topology.DeployConfig{}
	if curveadm.MemStorage().Get(comm.KEY_CLUSTER_DCS) != nil {
		dcs = curveadm.MemStorage().Get(comm.KEY_CLUSTER_DCS).([]*topology.DeployConfig)
	}
	endpoints := []string{}
	for _, dc := range dcs {
		ept := fmt.Sprint("etcd", strconv.Itoa(dc.GetHostSequence()), strconv.Itoa(dc.GetInstancesSequence()),
			"=", "http://", dc.GetListenIp(), ":", strconv.Itoa(dc.GetListenPort()))
		endpoints = append(endpoints, ept)
	}
	toService := migrates[0].To
	newName := fmt.Sprint("etcd", strconv.Itoa(toService.GetHostSequence()+3), strconv.Itoa(toService.GetInstancesSequence()))
	toSeriveEndpint := fmt.Sprint(newName, "=", "http://",
		toService.GetListenIp(), ":", strconv.Itoa(toService.GetListenPort()))
	endpoints = append(endpoints, toSeriveEndpint)
	endpointsStr := strings.Join(endpoints, ",")

	options[AMEND_NAME] = newName
	options[AMEND_ENDPOINTS] = endpointsStr

	t.AddStep(&step.SyncFile{ // sync etcd.conf config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  layout.ServiceConfPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: layout.ServiceConfPath,
		KVFieldSplit:      ETCD_CONFIG_DELIMITER,
		Mutate:            mutateEtcdConf(dc, ETCD_CONFIG_DELIMITER, false),
		ExecOptions:       curveadm.ExecOptions(),
	})

	return t, nil
}

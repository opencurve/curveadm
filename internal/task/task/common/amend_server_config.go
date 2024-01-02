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
	ETCD_ENDPOINT     = "etcd.endpoint"     // fs
	MDS_ETCD_ENDPOINT = "mds.etcd.endpoint" // bs
	MDS_LISTEN_ADDR   = "mds.listen.addr"   // fs and bs
)

func mutateServerConf(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}
		switch key {
		case ETCD_ENDPOINT:
			value = options[ETCD_ENDPOINT].(string)
		case MDS_LISTEN_ADDR:
			if dc.GetRole() != topology.ROLE_MDS { // bs mds.conf has config 'mds.listen.addr'
				value = options[MDS_LISTEN_ADDR].(string)
			}
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func NewAmendServerConfigTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	t := task.NewTask("Override Server configure", subname, hc.GetSSHConfig())

	layout := dc.GetProjectLayout()
	role := dc.GetRole()
	var endpoints string
	if role == topology.ROLE_MDS {
		endpoints, err = dc.GetVariables().Get("cluster_etcd_addr")
		if err != nil {
			return nil, err
		}
	} else if role == topology.ROLE_METASERVER ||
		role == topology.ROLE_CHUNKSERVER ||
		role == topology.ROLE_SNAPSHOTCLONE {
		endpoints, err = dc.GetVariables().Get("cluster_mds_addr")
		if err != nil {
			return nil, err
		}
	}

	migrates := []*configure.MigrateServer{}
	if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil {
		migrates = curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
	}
	fromService := migrates[0].From
	toService := migrates[0].To
	var fromListenOrClientPort, toListenOrClientPort string
	// override etcd addr in mds.config
	if role == topology.ROLE_MDS {
		fromListenOrClientPort = strconv.Itoa(fromService.GetListenClientPort())
		toListenOrClientPort = strconv.Itoa(fromService.GetListenClientPort())
		// orveride mds addr in metaserver.conf (FS) OR
		// override mds add in chunkserver.conf and snap_client.conf (BS)
	} else if role == topology.ROLE_METASERVER ||
		role == topology.ROLE_CHUNKSERVER ||
		role == topology.ROLE_SNAPSHOTCLONE {
		fromListenOrClientPort = strconv.Itoa(fromService.GetListenPort())
		toListenOrClientPort = strconv.Itoa(toService.GetListenPort())
	}
	fromServiceEndpoint := fmt.Sprint(fromService.GetListenIp(), ":", fromListenOrClientPort)
	toSeriveEndpint := fmt.Sprint(toService.GetListenIp(), ":", toListenOrClientPort)
	epSlice := strings.Split(endpoints, ",")
	removedFromService := []string{}
	for _, ep := range epSlice {
		if ep != fromServiceEndpoint {
			removedFromService = append(removedFromService, ep)
		}
	}
	removedFromService = append(removedFromService, toSeriveEndpint)
	endpoints = strings.Join(removedFromService, ",")
	if role == topology.ROLE_MDS {
		if dc.GetKind() == topology.KIND_CURVEFS {
			options[ETCD_ENDPOINT] = endpoints
		} else {
			options[MDS_ETCD_ENDPOINT] = endpoints
		}
	} else if role == topology.ROLE_METASERVER ||
		role == topology.ROLE_CHUNKSERVER ||
		role == topology.ROLE_SNAPSHOTCLONE {
		options[MDS_LISTEN_ADDR] = endpoints
	}

	configPath := layout.ServiceConfPath
	if role == topology.ROLE_SNAPSHOTCLONE {
		configPath = layout.ServiceConfDir + "/snap_client.conf"
	}

	t.AddStep(&step.SyncFile{ // sync mds.conf config again
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  configPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: configPath,
		KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
		Mutate:            mutateServerConf(dc, DEFAULT_CONFIG_DELIMITER, false),
		ExecOptions:       curveadm.ExecOptions(),
	})

	return t, nil
}

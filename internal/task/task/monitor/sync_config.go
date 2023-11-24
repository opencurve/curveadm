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
* Created Date: 2023-04-21
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"fmt"
	"path"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	TOOL_SYS_PATH             = "/usr/bin/curve_ops_tool"
	MONITOR_CONF_PATH         = "monitor"
	PROMETHEUS_CONTAINER_PATH = "/etc/prometheus"
	GRAFANA_CONTAINER_PATH    = "/etc/grafana/grafana.ini"
	DASHBOARD_CONTAINER_PATH  = "/etc/grafana/provisioning/dashboards"
	GRAFANA_DATA_SOURCE_PATH  = "/etc/grafana/provisioning/datasources/all.yml"
	CURVE_MANAGER_CONF_PATH   = "/curve-manager/conf/pigeon.yaml"
)

func getNodeExporterAddrs(hosts []string, port int) string {
	endpoint := []string{}
	for _, item := range hosts {
		endpoint = append(endpoint, fmt.Sprintf("'%s:%d'", item, port))
	}
	return fmt.Sprintf("[%s]", strings.Join(endpoint, ","))
}

func NewSyncConfigTask(curveadm *cli.CurveAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(cfg.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if IsSkip(cfg, []string{ROLE_MONITOR_CONF, ROLE_NODE_EXPORTER}) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	role, host := cfg.GetRole(), cfg.GetHost()
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Sync Config", subname, hc.GetSSHConfig(), hc.GetHttpConfig())
	// add step to task
	var out string
	t.AddStep(&step.ListContainers{ // gurantee container exist
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.CheckContainerExist(cfg.GetHost(), cfg.GetRole(), containerId, &out),
	})
	if role == ROLE_PROMETHEUS {
		t.AddStep(&step.CreateAndUploadDir{ // prepare prometheus conf upath
			HostDirName:       "prometheus",
			ContainerDestId:   &containerId,
			ContainerDestPath: "/etc",
			ExecOptions:       curveadm.ExecOptions(),
		})
		content := fmt.Sprintf(scripts.PROMETHEUS_YML, cfg.GetListenPort(),
			getNodeExporterAddrs(cfg.GetNodeIps(), cfg.GetNodeListenPort()))
		t.AddStep(&step.InstallFile{ // install prometheus.yml file
			ContainerId:       &containerId,
			ContainerDestPath: path.Join(PROMETHEUS_CONTAINER_PATH, "prometheus.yml"),
			Content:           &content,
			ExecOptions:       curveadm.ExecOptions(),
		})
		target := cfg.GetPrometheusTarget()
		t.AddStep(&step.InstallFile{ // install target.json file
			ContainerId:       &containerId,
			ContainerDestPath: path.Join(PROMETHEUS_CONTAINER_PATH, "target.json"),
			Content:           &target,
			ExecOptions:       curveadm.ExecOptions(),
		})
	} else if role == ROLE_GRAFANA {
		serviceId = curveadm.GetServiceId(fmt.Sprintf("%s_%s", ROLE_MONITOR_CONF, cfg.GetHost()))
		confContainerId, err := curveadm.GetContainerId(serviceId)
		if err != nil {
			return nil, err
		}
		t.AddStep(&step.SyncFileDirectly{ // sync grafana.ini file
			ContainerSrcId:    &confContainerId,
			ContainerSrcPath:  path.Join("/", cfg.GetKind(), MONITOR_CONF_PATH, "grafana/grafana.ini"),
			ContainerDestId:   &containerId,
			ContainerDestPath: GRAFANA_CONTAINER_PATH,
			IsDir:             false,
			ExecOptions:       curveadm.ExecOptions(),
		})
		t.AddStep(&step.SyncFileDirectly{ // sync dashboard dir
			ContainerSrcId:    &confContainerId,
			ContainerSrcPath:  path.Join("/", cfg.GetKind(), MONITOR_CONF_PATH, "grafana/provisioning/dashboards"),
			ContainerDestId:   &containerId,
			ContainerDestPath: DASHBOARD_CONTAINER_PATH,
			IsDir:             true,
			ExecOptions:       curveadm.ExecOptions(),
		})
		content := fmt.Sprintf(scripts.GRAFANA_DATA_SOURCE, cfg.GetPrometheusIp(), cfg.GetPrometheusListenPort())
		t.AddStep(&step.InstallFile{ // install grafana datasource file
			ContainerId:       &containerId,
			ContainerDestPath: GRAFANA_DATA_SOURCE_PATH,
			Content:           &content,
			ExecOptions:       curveadm.ExecOptions(),
		})
	}
	return t, nil
}

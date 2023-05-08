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
* Created Date: 2023-04-19
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
)

func getArguments(cfg *configure.MonitorConfig) string {
	role := cfg.GetRole()
	var argsMap map[string]interface{}
	switch role {
	case ROLE_NODE_EXPORTER:
		argsMap = map[string]interface{}{
			"path.rootfs":        "/host",
			"collector.cpu.info": nil,
			"web.listen-address": fmt.Sprintf(":%d", cfg.GetListenPort()),
		}
	case ROLE_PROMETHEUS:
		argsMap = map[string]interface{}{
			"config.file":                 "/etc/prometheus/prometheus.yml",
			"storage.tsdb.path":           "/prometheus",
			"storage.tsdb.retention.time": cfg.GetPrometheusRetentionTime(),
			"storage.tsdb.retention.size": cfg.GetPrometheusRetentionSize(),
			"web.console.libraries":       "/usr/share/prometheus/console_libraries",
			"web.console.templates":       "/usr/share/prometheus/consoles",
			"web.listen-address":          fmt.Sprintf(":%d", cfg.GetListenPort()),
		}
	}
	args := []string{}
	for k, v := range argsMap {
		var item string
		if v != nil {
			item = fmt.Sprintf("--%s=%v", k, v)
		} else {
			item = fmt.Sprintf("--%s", k)
		}
		args = append(args, item)
	}
	return strings.Join(args, " ")
}

func getMountVolumes(cfg *configure.MonitorConfig) []step.Volume {
	role := cfg.GetRole()
	volumes := []step.Volume{}
	switch role {
	case ROLE_NODE_EXPORTER:
		volumes = append(volumes, step.Volume{
			HostPath:      "/",
			ContainerPath: "/host:ro,rslave",
		},
			step.Volume{
				HostPath:      "/run/udev/data",
				ContainerPath: "/run/udev/data",
			},
			step.Volume{
				HostPath:      "/run/dbus/system_bus_socket",
				ContainerPath: "/var/run/dbus/system_bus_socket:ro",
			})
	case ROLE_PROMETHEUS:
		volumes = append(volumes, step.Volume{
			HostPath:      cfg.GetDataDir(),
			ContainerPath: "/prometheus",
		})
	case ROLE_GRAFANA:
		volumes = append(volumes, step.Volume{
			HostPath:      cfg.GetDataDir(),
			ContainerPath: "/var/lib/grafana",
		})
	}
	return volumes
}

func getEnvironments(cfg *configure.MonitorConfig) []string {
	role := cfg.GetRole()
	if role == ROLE_GRAFANA {
		return []string{
			"GF_INSTALL_PLUGINS=grafana-piechart-panel",
			fmt.Sprintf("GF_SECURITY_ADMIN_USER=%s", cfg.GetGrafanaUser()),
			fmt.Sprintf("GF_SECURITY_ADMIN_PASSWORD=%s", cfg.GetGrafanaPassword()),
			fmt.Sprintf("GF_SERVER_HTTP_PORT=%d", cfg.GetListenPort()),
		}
	}
	return []string{}
}

func NewCreateContainerTask(curveadm *cli.CurveAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
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
	mcId := cfg.GetId()
	serviceId := curveadm.GetServiceId(mcId)
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
		Paths:       []string{cfg.GetDataDir()},
		ExecOptions: options,
	})
	t.AddStep(&step.CreateContainer{
		Image:       cfg.GetImage(),
		Command:     getArguments(cfg),
		Envs:        getEnvironments(cfg),
		Init:        true,
		Name:        hostname,
		Privileged:  true,
		User:        "0:0",
		Pid:         "host",
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

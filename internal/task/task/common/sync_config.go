/*
 *  Copyright (c) 2021 NetEase Inc.
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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package common

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	DEFAULT_CONFIG_DELIMITER  = "="
	ETCD_CONFIG_DELIMITER     = ": "
	TOOLS_V2_CONFIG_DELIMITER = ":"

	CURVE_CRONTAB_FILE = "/tmp/curve_crontab"
)

func NewMutate(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
	serviceConfig := dc.GetServiceConfig()
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			if forceRender { // only for nginx.conf
				out, err = dc.GetVariables().Rendering(in)
			}
			return
		}

		// replace config
		v, ok := serviceConfig[strings.ToLower(key)]
		if ok {
			value = v
		}

		// replace variable
		value, err = dc.GetVariables().Rendering(value)
		if err != nil {
			return
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func newToolV2Mutate(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
	serviceConfig := dc.GetServiceConfig()
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			if forceRender { // only for nginx.conf
				out, err = dc.GetVariables().Rendering(in)
			}
			return
		}

		// replace config
		v, ok := serviceConfig[strings.ToLower(key)]
		if ok {
			value = v
		}

		// replace variable
		value, err = dc.GetVariables().Rendering(value)
		if err != nil {
			return
		}

		out = fmt.Sprintf("%s%s %s", key, delimiter, value)
		return
	}
}

func newCrontab(uuid string, dc *topology.DeployConfig, reportScriptPath string) string {
	var period, command string
	if dc.GetReportUsage() {
		period = func(minute, hour, day, month, week string) string {
			return fmt.Sprintf("%s %s %s %s %s", minute, hour, day, month, week)
		}("0", "*", "*", "*", "*") // every hour

		command = func(format string, args ...interface{}) string {
			return fmt.Sprintf(format, args...)
		}("bash %s %s %s %s", reportScriptPath, dc.GetKind(), uuid, dc.GetRole())
	}

	return fmt.Sprintf("%s %s\n", period, command)
}

func NewSyncConfigTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Sync Config", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	layout := dc.GetProjectLayout()
	role := dc.GetRole()
	reportScript := scripts.SCRIPT_REPORT
	reportScriptPath := fmt.Sprintf("%s/report.sh", layout.ToolsBinDir)
	crontab := newCrontab(curveadm.ClusterUUId(), dc, reportScriptPath)
	delimiter := DEFAULT_CONFIG_DELIMITER
	if role == topology.ROLE_ETCD {
		delimiter = ETCD_CONFIG_DELIMITER
	}

	t.AddStep(&step.ListContainers{ // gurantee container exist
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: CheckContainerExist(dc.GetHost(), dc.GetRole(), containerId, &out),
	})
	for _, conf := range layout.ServiceConfFiles {
		t.AddStep(&step.SyncFile{ // sync service config
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  conf.SourcePath,
			ContainerDestId:   &containerId,
			ContainerDestPath: conf.Path,
			KVFieldSplit:      delimiter,
			Mutate:            NewMutate(dc, delimiter, conf.Name == "nginx.conf"),
			ExecOptions:       curveadm.ExecOptions(),
		})
	}
	t.AddStep(&step.SyncFile{ // sync tools config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  layout.ToolsConfSrcPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: layout.ToolsConfSystemPath,
		KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
		Mutate:            NewMutate(dc, DEFAULT_CONFIG_DELIMITER, false),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.TrySyncFile{ // sync toolsv2 config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  layout.ToolsV2ConfSrcPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: layout.ToolsV2ConfSystemPath,
		KVFieldSplit:      TOOLS_V2_CONFIG_DELIMITER,
		Mutate:            newToolV2Mutate(dc, TOOLS_V2_CONFIG_DELIMITER, false),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install report script
		ContainerId:       &containerId,
		ContainerDestPath: reportScriptPath,
		Content:           &reportScript,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install crontab file
		ContainerId:       &containerId,
		ContainerDestPath: CURVE_CRONTAB_FILE,
		Content:           &crontab,
		ExecOptions:       curveadm.ExecOptions(),
	})

	return t, nil
}

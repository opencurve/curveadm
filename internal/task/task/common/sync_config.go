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
	DEFAULT_CONFIG_DELIMITER = "="
	ETCD_CONFIG_DELIMITER    = ": "

	CURVE_CRONTAB_FILE = "/tmp/curve_crontab"
)

func newMutate(dc *topology.DeployConfig, delimiter string, forceRender bool) step.Mutate {
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

func newCrontab(uuid string, dc *topology.DeployConfig, reportScriptPath string) string {
	var period, command string
	if dc.GetReportUsage() == true {
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
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Sync Config", subname, dc.GetSSHConfig())

	// add step
	layout := dc.GetProjectLayout()
	role := dc.GetRole()
	reportScript := scripts.SCRIPT_REPORT
	reportScriptPath := fmt.Sprintf("%s/report.sh", layout.ToolsBinDir)
	crontab := newCrontab(curveadm.ClusterUUId(), dc, reportScriptPath)
	delimiter := DEFAULT_CONFIG_DELIMITER
	if role == topology.ROLE_ETCD {
		delimiter = ETCD_CONFIG_DELIMITER
	}

	for _, conf := range layout.ServiceConfFiles {
		t.AddStep(&step.SyncFile{ // sync service config
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  conf.SourcePath,
			ContainerDestId:   &containerId,
			ContainerDestPath: conf.Path,
			KVFieldSplit:      delimiter,
			Mutate:            newMutate(dc, delimiter, conf.Name == "nginx.conf"),
			ExecWithSudo:      true,
			ExecInLocal:       false,
			ExecSudoAlias:     curveadm.SudoAlias(),
		})
	}
	t.AddStep(&step.SyncFile{ // sync tools config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  layout.ToolsConfSrcPath,
		ContainerDestId:   &containerId,
		ContainerDestPath: layout.ToolsConfSystemPath,
		KVFieldSplit:      DEFAULT_CONFIG_DELIMITER,
		Mutate:            newMutate(dc, DEFAULT_CONFIG_DELIMITER, false),
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{ // install report script
		ContainerId:       &containerId,
		ContainerDestPath: reportScriptPath,
		Content:           &reportScript,
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{ // install crontab file
		ContainerId:       &containerId,
		ContainerDestPath: CURVE_CRONTAB_FILE,
		Content:           &crontab,
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})

	return t, nil
}

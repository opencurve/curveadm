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
* Created Date: 2023-04-27
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

var (
	ROLE_NODE_EXPORTER = configure.ROLE_NODE_EXPORTER
	ROLE_PROMETHEUS    = configure.ROLE_PROMETHEUS
	ROLE_GRAFANA       = configure.ROLE_GRAFANA
	ROLE_MONITOR_CONF  = configure.ROLE_MONITOR_CONF
)

func getCleanFiles(clean map[string]bool, cfg *configure.MonitorConfig) []string {
	files := []string{}
	for item := range clean {
		switch item {
		case comm.CLEAN_ITEM_DATA:
			files = append(files, cfg.GetDataDir())
		}
	}
	return files
}

func NewCleanMonitorTask(curveadm *cli.CurveAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(cfg.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}
	if cfg.GetRole() == ROLE_MONITOR_CONF &&
		(len(containerId) == 0 || containerId == comm.CLEANED_CONTAINER_ID) {
		return nil, nil
	}
	hc, err := curveadm.GetHost(cfg.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	only := curveadm.MemStorage().Get(comm.KEY_CLEAN_ITEMS).([]string)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Monitor", subname, hc.GetSSHConfig())

	// add step to task
	clean := utils.Slice2Map(only)
	files := getCleanFiles(clean, cfg) // directorys which need cleaned
	t.AddStep(&step.RemoveFile{
		Files:       files,
		ExecOptions: curveadm.ExecOptions(),
	})
	if clean[comm.CLEAN_ITEM_CONTAINER] {
		t.AddStep(&common.Step2CleanContainer{
			ServiceId:   serviceId,
			ContainerId: containerId,
			Storage:     curveadm.Storage(),
			ExecOptions: curveadm.ExecOptions(),
		})
	}
	return t, nil
}

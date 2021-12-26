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
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	KEY_CLEAN_ITEMS = "CLEAN_ITEMS"
	ITEM_LOG        = "log"
	ITEM_DATA       = "data"
	ITEM_CONTAINER  = "container"
)

type step2CleanContainer struct {
	config       *topology.DeployConfig
	serviceId    string
	containerId  string
	storage      *storage.Storage
	execWithSudo bool
	execInLocal  bool
}

func (s *step2CleanContainer) Execute(ctx *context.Context) error {
	containerId := s.containerId
	if containerId == "" { // container not created
		return nil
	} else if containerId == "-" { // container has removed
		return nil
	}

	cli := ctx.Module().DockerCli().RemoveContainer(s.containerId)
	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo: s.execWithSudo,
		ExecInLocal:  s.execInLocal,
	})

	// container has removed
	if err != nil && !strings.Contains(out, "No such container") {
		return err
	}
	return s.storage.SetContainId(s.serviceId, "-")
}

func getCleanDirs(clean map[string]bool, dc *topology.DeployConfig) []string {
	dirs := []string{}
	for item := range clean {
		switch item {
		case ITEM_LOG:
			dirs = append(dirs, dc.GetLogDir())
		case ITEM_DATA:
			dirs = append(dirs, dc.GetDataDir())
		}
	}
	return dirs
}

func NewCleanServiceTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	only := curveadm.MemStorage().Get(KEY_CLEAN_ITEMS).([]string)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Service", subname, dc.GetSSHConfig())

	// add step
	clean := utils.Slice2Map(only)
	t.AddStep(&step.RemoveFile{
		Files:        getCleanDirs(clean, dc),
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	if clean[ITEM_CONTAINER] == true {
		t.AddStep(&step2CleanContainer{
			config:       dc,
			serviceId:    serviceId,
			containerId:  containerId,
			storage:      curveadm.Storage(),
			execWithSudo: true,
			execInLocal:  false,
		})
	}

	return t, nil
}

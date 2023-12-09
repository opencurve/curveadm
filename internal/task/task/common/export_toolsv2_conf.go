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
 * Created Date: 2023-11-12
 * Author: Jiang Jun (youarefree123)
 */

package common

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func NewExportToolsV2ConfTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}

	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	var ToolsV2Conf string
	localPath := curveadm.MemStorage().Get(comm.KEY_TOOLSV2_CONF_PATH).(string)
	subname := fmt.Sprintf("output=%s", localPath)
	t := task.NewTask("Export curve.yaml", subname, hc.GetSSHConfig())

	t.AddStep(&step.ReadFile{
		ContainerId:      containerId,
		ContainerSrcPath: dc.GetProjectLayout().ToolsV2ConfSystemPath,
		Content:          &ToolsV2Conf,
		ExecOptions:      curveadm.ExecOptions(),
	})

	t.AddStep(&step.InstallFile{
		Content:      &ToolsV2Conf,
		HostDestPath: localPath,
		ExecOptions:  curveadm.ExecOptions(),
	})

	return t, nil
}

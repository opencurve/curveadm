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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	KEY_ADD_TARGET_OPTION = "ADD_TARGET_OPTION"
)

type AddTargetOption struct {
	User   string
	Volume string
	Create bool
	Size   int
}

func NewAddTargetTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	option := curveadm.MemStorage().Get(KEY_ADD_TARGET_OPTION).(AddTargetOption)
	user, volume := option.User, option.Volume
	subname := fmt.Sprintf("hostname=%s volume=%s", cc.GetHost(), volume)
	t := task.NewTask("Add Target", subname, cc.GetSSHConfig())

	// add step
	containerId := DEFAULT_TGTD_CONTAINER_NAME
	targetScriptPath := "/curvebs/tools/sbin/target.sh"
	targetScript := scripts.TARGET
	cmd := fmt.Sprintf("/bin/bash %s %s %s %v %d", targetScriptPath, user, volume, option.Create, option.Size)
	toolsConf := fmt.Sprintf(FORMAT_TOOLS_CONF, cc.GetClusterMDSAddr())

	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerId,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{ // install target.sh
		Content:           &targetScript,
		ContainerId:       &containerId,
		ContainerDestPath: targetScriptPath,
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId:   &containerId,
		Command:       cmd,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})

	return t, nil
}

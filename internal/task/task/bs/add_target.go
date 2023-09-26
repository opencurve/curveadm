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
	"regexp"
	"strconv"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

type TargetOption struct {
	Host           string
	User           string
	Volume         string
	Create         bool
	Size           int
	Tid            string
	Blocksize      uint64
	Devno          string
	OriginBdevname string
	Spdk           bool
}

func NewAddTargetTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_TARGET_OPTIONS).(TargetOption)
	user, volume := options.User, options.Volume
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s volume=%s", options.Host, volume)
	t := task.NewTask("Add Target", subname, hc.GetSSHConfig())

	// add step
	var output string
	containerId := DEFAULT_TGTD_CONTAINER_NAME
	targetScriptPath := "/curvebs/tools/sbin/target.sh"
	targetScript := scripts.TARGET
	cmd := fmt.Sprintf("/bin/bash %s %s %d %s %v %d %v",
		targetScriptPath, volume,
		options.Devno, options.OriginBdevname,
		options.Blocksize, user, options.Create, options.Size, options.Spdk)
	toolsConf := fmt.Sprintf(FORMAT_TOOLS_CONF, cc.GetClusterMDSAddr())

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}} {{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2CheckTgtdStatus{
		output: &output,
	})
	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerId,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install target.sh
		Content:           &targetScript,
		ContainerId:       &containerId,
		ContainerDestPath: targetScriptPath,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     cmd,
		ExecOptions: curveadm.ExecOptions(),
	})

	t.AddStep(&step.AddDaemonTask{ // install addTarget.task
		ContainerId: &containerId,
		Cmd:         "/bin/bash",
		Args:        []string{targetScriptPath, options.Volume, options.Devno,
				      options.OriginBdevname, strconv.FormatUint(options.Blocksize, 10),
				      user, strconv.FormatBool(options.Create), strconv.Itoa(options.Size), strconv.FormatBool(options.Spdk)},
		TaskName:    "addTarget" + TranslateVolumeName(volume, user),
		ExecOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

func TranslateVolumeName(volume, user string) string {
	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	return reg.ReplaceAllString(volume, "_") + "_" + user + "_"
}

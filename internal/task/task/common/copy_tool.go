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
 * Created Date: 2023-12-25
 * Author: Xinyu Zhuo (0fatal)
 */

package common

import (
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/pkg/module"
	"path/filepath"
)

func checkPathExist(path string, sshConfig *module.SSHConfig, curveadm *cli.CurveAdm) error {
	sshClient, err := module.NewSSHClient(*sshConfig)
	if err != nil {
		return errno.ERR_SSH_CONNECT_FAILED.E(err)
	}

	module := module.NewModule(sshClient)
	cmd := module.Shell().Stat(path)
	if _, err := cmd.Execute(curveadm.ExecOptions()); err == nil {
		if pass := tui.ConfirmYes(tui.PromptPathExist(path)); !pass {
			return errno.ERR_CANCEL_OPERATION
		}
	}
	return nil
}

func NewCopyToolTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	layout := dc.GetProjectLayout()
	path := curveadm.MemStorage().Get(comm.KEY_COPY_PATH).(string)
	confPath := curveadm.MemStorage().Get(comm.KEY_COPY_CONF_PATH).(string)
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}

	if err = checkPathExist(path, hc.GetSSHConfig(), curveadm); err != nil {
		return nil, err
	}
	if err = checkPathExist(confPath, hc.GetSSHConfig(), curveadm); err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s", dc.GetHost())
	t := task.NewTask("Copy version 2 tool to host", subname, hc.GetSSHConfig())

	t.AddStep(&step.CreateDirectory{
		Paths:       []string{filepath.Dir(path)},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2BinaryPath,
		ContainerId:      containerId,
		HostDestPath:     path,
		ExecOptions:      curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{filepath.Dir(confPath)},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CopyFromContainer{
		ContainerSrcPath: layout.ToolsV2ConfSystemPath,
		ContainerId:      containerId,
		HostDestPath:     confPath,
		ExecOptions:      curveadm.ExecOptions(),
	})

	return t, nil
}

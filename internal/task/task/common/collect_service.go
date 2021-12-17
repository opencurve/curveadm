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
 * Created Date: 2021-11-26
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"path/filepath"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/module"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	TEMP_DIR = "/tmp"

	KEY_COLLECT_SAVE_DIR   = "COLLECT_SAVE_DIR"
	KEY_COLLECT_SECRET_KEY = "COLLECT_SECRET_KEY"
)

type (
	step2CollectService struct {
		prefix      string
		script      string
		saveDir     string
		serviceId   string
		containerId string
	}
)

func (s *step2CollectService) Execute(ctx *context.Context) error {
	script := s.script
	prefix := s.prefix
	containerId := s.containerId
	serviceId := s.serviceId
	localPath := fmt.Sprintf("%s/%s.tar.gz", s.saveDir, serviceId)
	remoteSaveDir := fmt.Sprintf("/tmp/%s_%s", serviceId, utils.RandString(5))
	remotePath := fmt.Sprintf("%s.tar.gz", remoteSaveDir)
	cmd1 := fmt.Sprintf("bash %s %s %s %s", script, prefix, containerId, remoteSaveDir)
	cmd2 := fmt.Sprintf("cd %s && tar -zcvf %s %s",
		filepath.Dir(remotePath), filepath.Base(remotePath), filepath.Base(remoteSaveDir))
	option := module.ExecOption{ExecWithSudo: false, ExecInLocal: false}
	defer func() {
		ctx.Module().Shell().Remove(remoteSaveDir).AddOption("--parents").Execute(option)
		ctx.Module().Shell().Remove(remotePath).AddOption("--parents").Execute(option)
		ctx.Module().Shell().Remove(script).AddOption("--parents").Execute(option)
	}()

	if _, err := ctx.Module().Shell().Command(cmd1).Execute(option); err != nil {
		return err
	} else if _, err := ctx.Module().Shell().Command(cmd2).Execute(option); err != nil {
		return err
	} else if err := ctx.Module().File().Download(remotePath, localPath); err != nil {
		return err
	}

	return nil
}

func (s *step2CollectService) Rollback(ctx *context.Context) {
}

func NewCollectServiceTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, nil
	}

	memStorage := curveadm.MemStorage()
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Collect Service", subname, dc.GetSshConfig())

	// add step
	collectScript := scripts.Get("collect")
	collectScriptPath := utils.RandFilename(TEMP_DIR)
	t.AddStep(&step.InstallFile{
		HostDestPath: collectScriptPath,
		Content:      &collectScript,
	})
	t.AddStep(&step2CollectService{
		prefix:      dc.GetServicePrefix(),
		script:      collectScriptPath,
		saveDir:     memStorage.Get(KEY_COLLECT_SAVE_DIR).(string),
		serviceId:   serviceId,
		containerId: containerId,
	})
	return t, nil
}

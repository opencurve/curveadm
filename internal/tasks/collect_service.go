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

package tasks

import (
	"fmt"
	"path/filepath"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	KEY_COLLECT_SAVE_DIR   = "COLLECT_SAVE_DIR"
	KEY_COLLECT_SECRET_KEY = "COLLECT_SECRET_KEY"
)

type (
	step2CollectService struct {
		saveDir     string
		serviceId   string
		containerId string
	}
)

func (s *step2CollectService) Execute(ctx *task.Context) error {
	script := fmt.Sprintf("/tmp/collect_service_%s.sh", utils.RandString(5))
	prefix := ctx.Config().GetServicePrefix()
	containerId := s.containerId
	serviceId := s.serviceId
	localPath := fmt.Sprintf("%s/%s.tar.gz", s.saveDir, serviceId)
	remoteSaveDir := fmt.Sprintf("/tmp/%s_%s", serviceId, utils.RandString(5))
	remotePath := fmt.Sprintf("%s.tar.gz", remoteSaveDir)
	cmd1 := fmt.Sprintf("bash %s %s %s %s", script, prefix, containerId, remoteSaveDir)
	cmd2 := fmt.Sprintf("cd %s && tar -zcvf %s %s",
		filepath.Dir(remotePath), filepath.Base(remotePath), filepath.Base(remoteSaveDir))
	defer func() {
		ctx.Module().SshRemovePath(remoteSaveDir, false)
		ctx.Module().SshRemovePath(remotePath, false)
		ctx.Module().SshRemovePath(script, false)
	}()

	if err := ctx.Module().SshMountScript("collect", script); err != nil {
		return err
	} else if _, err := ctx.Module().SshShell(cmd1); err != nil {
		return err
	} else if _, err := ctx.Module().SshShell(cmd2); err != nil {
		return err
	} else if err := ctx.Module().Download(remotePath, localPath); err != nil {
		return err
	}

	return nil
}

func (s *step2CollectService) Rollback(ctx *task.Context) {
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
	t := task.NewTask("Collect Service", subname, dc)
	t.AddStep(&step2CollectService{
		saveDir:     memStorage.Get(KEY_COLLECT_SAVE_DIR).(string),
		serviceId:   serviceId,
		containerId: containerId,
	})
	return t, nil
}

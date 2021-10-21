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

package tasks

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	KEY_CLEAN_ITEMS = "CLEAN_ITEMS"
	KEY_LOG         = "log"
	KEY_DATA        = "data"
	KEY_CONTAINER   = "container"
)

type (
	step2CleanService struct {
		serviceId   string
		containerId string
		clean       map[string]bool
		storage     *storage.Storage
	}
)

func (s *step2CleanService) cleanLogDir(ctx *task.Context) error {
	if !s.clean[KEY_LOG] {
		return nil
	}
	return ctx.Module().SshRemoveDir(ctx.Config().GetLogDir(), true)
}

func (s *step2CleanService) cleanDataDir(ctx *task.Context) error {
	if !s.clean[KEY_DATA] {
		return nil
	}
	return ctx.Module().SshRemoveDir(ctx.Config().GetDataDir(), true)
}

func (s *step2CleanService) cleanContainer(ctx *task.Context) error {
	containerId := s.containerId
	if !s.clean[KEY_CONTAINER] || containerId == "" || containerId == "-" {
		return nil
	}

	if out, err := ctx.Module().SshShell("sudo docker rm %s", containerId); err != nil &&
		!strings.Contains(out, "No such container") {
		return err
	} else if err := s.storage.SetConatinId(s.serviceId, "-"); err != nil {
		return err
	}
	return nil
}

func (s *step2CleanService) Execute(ctx *task.Context) error {
	if err := s.cleanLogDir(ctx); err != nil {
		return err
	} else if err := s.cleanDataDir(ctx); err != nil {
		return err
	} else if err := s.cleanContainer(ctx); err != nil {
		return err
	}
	return nil
}

func (s *step2CleanService) Rollback(ctx *task.Context) {
}

func NewCleanServiceTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	v, _ := curveadm.MemStorage().Get(KEY_CLEAN_ITEMS)
	only := v.([]string)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Service", subname, dc)
	t.AddStep(&step2CleanService{
		serviceId:   serviceId,
		containerId: containerId,
		clean:       utils.Slice2Map(only),
		storage:     curveadm.Storage(),
	})
	return t, nil
}

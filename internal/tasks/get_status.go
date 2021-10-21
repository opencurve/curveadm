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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2GetContainerStatus struct {
		serviceId   string
		containerId string
		memStorage  *utils.SafeMap
	}
)

type ServiceStatus struct {
	Id          string
	Role        string
	Host        string
	ContainerId string
	Status      string
	LogDir      string
	DataDir     string
}

// see also: https://docs.docker.com/engine/reference/commandline/ps/
func (s *step2GetContainerStatus) getContainerStatus(ctx *task.Context, containerId string) string {
	format := `"{{.Status}}"`
	out, err := ctx.Module().SshShell("sudo docker ps -a --format %s --filter id=%s", format, containerId)
	if err != nil {
		log.Error("GetContainerStatus",
			log.Field("containerId", containerId),
			log.Field("error", err))
		return "-"
	} else if len(out) == 0 {
		log.Warn("GetContainerStatus",
			log.Field("containerId", containerId),
			log.Field("error", "not found"))
		return "Losed"
	}

	return utils.TrimNewline(out)
}

func (s *step2GetContainerStatus) Execute(ctx *task.Context) error {
	status := "-"
	if s.containerId == "-" { // container has removed
		status = "Cleaned"
	} else {
		status = s.getContainerStatus(ctx, s.containerId)
	}

	id := s.serviceId
	config := ctx.Config()
	s.memStorage.Set(id, ServiceStatus{
		Id:          id,
		Role:        config.GetRole(),
		Host:        config.GetHost(),
		ContainerId: tui.TrimContainerId(s.containerId),
		Status:      status,
		LogDir:      config.GetLogDir(),
		DataDir:     config.GetDataDir(),
	})
	return nil
}

func (s *step2GetContainerStatus) Rollback(ctx *task.Context) {

}

func NewGetServiceStatusTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Get Service Status", subname, dc)
	t.AddStep(&step2GetContainerStatus{
		serviceId:   serviceId,
		containerId: containerId,
		memStorage:  curveadm.MemStorage(),
	})
	return t, nil
}

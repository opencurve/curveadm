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
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errors"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	STATUS_CLEANED = "Cleaned"
	STATUS_LOSED   = "Losed"
)

type (
	step2FormatStatus struct {
		config      *topology.DeployConfig
		serviceId   string
		containerId string
		status      *string
		memStorage  *utils.SafeMap
		mds         *string
	}

	ServiceStatus struct {
		Id               string
		ParentId         string
		Role             string
		Host             string
		ListenPort       string
		ListenDummyPort  string
		ListenProxyPort  string
		ListenClientPort string
		Replica          string
		ContainerId      string
		Status           string
		LogDir           string
		DataDir          string
		SortedKey        string
	}
)

func (s *step2FormatStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if s.containerId == "-" { // container cleaned
		status = STATUS_CLEANED
	} else if len(status) == 0 { // container losed
		status = STATUS_LOSED
	}

	id := s.serviceId
	config := s.config

	var role string = config.GetRole()
	mds := strings.Split(*s.mds, ":")
	if len(mds) == 2 &&
		strings.TrimSpace(mds[0]) == strings.TrimSpace(config.GetHost()) &&
		strings.TrimSpace(mds[1]) == strings.TrimSpace(strconv.Itoa(config.GetListenPort())) {
		role = role + "*"
	}

	s.memStorage.Set(id, ServiceStatus{
		Id:               id,
		ParentId:         config.GetParentId(),
		Role:             role,
		Host:             config.GetHost(),
		ListenPort:       strconv.Itoa(config.GetListenPort()),
		ListenDummyPort:  strconv.Itoa(config.GetListenDummyPort()),
		ListenProxyPort:  strconv.Itoa(config.GetListenProxyPort()),
		ListenClientPort: strconv.Itoa(config.GetListenClientPort()),
		Replica:          fmt.Sprintf("1/%d", config.GetReplica()),
		ContainerId:      tui.TrimContainerId(s.containerId),
		Status:           status,
		LogDir:           config.GetLogDir(),
		DataDir:          config.GetDataDir(),
		SortedKey:        config.GetId(),
	})
	return nil
}

func NewGetServiceStatusTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if len(containerId) == 0 {
		return nil, errors.ERR_SERVICE_NOT_FOUND.Format(serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Get Service Status", subname, dc.GetSSHConfig())

	// add step
	var status string
	var out string
	if dc.GetRole() == "mds" {
		t.AddStep(&step.ContainerExec{
			ContainerId:   &containerId,
			Command:       fmt.Sprintf("curve_ops_tool mds-status | grep \"current MDS:\" | grep -o \"[0-9]*\\.[0-9]*\\.[0-9]*\\.[0-9]*:[0-9]*\""),
			Out:           &out,
			ExecWithSudo:  true,
			ExecInLocal:   false,
			ExecSudoAlias: curveadm.SudoAlias(),
		})
	}
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        `"{{.Status}}"`,
		Filter:        fmt.Sprintf("id=%s", containerId),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
		Out:           &status,
	})
	t.AddStep(&step2FormatStatus{
		config:      dc,
		serviceId:   serviceId,
		containerId: containerId,
		status:      &status,
		memStorage:  curveadm.MemStorage(),
		mds:         &out,
	})

	return t, nil
}

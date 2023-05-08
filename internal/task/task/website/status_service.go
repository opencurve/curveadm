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
* Project: Curveadm
* Created Date: 2023-05-08
* Author: wanghai (SeanHai)
 */

package website

import (
	"fmt"
	"strconv"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

type step2InitWebsiteStatus struct {
	wc          *configure.WebsiteConfig
	serviceId   string
	containerId string
	memStorage  *utils.SafeMap
}

type step2FormatWebsiteStatus struct {
	wc          *configure.WebsiteConfig
	serviceId   string
	containerId string
	ports       *string
	status      *string
	memStorage  *utils.SafeMap
}

type WebsiteStatus struct {
	Id          string
	Role        string
	Host        string
	ContainerId string
	Ports       string
	Status      string
	DataDir     string
	LogDir      string
	Config      *configure.WebsiteConfig
}

func setWebsiteStatus(memStorage *utils.SafeMap, id string, status WebsiteStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]WebsiteStatus{}
		v := kv.Get(comm.KEY_WEBSITE_STATUS)
		if v != nil {
			m = v.(map[string]WebsiteStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_WEBSITE_STATUS, m)
		return nil
	})
}

func (s *step2InitWebsiteStatus) Execute(ctx *context.Context) error {
	wc := s.wc
	id := s.serviceId
	setWebsiteStatus(s.memStorage, id, WebsiteStatus{
		Id:          id,
		Role:        wc.GetRole(),
		Host:        wc.GetHost(),
		ContainerId: tui.TrimContainerId(s.containerId),
		Status:      comm.SERVICE_STATUS_UNKNOWN,
		DataDir:     wc.GetDataDir(),
		LogDir:      wc.GetLogDir(),
		Config:      wc,
	})
	return nil
}

func (s *step2FormatWebsiteStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if s.containerId == comm.CLEANED_CONTAINER_ID { // container cleaned
		status = comm.SERVICE_STATUS_CLEANED
	} else if len(status) == 0 { // container losed
		status = comm.SERVICE_STATUS_LOSED
	}

	wc := s.wc
	id := s.serviceId
	setWebsiteStatus(s.memStorage, id, WebsiteStatus{
		Id:          id,
		Role:        wc.GetRole(),
		Host:        wc.GetHost(),
		ContainerId: tui.TrimContainerId(s.containerId),
		Ports:       *s.ports,
		Status:      status,
		DataDir:     wc.GetDataDir(),
		LogDir:      wc.GetLogDir(),
		Config:      wc,
	})
	return nil
}

func NewInitWebsiteStatusTask(curveadm *cli.CurveAdm, cfg *configure.WebsiteConfig) (*task.Task, error) {
	serviceId := curveadm.GetWebsiteServiceId(cfg.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Init Website Status", subname, nil)

	t.AddStep(&step2InitWebsiteStatus{
		wc:          cfg,
		serviceId:   serviceId,
		containerId: containerId,
		memStorage:  curveadm.MemStorage(),
	})

	return t, nil
}

func NewGetWebsiteStatusTask(curveadm *cli.CurveAdm, cfg *configure.WebsiteConfig) (*task.Task, error) {
	serviceId := curveadm.GetWebsiteServiceId(cfg.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return nil, err
	}
	hc, err := curveadm.GetHost(cfg.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		cfg.GetHost(), cfg.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Get Website Status", subname, hc.GetSSHConfig())

	// add step to task
	var status string
	ports := strconv.Itoa(cfg.GetListenPort())
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.Status}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: common.TrimContainerStatus(&status),
	})
	t.AddStep(&step2FormatWebsiteStatus{
		wc:          cfg,
		serviceId:   serviceId,
		containerId: containerId,
		ports:       &ports,
		status:      &status,
		memStorage:  curveadm.MemStorage(),
	})
	return t, nil
}

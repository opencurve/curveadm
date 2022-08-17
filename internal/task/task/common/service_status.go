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

// __SIGN_BY_WINE93__

package common

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	SIGNATURE_LEADER          = "leader"
	URL_CURVEBS_METRIC_LEADER = "http://%s:%d/vars/mds_status?console=1"
	URL_CURVEFS_METRIC_LEADER = "http://%s:%d/vars/curvefs_mds_status?console=1"
	COMMAND_CURL_MDS          = "curl %s --connect-timeout 1 --max-time 3"
)

type (
	step2InitStatus struct {
		dc          *topology.DeployConfig
		serviceId   string
		containerId string
		memStorage  *utils.SafeMap
	}

	step2GetListenPorts struct {
		dc          *topology.DeployConfig
		containerId string
		status      *string
		ports       *string
		execOptions module.ExecOptions
	}

	step2GetLeader struct {
		dc          *topology.DeployConfig
		containerId string
		status      *string
		isLeader    *bool
		execOptions module.ExecOptions
	}

	step2FormatServiceStatus struct {
		dc          *topology.DeployConfig
		serviceId   string
		containerId string
		isLeader    *bool
		ports       *string
		status      *string
		memStorage  *utils.SafeMap
	}

	ServiceStatus struct {
		Id          string
		ParentId    string
		Role        string
		Host        string
		Replica     string
		ContainerId string
		Ports       string
		IsLeader    bool
		Status      string
		LogDir      string
		DataDir     string
		Config      *topology.DeployConfig
	}
)

func setServiceStatus(memStorage *utils.SafeMap, id string, status ServiceStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]ServiceStatus{}
		v := kv.Get(comm.KEY_ALL_SERVICE_STATUS)
		if v != nil {
			m = v.(map[string]ServiceStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_SERVICE_STATUS, m)
		return nil
	})
}

func (s *step2InitStatus) Execute(ctx *context.Context) error {
	dc := s.dc
	id := s.serviceId
	setServiceStatus(s.memStorage, id, ServiceStatus{
		Id:          id,
		ParentId:    dc.GetParentId(),
		Role:        dc.GetRole(),
		Host:        dc.GetHost(),
		Replica:     fmt.Sprintf("1/%d", dc.GetReplicas()),
		ContainerId: tui.TrimContainerId(s.containerId),
		Status:      comm.SERVICE_STATUS_UNKNOWN,
		LogDir:      dc.GetLogDir(),
		DataDir:     dc.GetDataDir(),
		Config:      dc,
	})
	return nil
}

func (s *step2GetListenPorts) extractPort(line string) string {
	// e.g: tcp LISTEN 0 128 10.246.159.123:2379 *:* users:(("etcd",pid=7,fd=5))
	regex, err := regexp.Compile("^.*:([0-9]+).*users.*$")
	if err == nil {
		mu := regex.FindStringSubmatch(line)
		if len(mu) > 0 {
			return mu[1]
		}
	}
	return ""
}

func (s *step2GetListenPorts) Execute(ctx *context.Context) error {
	if !strings.HasPrefix(*s.status, "Up") {
		return nil
	}

	// execute "ss" command in container
	cli := ctx.Module().Shell().SocketStatistics("")
	cli.AddOption("--no-header")
	cli.AddOption("--processes")
	cli.AddOption("--listening")
	command, err := cli.String()
	if err != nil {
		return nil
	}

	cmd := ctx.Module().DockerCli().ContainerExec(s.containerId, command)
	out, err := cmd.Execute(s.execOptions)
	if err != nil {
		return nil
	}

	// handle output
	ports := []string{}
	lines := strings.Split(out, "\n")
	for _, line := range lines {
		port := s.extractPort(line)
		if len(port) > 0 {
			ports = append(ports, port)
		}
	}
	*s.ports = strings.Join(ports, ",")
	return nil
}

func (s *step2GetLeader) Execute(ctx *context.Context) error {
	dc := s.dc
	if !strings.HasPrefix(*s.status, "Up") {
		return nil
	} else if dc.GetRole() != topology.ROLE_MDS {
		return nil
	}

	url := utils.Choose(dc.GetKind() == topology.KIND_CURVEBS,
		URL_CURVEBS_METRIC_LEADER, URL_CURVEFS_METRIC_LEADER)
	url = fmt.Sprintf(url, dc.GetListenIp(), dc.GetListenDummyPort())
	command := fmt.Sprintf(COMMAND_CURL_MDS, url)
	cmd := ctx.Module().DockerCli().ContainerExec(s.containerId, command)
	out, _ := cmd.Execute(s.execOptions)
	*s.isLeader = strings.Contains(out, SIGNATURE_LEADER)
	return nil
}

func (s *step2FormatServiceStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if s.containerId == comm.CLEANED_CONTAINER_ID { // container cleaned
		status = comm.SERVICE_STATUS_CLEANED
	} else if len(status) == 0 { // container losed
		status = comm.SERVICE_STATUS_LOSED
	}

	dc := s.dc
	id := s.serviceId
	setServiceStatus(s.memStorage, id, ServiceStatus{
		Id:          id,
		ParentId:    dc.GetParentId(),
		Role:        dc.GetRole(),
		Host:        dc.GetHost(),
		Replica:     fmt.Sprintf("1/%d", dc.GetReplicas()),
		ContainerId: tui.TrimContainerId(s.containerId),
		Ports:       *s.ports,
		IsLeader:    *s.isLeader,
		Status:      status,
		LogDir:      dc.GetLogDir(),
		DataDir:     dc.GetDataDir(),
		Config:      dc,
	})
	return nil
}

func NewInitServiceStatusTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if errors.Is(err, errno.ERR_SERVICE_CONTAINER_ID_NOT_FOUND) && // FIXME
		dc.GetRole() == topology.ROLE_SNAPSHOTCLONE {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Init Service Status", subname, nil)

	t.AddStep(&step2InitStatus{
		dc:          dc,
		serviceId:   serviceId,
		containerId: containerId,
		memStorage:  curveadm.MemStorage(),
	})

	return t, nil
}

func trimContainerStatus(status *string) step.LambdaType {
	return func(ctx *context.Context) error {
		items := strings.Split(*status, "\n")
		*status = items[len(items)-1]
		return nil
	}
}

func NewGetServiceStatusTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if errors.Is(err, errno.ERR_SERVICE_CONTAINER_ID_NOT_FOUND) && // FIXME
		dc.GetRole() == topology.ROLE_SNAPSHOTCLONE {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Get Service Status", subname, hc.GetSSHConfig())

	// add step to task
	var status string
	var ports string
	var isLeader bool
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.Status}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: trimContainerStatus(&status),
	})
	t.AddStep(&step2GetListenPorts{
		dc:          dc,
		containerId: containerId,
		status:      &status,
		ports:       &ports,
		execOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2GetLeader{
		dc:          dc,
		containerId: containerId,
		status:      &status,
		isLeader:    &isLeader,
		execOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2FormatServiceStatus{
		dc:          dc,
		serviceId:   serviceId,
		containerId: containerId,
		isLeader:    &isLeader,
		ports:       &ports,
		status:      &status,
		memStorage:  curveadm.MemStorage(),
	})

	return t, nil
}

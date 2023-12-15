/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-07-31
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

type (
	step2FormatClientStatus struct {
		client      storage.Client
		containerId string
		status      *string
		memStorage  *utils.SafeMap
		address     *string
	}

	step2GetAddress struct {
		containerId string
		address     *string
		execOptions module.ExecOptions
	}

	ClientStatus struct {
		Id          string
		Host        string
		Kind        string
		ContainerId string
		Status      string
		Address     string
		AuxInfo     string
	}
)

// TODO(P0): init client status
func setClientStatus(memStorage *utils.SafeMap, id string, status ClientStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]ClientStatus{}
		v := kv.Get(comm.KEY_ALL_CLIENT_STATUS)
		if v != nil {
			m = v.(map[string]ClientStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_CLIENT_STATUS, m)
		return nil
	})
}

func (s *step2FormatClientStatus) Execute(ctx *context.Context) error {
	status := *s.status
	address := *s.address
	if len(status) == 0 { // container losed
		status = comm.CLIENT_STATUS_LOSED
	}

	client := s.client
	id := client.Id
	setClientStatus(s.memStorage, id, ClientStatus{
		Id:          client.Id,
		Host:        client.Host,
		Kind:        client.Kind,
		ContainerId: s.containerId,
		Status:      status,
		Address:     address,
		AuxInfo:     client.AuxInfo,
	})
	return nil
}

func (s *step2GetAddress) Execute(ctx *context.Context) error {
	cmd := ctx.Module().DockerCli().TopContainer(s.containerId)
	out, err := cmd.Execute(s.execOptions)
	if err != nil {
		return err
	}

	lines := strings.Split(out, "\n")
	var pid string
	if len(lines) > 1 {
		reg := regexp.MustCompile(`\s+`)
		res := reg.Split(lines[1], -1)
		if len(res) > 1 {
			pid = res[1]
		}
	}

	if len(pid) == 0 {
		return nil
	}

	// execute "ss" command in container
	cli := ctx.Module().Shell().SocketStatistics("")
	cli.AddOption("--no-header")
	cli.AddOption("--processes")
	cli.AddOption("--listening")
	command, err := cli.String()
	if err != nil {
		return err
	}

	cmd = ctx.Module().DockerCli().ContainerExec(s.containerId, command)
	out, err = cmd.Execute(s.execOptions)
	if err != nil {
		return err
	}

	// handle output
	lines = strings.Split(out, "\n")
	for _, line := range lines {
		address := s.extractAddress(line, pid)
		if len(address) > 0 {
			*s.address = address
			return nil
		}
	}

	return nil
}

// e.g: tcp LISTEN 0 128 10.246.159.123:2379 *:* users:(("etcd",pid=7,fd=5))
// e.g: tcp LISTEN 0 128 *:2379 *:* users:(("etcd",pid=7,fd=5))
func (s *step2GetAddress) extractAddress(line, pid string) string {
	regex, err := regexp.Compile(`^.* ((\d+\.\d+\.\d+\.\d+)|\*:\d+).*pid=` + pid + ".*$")
	if err == nil {
		mu := regex.FindStringSubmatch(line)
		if len(mu) > 1 {
			return mu[1]
		}
	}
	return ""
}

func NewGetClientStatusTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	client := v.(storage.Client)
	hc, err := curveadm.GetHost(client.Host)
	if err != nil {
		return nil, err
	}

	containerId := client.ContainerId
	subname := fmt.Sprintf("host=%s kind=%s containerId=%s",
		hc.GetHost(), client.Kind, tui.TrimContainerId(containerId))
	t := task.NewTask("Get Client Status", subname, hc.GetSSHConfig())

	// add step
	var status string
	var address string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.Status}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2GetAddress{
		containerId: containerId,
		address:     &address,
		execOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2FormatClientStatus{
		client:      client,
		containerId: containerId,
		status:      &status,
		memStorage:  curveadm.MemStorage(),
		address:     &address,
	})

	return t, nil
}

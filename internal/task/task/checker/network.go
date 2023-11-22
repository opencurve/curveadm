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
 * Created Date: 2022-07-14
 * Author: Jingli Chen (Wine93)
 */

package checker

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	FORMAT_FILTER_SPORT = "( sport = :%d )"

	HTTP_SERVER_CONTAINER_NAME = "curveadm-precheck-nginx"
	CHECK_PORT_CONTAINER_NAME  = "curveadm-precheck-port"
)

// TASK: check port in use
func checkPortInUse(success *bool, out *string, host string, port int) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_GET_CONNECTION_INFORMATION_FAILED.S(*out)
		}

		if len(*out) > 0 {
			if *out == "RTNETLINK answers: Invalid argument" {
				return nil
			}

			return errno.ERR_PORT_ALREADY_IN_USE.
				F("host=%s, port=%d", host, port)
		}

		return nil
	}
}

func joinPorts(dc *topology.DeployConfig, addresses []Address) string {
	ports := []string{}
	for _, address := range addresses {
		ports = append(ports, strconv.Itoa(address.Port))
	}
	if dc.GetInstances() > 1 { // instances service
		ports = append(ports, "...")
	}
	return strings.Join(ports, ",")
}

func getCheckPortContainerName(curveadm *cli.CurveAdm, dc *topology.DeployConfig) string {
	return fmt.Sprintf("%s-%s-%s",
		CHECK_PORT_CONTAINER_NAME,
		dc.GetRole(),
		curveadm.GetServiceId(dc.GetId()))
}

type step2CheckPortStatus struct {
	containerId *string
	success     *bool
	dc          *topology.DeployConfig
	curveadm    *cli.CurveAdm
	port        int
}

// execute the "ss" command within a temporary container
func (s *step2CheckPortStatus) Execute(ctx *context.Context) error {
	filter := fmt.Sprintf(FORMAT_FILTER_SPORT, s.port)
	cli := ctx.Module().Shell().SocketStatistics(filter)
	cli.AddOption("--no-header")
	cli.AddOption("--listening")
	command, err := cli.String()
	if err != nil {
		return err
	}

	var out string
	steps := []task.Step{}
	steps = append(steps, &step.ContainerExec{
		ContainerId: s.containerId,
		Command:     command,
		Success:     s.success,
		Out:         &out,
		ExecOptions: s.curveadm.ExecOptions(),
	})
	steps = append(steps, &step.Lambda{
		Lambda: checkPortInUse(s.success, &out, s.dc.GetHost(), s.port),
	})

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCheckPortInUseTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	addresses := getServiceListenAddresses(dc)
	subname := fmt.Sprintf("host=%s role=%s ports={%s}",
		dc.GetHost(), dc.GetRole(), joinPorts(dc, addresses))
	t := task.NewTask("Check Port In Use <network>", subname, hc.GetSSHConfig())

	var containerId, out string
	var success bool
	t.AddStep(&step.PullImage{
		Image:       dc.GetContainerImage(),
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateContainer{
		Image:       dc.GetContainerImage(),
		Command:     "-c 'sleep infinity'", // keep the container running
		Entrypoint:  "/bin/bash",
		Name:        getCheckPortContainerName(curveadm, dc),
		Remove:      true,
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})

	for _, address := range addresses {
		t.AddStep(&step2CheckPortStatus{
			containerId: &containerId,
			success:     &success,
			dc:          dc,
			curveadm:    curveadm,
			port:        address.Port,
		})
	}

	return t, nil
}

// TASK: check destination reachable
func unique(address []Address) []string {
	out := []string{}
	m := map[string]bool{}
	for _, address := range address {
		if !m[address.IP] {
			out = append(out, address.IP)
		}
		m[address.IP] = true
	}
	return out
}

func checkReachable(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_DESTINATION_UNREACHABLE.S(*out)
	}
}

func NewCheckDestinationReachableTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	dcs := curveadm.MemStorage().Get(comm.KEY_ALL_DEPLOY_CONFIGS).([]*topology.DeployConfig)
	addresses := unique(getServiceConnectAddress(dc, dcs))
	subname := fmt.Sprintf("host=%s role=%s ping={%s}",
		dc.GetHost(), dc.GetRole(), tui.TrimAddress(strings.Join(addresses, ",")))
	t := task.NewTask("Check Destination Reachable <network>", subname, hc.GetSSHConfig())

	var out string
	var success bool
	for _, address := range addresses {
		t.AddStep(&step.Ping{
			Destination: &address,
			Count:       1,
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step.Lambda{
			Lambda: checkReachable(&success, &out),
		})
	}

	return t, nil
}

// TASK: start http server
func getNginxListens(dc *topology.DeployConfig) string {
	addresses := getServiceListenAddresses(dc)
	listens := []string{}
	for _, address := range addresses {
		listens = append(listens, fmt.Sprintf("listen %s:%d;",
			address.IP, address.Port))
	}
	return strings.Join(listens, " ")
}

func getHTTPServerContainerName(curveadm *cli.CurveAdm, dc *topology.DeployConfig) string {
	return fmt.Sprintf("%s-%s-%s",
		HTTP_SERVER_CONTAINER_NAME,
		dc.GetRole(),
		curveadm.GetServiceId(dc.GetId()))
}

func waitNginxStarted(seconds int) step.LambdaType {
	return func(ctx *context.Context) error {
		time.Sleep(time.Duration(seconds) * time.Second)
		return nil
	}
}

func NewStartHTTPServerTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// add task
	addresses := getServiceListenAddresses(dc)
	subname := fmt.Sprintf("host=%s role=%s ports={%s}",
		dc.GetHost(), dc.GetRole(), joinPorts(dc, addresses))
	t := task.NewTask("Start Mock HTTP Server <network>", subname, hc.GetSSHConfig())

	// add step to task
	var containerId, out string
	var success bool
	script := scripts.START_NGINX
	scriptPath := "/usr/bin/start_nginx"
	command := fmt.Sprintf("%s '%s'", scriptPath, getNginxListens(dc))
	t.AddStep(&step.PullImage{
		Image:       dc.GetContainerImage(),
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateContainer{
		Image:       dc.GetContainerImage(),
		Command:     command,
		Entrypoint:  "/bin/bash",
		Name:        getHTTPServerContainerName(curveadm, dc),
		Remove:      true,
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		Content:           &script,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{ // TODO(P1): maybe we should check all ports
		Lambda: waitNginxStarted(5),
	})

	return t, nil
}

// TASK: check network firewall
type (
	step2CheckConnectStatus struct {
		success *bool
		out     *string
		address Address
		dc      *topology.DeployConfig
	}
)

func (s *step2CheckConnectStatus) Execute(ctx *context.Context) error {
	if *s.success {
		return nil
	}

	return errno.ERR_CONNET_MOCK_SERVICE_ADDRESS_FAILED.
		F("role=%s src=%s dest=%s:%d",
			s.dc.GetRole(), s.dc.GetHost(), s.address.IP, s.address.Port)
}

func NewCheckNetworkFirewallTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	dcs := curveadm.MemStorage().Get(comm.KEY_ALL_DEPLOY_CONFIGS).([]*topology.DeployConfig)
	addresses := getServiceConnectAddress(dc, dcs)

	// add task
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Check Network Firewall <network>", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	var success bool
	for _, address := range addresses {
		t.AddStep(&step.Curl{
			Url:         fmt.Sprintf("http://%s:%d", address.IP, address.Port),
			Output:      "/dev/null",
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step2CheckConnectStatus{
			success: &success,
			out:     &out,
			dc:      dc,
			address: address,
		})
	}

	return t, nil
}

// TASK: stop container
type step2StopContainer struct {
	containerId *string
	dc          *topology.DeployConfig
	curveadm    *cli.CurveAdm
}

func (s *step2StopContainer) Execute(ctx *context.Context) error {
	if len(*s.containerId) == 0 {
		return nil
	}

	var success bool
	steps := []task.Step{}
	steps = append(steps, &step.StopContainer{
		ContainerId: *s.containerId,
		Time:        1,
		ExecOptions: s.curveadm.ExecOptions(),
	})
	steps = append(steps, &step.RemoveContainer{
		Success:     &success, // FIXME(P1): rmeove iff container exist
		ContainerId: *s.containerId,
		ExecOptions: s.curveadm.ExecOptions(),
	})

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewCleanEnvironmentTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Clean Precheck Environment", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("name=%s", getCheckPortContainerName(curveadm, dc)),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2StopContainer{
		containerId: &out,
		dc:          dc,
		curveadm:    curveadm,
	})
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("name=%s", getHTTPServerContainerName(curveadm, dc)),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2StopContainer{
		containerId: &out,
		dc:          dc,
		curveadm:    curveadm,
	})

	return t, nil
}

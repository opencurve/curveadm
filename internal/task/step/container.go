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
 * Created Date: 2021-12-13
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package step

import (
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/pkg/module"
)

type (
	DockerInfo struct {
		Success *bool
		Out     *string
		module.ExecOptions
	}

	PullImage struct {
		Image string
		Out   *string
		module.ExecOptions
	}

	Volume struct { // bind mount a volume
		HostPath      string
		ContainerPath string
	}

	CreateContainer struct {
		Image             string
		Command           string
		AddHost           []string
		Devices           []string
		Entrypoint        string
		Envs              []string
		Hostname          string
		Init              bool
		LinuxCapabilities []string
		Mount             string
		Name              string
		Network           string
		Pid               string
		Privileged        bool
		Remove            bool // automatically remove the container when it exits
		Restart           string
		SecurityOptions   []string
		Ulimits           []string
		Volumes           []Volume
		Out               *string
		module.ExecOptions
	}

	StartContainer struct {
		ContainerId *string
		Success     *bool
		Out         *string
		module.ExecOptions
	}

	StopContainer struct {
		ContainerId string
		Time        int
		Out         *string
		module.ExecOptions
	}

	RestartContainer struct {
		ContainerId string
		Out         *string
		module.ExecOptions
	}

	WaitContainer struct {
		ContainerId string
		Out         *string
		module.ExecOptions
	}

	RemoveContainer struct {
		ContainerId string
		Success     *bool
		Out         *string
		module.ExecOptions
	}

	ListContainers struct {
		Format  string
		Filter  string
		Quiet   bool // Only display numeric IDs
		ShowAll bool // Show all containers (default shows just running)
		Out     *string
		module.ExecOptions
	}

	ContainerExec struct {
		ContainerId *string
		Command     string
		Success     *bool
		Out         *string
		module.ExecOptions
	}

	CopyFromContainer struct {
		ContainerId      string
		ContainerSrcPath string
		HostDestPath     string
		Out              *string
		module.ExecOptions
	}

	CopyIntoContainer struct {
		HostSrcPath       string
		ContainerId       string
		ContainerDestPath string
		Out               *string
		module.ExecOptions
	}

	InspectContainer struct {
		ContainerId string
		Format      string
		Out         *string
		Success     *bool
		module.ExecOptions
	}

	ContainerLogs struct {
		ContainerId string
		Out         *string
		Success     *bool
		module.ExecOptions
	}

	UpdateContainer struct {
		ContainerId *string
		Restart     string
		Out         *string
		Success     *bool
		module.ExecOptions
	}
)

func (s *DockerInfo) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().DockerInfo()
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_DOCKER_INFO_FAILED)
}

func (s *PullImage) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().PullImage(s.Image)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_PULL_IMAGE_FAILED)
}

func (s *CreateContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CreateContainer(s.Image, s.Command)
	for _, host := range s.AddHost {
		cli.AddOption("--add-host %s", host)
	}
	for _, device := range s.Devices {
		cli.AddOption("--device %s", device)
	}
	if len(s.Entrypoint) > 0 {
		cli.AddOption("--entrypoint %s", s.Entrypoint)
	}
	for _, env := range s.Envs {
		cli.AddOption("--env %s", env)
	}
	if len(s.Hostname) > 0 {
		cli.AddOption("--hostname %s", s.Hostname)
	}
	if s.Init {
		cli.AddOption("--init")
	}
	for _, capability := range s.LinuxCapabilities {
		cli.AddOption("--cap-add %s", capability)
	}
	if len(s.Mount) > 0 {
		cli.AddOption("--mount %s", s.Mount)
	}
	if len(s.Name) > 0 {
		cli.AddOption("--name %s", s.Name)
	}
	if len(s.Network) > 0 {
		cli.AddOption("--network %s", s.Network)
	} else {
		cli.AddOption("--network host")
	}
	if len(s.Pid) > 0 {
		cli.AddOption("--pid %s", s.Pid)
	}
	if s.Privileged {
		cli.AddOption("--privileged")
	}
	if s.Remove {
		cli.AddOption("--rm")
	}
	if len(s.Restart) > 0 {
		cli.AddOption("--restart %s", s.Restart)
	}
	for _, security := range s.SecurityOptions {
		cli.AddOption("--security-opt %s", security)
	}
	for _, ulimit := range s.Ulimits {
		cli.AddOption("--ulimit %s", ulimit)
	}
	for _, volume := range s.Volumes {
		cli.AddOption("--volume %s:%s", volume.HostPath, volume.ContainerPath)
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_CREATE_CONTAINER_FAILED)
}

func (s *StartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StartContainer(*s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_START_CONTAINER_FAILED)
}

func (s *StopContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StopContainer(s.ContainerId)
	if s.Time > 0 {
		cli.AddOption("--time %d", s.Time)
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_STOP_CONTAINER_FAILED)
}

func (s *RestartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RestartContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_RESTART_CONTAINER_FAILED)
}

func (s *WaitContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().WaitContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_WAIT_CONTAINER_STOP_FAILED)
}

func (s *RemoveContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RemoveContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_REMOVE_CONTAINER_FAILED)
}

func (s *ListContainers) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ListContainers()
	if len(s.Format) > 0 {
		cli.AddOption("--format %s", s.Format)
	}
	if len(s.Filter) > 0 {
		cli.AddOption("--filter %s", s.Filter)
	}
	if s.Quiet {
		cli.AddOption("--quiet")
	}
	if s.ShowAll {
		cli.AddOption("--all")
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_LIST_CONTAINERS_FAILED)
}

func (s *ContainerExec) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ContainerExec(*s.ContainerId, s.Command)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_RUN_COMMAND_IN_CONTAINER_FAILED)
}

func (s *CopyFromContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyFromContainer(s.ContainerId, s.ContainerSrcPath, s.HostDestPath)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_COPY_FROM_CONTAINER_FAILED)
}

func (s *CopyIntoContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyIntoContainer(s.HostSrcPath, s.ContainerId, s.ContainerDestPath)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_COPY_INTO_CONTAINER_FAILED)
}

func (s *InspectContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().InspectContainer(s.ContainerId)
	if len(s.Format) > 0 {
		cli.AddOption("--format=%s", s.Format)
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_INSPECT_CONTAINER_FAILED)
}

func (s *ContainerLogs) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ContainerLogs(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_CONTAINER_LOGS_FAILED)
}

func (s *UpdateContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().UpdateContainer(*s.ContainerId)
	if len(s.Restart) > 0 {
		cli.AddOption("--restart %s", s.Restart)
	}
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_UPDATE_CONTAINER_FAILED)
}

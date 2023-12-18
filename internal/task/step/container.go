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
	EngineInfo struct {
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
		User              string
		Pid               string
		Publish           string
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
		Detached    bool
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

	TopContainer struct {
		ContainerId string
		Out         *string
		module.ExecOptions
	}
)

func (s *EngineInfo) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().DockerInfo()
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_CONTAINER_ENGINE_INFO_FAILED.FD("(%s info)", s.ExecWithEngine))
}

func (s *PullImage) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().PullImage(s.Image)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_PULL_IMAGE_FAILED.FD("(%s pull IMAGE)", s.ExecWithEngine))
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
	if len(s.Publish) > 0 {
		cli.AddOption("--publish %s", s.Publish)
	}
	if len(s.User) > 0 {
		cli.AddOption("--user %s", s.User)
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
	return PostHandle(nil, s.Out, out, err, errno.ERR_CREATE_CONTAINER_FAILED.FD("(%s create IMAGE)", s.ExecWithEngine))
}

func (s *StartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StartContainer(*s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_START_CONTAINER_FAILED.FD("(%s start CONTAINER)", s.ExecWithEngine))
}

func (s *StopContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StopContainer(s.ContainerId)
	if s.Time > 0 {
		cli.AddOption("--time %d", s.Time)
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_STOP_CONTAINER_FAILED.FD("(%s stop CONTAINER)", s.ExecWithEngine))
}

func (s *RestartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RestartContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_RESTART_CONTAINER_FAILED.FD("(%s restart CONTAINER)", s.ExecWithEngine))
}

func (s *WaitContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().WaitContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_WAIT_CONTAINER_STOP_FAILED.FD("(%s wait CONTAINER)", s.ExecWithEngine))
}

func (s *RemoveContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RemoveContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_REMOVE_CONTAINER_FAILED.FD("(%s rm CONTAINER)", s.ExecWithEngine))
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
	return PostHandle(nil, s.Out, out, err, errno.ERR_LIST_CONTAINERS_FAILED.FD("(%s ps)", s.ExecWithEngine))
}

func (s *ContainerExec) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ContainerExec(*s.ContainerId, s.Command)
	if s.Detached {
		cli.AddOption("--detach")
	}
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_RUN_COMMAND_IN_CONTAINER_FAILED.FD("(%s exec CONTAINER COMMAND)", s.ExecWithEngine))
}

func (s *CopyFromContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyFromContainer(s.ContainerId, s.ContainerSrcPath, s.HostDestPath)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_COPY_FROM_CONTAINER_FAILED.FD("(%s cp CONTAINER:SRC_PATH DEST_PATH)", s.ExecWithEngine))
}

func (s *CopyIntoContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyIntoContainer(s.HostSrcPath, s.ContainerId, s.ContainerDestPath)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_COPY_INTO_CONTAINER_FAILED.FD("(%s cp SRC_PATH CONTAINER:DEST_PATH)", s.ExecWithEngine))
}

func (s *InspectContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().InspectContainer(s.ContainerId)
	if len(s.Format) > 0 {
		cli.AddOption("--format=%s", s.Format)
	}

	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_INSPECT_CONTAINER_FAILED.FD("(%s inspect ID)", s.ExecWithEngine))
}

func (s *ContainerLogs) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ContainerLogs(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_CONTAINER_LOGS_FAILED.FD("(%s logs ID)", s.ExecWithEngine))
}

func (s *TopContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().TopContainer(s.ContainerId)
	out, err := cli.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_TOP_CONTAINER_FAILED.FD("(%s top ID)", s.ExecWithEngine))
}

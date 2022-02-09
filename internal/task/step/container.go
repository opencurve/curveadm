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

package step

import (
	"strings"

	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/pkg/module"
)

type (
	PullImage struct {
		Image         string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	Volume struct { // bind mount a volume
		HostPath      string
		ContainerPath string
	}

	CreateContainer struct {
		Image             string
		Command           string
		Devices           []string
		Entrypoint        string
		Envs              []string
		Hostname          string
		Init              bool
		LinuxCapabilities []string
		Mount             string
		Name              string
		Pid               string
		Privileged        bool
		Remove            bool // automatically remove the container when it exits
		Restart           string
		SecurityOptions   []string
		Ulimits           []string
		Volumes           []Volume
		Out               *string
		ExecWithSudo      bool
		ExecInLocal       bool
		ExecSudoAlias     string
	}

	StartContainer struct {
		ContainerId   *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	StopContainer struct {
		ContainerId   string
		Time          int
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	RestartContainer struct {
		ContainerId   string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	WaitContainer struct {
		ContainerId   string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	RemoveContainer struct {
		ContainerId   string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	ListContainers struct {
		Format        string
		Filter        string
		Quiet         bool // Only display numeric IDs
		ShowAll       bool // Show all containers (default shows just running)
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	ContainerExec struct {
		ContainerId   *string
		Command       string
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}

	CopyFromContainer struct {
		ContainerId      string
		ContainerSrcPath string
		HostDestPath     string
		ExecWithSudo     bool
		ExecInLocal      bool
		ExecSudoAlias    string
	}

	CopyIntoContainer struct {
		HostSrcPath       string
		ContainerId       string
		ContainerDestPath string
		ExecWithSudo      bool
		ExecInLocal       bool
		ExecSudoAlias     string
	}

	InspectContainer struct {
		ContainerId   string
		Format        string
		Out           *string
		ExecWithSudo  bool
		ExecInLocal   bool
		ExecSudoAlias string
	}
)

func (s *PullImage) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().PullImage(s.Image)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *CreateContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CreateContainer(s.Image, s.Command)
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
	cli.AddOption("--network host")

	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	*s.Out = strings.TrimSuffix(out, "\n")
	return err
}

func (s *StartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StartContainer(*s.ContainerId)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *StopContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().StopContainer(s.ContainerId)
	if s.Time > 0 {
		cli.AddOption("--time %d", s.Time)
	}
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *RestartContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RestartContainer(s.ContainerId)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *WaitContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().WaitContainer(s.ContainerId)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *RemoveContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().RemoveContainer(s.ContainerId)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
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
	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	*s.Out = strings.TrimSuffix(out, "\n")
	return err
}

func (s *InspectContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().InspectContainer(s.ContainerId)
	if len(s.Format) > 0 {
		cli.AddOption("--format = %s", s.Format)
	}
	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	*s.Out = out
	return err
}

func (s *ContainerExec) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().ContainerExec(*s.ContainerId, s.Command)
	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	if s.Out != nil {
		*s.Out = out
	}
	return err
}

func (s *CopyFromContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyFromContainer(s.ContainerId, s.ContainerSrcPath, s.HostDestPath)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

func (s *CopyIntoContainer) Execute(ctx *context.Context) error {
	cli := ctx.Module().DockerCli().CopyIntoContainer(s.HostSrcPath, s.ContainerId, s.ContainerDestPath)
	_, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	})
	return err
}

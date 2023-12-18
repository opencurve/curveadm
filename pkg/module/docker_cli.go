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
 * Created Date: 2021-12-09
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package module

import (
	"fmt"
	"strings"
	"text/template"
)

const (
	TEMPLATE_DOCKER_INFO         = "{{.engine}} info"
	TEMPLATE_PULL_IMAGE          = "{{.engine}} pull {{.options}} {{.name}}"
	TEMPLATE_CREATE_CONTAINER    = "{{.engine}} create {{.options}} {{.image}} {{.command}}"
	TEMPLATE_START_CONTAINER     = "{{.engine}} start {{.options}} {{.containers}}"
	TEMPLATE_STOP_CONTAINER      = "{{.engine}} stop {{.options}} {{.containers}}"
	TEMPLATE_RESTART_CONTAINER   = "{{.engine}} restart {{.options}} {{.containers}}"
	TEMPLATE_WAIT_CONTAINER      = "{{.engine}} wait {{.options}} {{.containers}}"
	TEMPLATE_REMOVE_CONTAINER    = "{{.engine}} rm {{.options}} {{.containers}}"
	TEMPLATE_LIST_CONTAINERS     = "{{.engine}} ps {{.options}}"
	TEMPLATE_CONTAINER_EXEC      = "{{.engine}} exec {{.options}} {{.container}} {{.command}}"
	TEMPLATE_COPY_FROM_CONTAINER = "{{.engine}} cp {{.options}} {{.container}}:{{.srcPath}} {{.destPath}}"
	TEMPLATE_COPY_INTO_CONTAINER = "{{.engine}} cp {{.options}}  {{.srcPath}} {{.container}}:{{.destPath}}"
	TEMPLATE_INSPECT_CONTAINER   = "{{.engine}} inspect {{.options}} {{.container}}"
	TEMPLATE_CONTAINER_LOGS      = "{{.engine}} logs {{.options}} {{.container}}"
	TEMPLATE_UPDATE_CONTAINER    = "{{.engine}} update {{.options}} {{.container}}"
	TEMPLATE_TOP_CONTAINER       = "{{.engine}} top {{.container}}"
)

type DockerCli struct {
	sshClient *SSHClient
	options   []string
	tmpl      *template.Template
	data      map[string]interface{}
}

func NewDockerCli(sshClient *SSHClient) *DockerCli {
	return &DockerCli{
		sshClient: sshClient,
		options:   []string{},
		tmpl:      nil,
		data:      map[string]interface{}{},
	}
}

func (s *DockerCli) AddOption(format string, args ...interface{}) *DockerCli {
	s.options = append(s.options, fmt.Sprintf(format, args...))
	return s
}

func (cli *DockerCli) Execute(options ExecOptions) (string, error) {
	cli.data["options"] = strings.Join(cli.options, " ")
	cli.data["engine"] = options.ExecWithEngine
	return execCommand(cli.sshClient, cli.tmpl, cli.data, options)
}

func (cli *DockerCli) DockerInfo() *DockerCli {
	cli.tmpl = template.Must(template.New("DockerInfo").Parse(TEMPLATE_DOCKER_INFO))
	return cli
}

func (cli *DockerCli) PullImage(image string) *DockerCli {
	cli.tmpl = template.Must(template.New("PullImage").Parse(TEMPLATE_PULL_IMAGE))
	cli.data["name"] = image
	return cli
}

func (cli *DockerCli) CreateContainer(image, command string) *DockerCli {
	cli.tmpl = template.Must(template.New("CreateContainer").Parse(TEMPLATE_CREATE_CONTAINER))
	cli.data["image"] = image
	cli.data["command"] = command
	return cli
}

func (cli *DockerCli) StartContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("StartContainer").Parse(TEMPLATE_START_CONTAINER))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) StopContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("StopContainer").Parse(TEMPLATE_STOP_CONTAINER))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) RestartContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("RestartContainer").Parse(TEMPLATE_RESTART_CONTAINER))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) WaitContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("WaitContainer").Parse(TEMPLATE_WAIT_CONTAINER))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) RemoveContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("RemoveContainer").Parse(TEMPLATE_REMOVE_CONTAINER))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) ListContainers() *DockerCli {
	cli.tmpl = template.Must(template.New("ListContainers").Parse(TEMPLATE_LIST_CONTAINERS))
	return cli
}

func (cli *DockerCli) ContainerExec(containerId, command string) *DockerCli {
	cli.tmpl = template.Must(template.New("ContainerExec").Parse(TEMPLATE_CONTAINER_EXEC))
	cli.data["container"] = containerId
	cli.data["command"] = command
	return cli
}

func (cli *DockerCli) CopyFromContainer(containerId, srcPath, destPath string) *DockerCli {
	cli.tmpl = template.Must(template.New("CopyFromContainer").Parse(TEMPLATE_COPY_FROM_CONTAINER))
	cli.data["container"] = containerId
	cli.data["srcPath"] = srcPath
	cli.data["destPath"] = destPath
	return cli
}

func (cli *DockerCli) CopyIntoContainer(srcPath, containerId, destPath string) *DockerCli {
	cli.tmpl = template.Must(template.New("CopyIntoContainer").Parse(TEMPLATE_COPY_INTO_CONTAINER))
	cli.data["srcPath"] = srcPath
	cli.data["container"] = containerId
	cli.data["destPath"] = destPath
	return cli
}

func (cli *DockerCli) InspectContainer(containerId string) *DockerCli {
	cli.tmpl = template.Must(template.New("InspectContainer").Parse(TEMPLATE_INSPECT_CONTAINER))
	cli.data["container"] = containerId
	return cli
}

func (cli *DockerCli) ContainerLogs(containerId string) *DockerCli {
	cli.tmpl = template.Must(template.New("ContainerLogs").Parse(TEMPLATE_CONTAINER_LOGS))
	cli.data["container"] = containerId
	return cli
}

func (cli *DockerCli) TopContainer(containerId string) *DockerCli {
	cli.tmpl = template.Must(template.New("TopContainer").Parse(TEMPLATE_TOP_CONTAINER))
	cli.data["container"] = containerId
	return cli
}

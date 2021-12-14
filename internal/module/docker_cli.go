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

package module

import (
	"fmt"
	"strings"
	"text/template"

	ssh "github.com/melbahja/goph"
)

// docker pull [OPTIONS] NAME[:TAG|@DIGEST]
// docker create [OPTIONS] IMAGE [COMMAND] [ARG...]
// docker start [OPTIONS] CONTAINER [CONTAINER...]
// docker stop [OPTIONS] CONTAINER [CONTAINER...]
// docker restart [OPTIONS] CONTAINER [CONTAINER...]
// docker rm [OPTIONS] CONTAINER [CONTAINER...]
// docker exec [OPTIONS] CONTAINER COMMAND [ARG...]
// docker cp [OPTIONS] CONTAINER:SRC_PATH DEST_PATH|-
const (
	TEMPLATE_PULL   = "docker pull {{.options}} {{.name}}"
	TEMPLATE_CREATE = "docker create {{.options}} {{.image}} {{.command}}"
	TEMPLATE_START  = "docker start {{.options}} {{.containers}}"
	TEMPLATE_EXEC   = "docker exec {{.options}} {{.container}} {{.command}}"
)

type DockerCli struct {
	sshClient *ssh.Client
	options   []string
	tmpl      *template.Template
	data      map[string]interface{}
}

func NewDockerCli(sshClient *ssh.Client) *DockerCli {
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

func (cli *DockerCli) PullImage(image string) *DockerCli {
	cli.tmpl = template.Must(template.New("PullImage").Parse(TEMPLATE_PULL))
	cli.data["name"] = image
	return cli
}

func (cli *DockerCli) CreateContainer(image, command string) *DockerCli {
	cli.tmpl = template.Must(template.New("CreateContainer").Parse(TEMPLATE_CREATE))
	cli.data["image"] = image
	cli.data["command"] = command
	return cli
}

func (cli *DockerCli) StartContainer(containerId ...string) *DockerCli {
	cli.tmpl = template.Must(template.New("StartContainer").Parse(TEMPLATE_START))
	cli.data["containers"] = strings.Join(containerId, " ")
	return cli
}

func (cli *DockerCli) ContainerExec(containerId, command string) *DockerCli {
	cli.tmpl = template.Must(template.New("ContainerExec").Parse(TEMPLATE_EXEC))
	cli.data["container"] = containerId
	cli.data["command"] = command
	return cli
}

func (cli *DockerCli) Execute(options ExecOption) (string, error) {
	cli.data["options"] = strings.Join(cli.options, " ")
	return execCommand(cli.sshClient, cli.tmpl, cli.data, options)
}

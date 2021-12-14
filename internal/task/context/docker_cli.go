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

package context

import (
	"fmt"

	"github.com/melbahja/goph"
)

/*
 * docker pull [OPTIONS] NAME[:TAG|@DIGEST]
 * docker create [OPTIONS] IMAGE [COMMAND] [ARG...]
 * docker start [OPTIONS] CONTAINER [CONTAINER...]
 * docker stop [OPTIONS] CONTAINER [CONTAINER...]
 * docker restart [OPTIONS] CONTAINER [CONTAINER...]
 * docker rm [OPTIONS] CONTAINER [CONTAINER...]
 * docker exec [OPTIONS] CONTAINER COMMAND [ARG...]
 * docker cp [OPTIONS] CONTAINER:SRC_PATH DEST_PATH|-
 */

type ExecOption struct {
	Sudo bool
}

type DockerCli struct {
	sshClient   *goph.Client
	image       string
	containerId string
	command     string
	srcPath     string
	destPath    string
	options     []string
}

func NewDockerCli(sshClient *goph.Client) *DockerCli {
	return &DockerCli{sshClient: sshClient}
}

func (cli *DockerCli) PullImage(image string) *DockerCli {
	return &DockerCli{image: image}
}

func (cli *DockerCli) CreateContainer(image, command string) *DockerCli {
	return &DockerCli{image: image}
}

func (cli *DockerCli) ListContainers() *DockerCli {
	return &DockerCli{}
}

func (cli *DockerCli) StartContainer(containerId string) *DockerCli {

}

func (cli *DockerCli) RemoveContainer(containerId string) *DockerCli {

}

func (cli *DockerCli) RestartContainer(containerId string) *DockerCli {

}

func (cli *DockerCli) StopContainer(containerId string) *DockerCli {

}

func (cli *DockerCli) CopyIntoContainer(containerId string) *DockerCli {

}

func (cli *DockerCli) Exec(containerId, command string) *DockerCli {

}

func (cli *DockerCli) AddOption(format string, args ...interface{}) *DockerCli {
	cli.options = append(cli.options, fmt.Sprintf(format, args...))
	return cli
}

func (cli *DockerCli) Execute(options ExecOption) (string, error) {
}

func (cli *DockerCli) ExecuteWithSudo() (string, error) {
}

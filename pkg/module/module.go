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

package module

import (
	"bytes"
	"fmt"
	"os/exec"
	"text/template"

	ssh "github.com/melbahja/goph"
	"github.com/opencurve/curveadm/pkg/log"
)

type ExecOption struct {
	ExecWithSudo bool
	ExecInLocal  bool
}

type Module struct {
	sshClient *ssh.Client
}

func NewModule(sshClient *ssh.Client) *Module {
	return &Module{sshClient: sshClient}
}

func (m *Module) Shell() *Shell {
	return NewShell(m.sshClient)
}

func (m *Module) File() *FileManager {
	return NewFileManager(m.sshClient)
}

func (m *Module) DockerCli() *DockerCli {
	return NewDockerCli(m.sshClient)
}

// common utils
func remoteAddr(client *ssh.Client) string {
	if client == nil {
		return "-"
	}

	config := client.Config
	return fmt.Sprintf("%s@%s:%d", config.User, config.Addr, config.Port)
}

func execCommand(sshClient *ssh.Client,
	tmpl *template.Template,
	data map[string]interface{},
	options ExecOption) (string, error) {
	buffer := bytes.NewBufferString("")
	err := tmpl.Execute(buffer, data)
	if err != nil {
		return "", err
	}

	cmd := buffer.String()
	if options.ExecWithSudo {
		cmd = "sudo " + cmd
	}

	var out []byte
	if options.ExecInLocal {
		out, err = exec.Command("bash", "-c", cmd).CombinedOutput()
	} else {
		out, err = sshClient.Run(cmd)
	}

	log.SwitchLevel(err)("execCommand",
		log.Field("remoteAddr", remoteAddr(sshClient)),
		log.Field("command", cmd),
		log.Field("error", err),
		log.Field("output", string(out)))
	return string(out), err
}

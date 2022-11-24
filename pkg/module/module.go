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

package module

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"text/template"
	"time"

	"github.com/melbahja/goph"
	log "github.com/opencurve/curveadm/pkg/log/glg"
)

type (
	Module struct {
		sshClient *SSHClient
	}

	ExecOptions struct {
		ExecWithSudo   bool
		ExecInLocal    bool
		ExecSudoAlias  string
		ExecTimeoutSec int
	}

	TimeoutError struct {
		timeout int
	}
)

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("execute command timed out (timeout: %d seconds)",
		e.timeout)
}

func NewModule(sshClient *SSHClient) *Module {
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
func remoteAddr(client *SSHClient) string {
	if client == nil {
		return "-"
	}

	config := client.Config()
	return fmt.Sprintf("%s@%s:%d", config.User, config.Host, config.Port)
}

func execCommand(sshClient *SSHClient,
	tmpl *template.Template,
	data map[string]interface{},
	options ExecOptions) (string, error) {
	// (1) rendering command template
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, data); err != nil {
		return "", err
	}

	// (2) handle 'sudo_alias'
	command := buffer.String()
	if options.ExecWithSudo {
		sudo := "sudo"
		if len(options.ExecSudoAlias) > 0 {
			sudo = options.ExecSudoAlias
		}
		command = strings.Join([]string{sudo, command}, " ")
	}
	command = strings.TrimLeft(command, " ")

	// (3) handle 'become_user'
	if sshClient != nil {
		becomeMethod := sshClient.Config().BecomeMethod
		becomeFlags := sshClient.Config().BecomeFlags
		becomeUser := sshClient.Config().BecomeUser
		if len(becomeUser) > 0 && !options.ExecInLocal {
			become := strings.Join([]string{becomeMethod, becomeFlags, becomeUser}, " ")
			command = strings.Join([]string{become, command}, " ")
		}
	}

	// (4) create context for timeout
	ctx := context.Background()
	if options.ExecTimeoutSec > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(options.ExecTimeoutSec)*time.Second)
		defer cancel()
	}

	// (5) execute command
	var out []byte
	var err error
	if options.ExecInLocal {
		cmd := exec.CommandContext(ctx, "bash", "-c", command)
		cmd.Env = []string{"LANG=en_US.UTF-8"}
		out, err = cmd.CombinedOutput()
	} else {
		var cmd *goph.Cmd
		cmd, err = sshClient.Client().CommandContext(ctx, command)
		if err == nil {
			cmd.Env = []string{"LANG=en_US.UTF-8"}
			out, err = cmd.CombinedOutput()
		}
	}

	if ctx.Err() == context.DeadlineExceeded {
		err = &TimeoutError{options.ExecTimeoutSec}
	}

	log.SwitchLevel(err)("Execute command",
		log.Field("remoteAddr", remoteAddr(sshClient)),
		log.Field("command", command),
		log.Field("output", strings.TrimSuffix(string(out), "\n")),
		log.Field("error", err))
	return string(out), err
}

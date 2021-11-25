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

package task

import (
	"fmt"
	"os/exec"

	ssh "github.com/melbahja/goph"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/scripts"
)

type Module struct {
	sshClient *ssh.Client
}

func NewModule(sshClient *ssh.Client) *Module {
	return &Module{sshClient: sshClient}
}

func (m *Module) Clean() {
	if m.sshClient != nil {
		m.sshClient.Close()
	}
}

func (m *Module) RemoteAddr() string {
	if m.sshClient == nil {
		return ""
	}

	user := m.sshClient.Config.User
	host := m.sshClient.Config.Addr
	port := m.sshClient.Config.Port
	return fmt.Sprintf("%s@%s:%d", user, host, port)
}

func (m *Module) Scp(localPath, remotePath string) error {
	if m.sshClient == nil {
		return fmt.Errorf("remote unreached")
	}

	err := m.sshClient.Upload(localPath, remotePath)

	log.SwitchLevel(err)("Scp",
		log.Field("remoteAddr", m.RemoteAddr()),
		log.Field("localPath", localPath),
		log.Field("remotePath", remotePath),
		log.Field("error", err))

	return err
}

func (m *Module) Download(remotePath, localPath string) error {
	if m.sshClient == nil {
		return fmt.Errorf("remote unreached")
	}

	err := m.sshClient.Download(remotePath, localPath)

	log.SwitchLevel(err)("Download",
		log.Field("remoteAddr", m.RemoteAddr()),
		log.Field("remotePath", remotePath),
		log.Field("localPath", localPath),
		log.Field("error", err))

	return err
}

func (m *Module) SshShell(format string, a ...interface{}) (string, error) {
	if m.sshClient == nil {
		return "", fmt.Errorf("remote unreached")
	}

	cmd := fmt.Sprintf(format, a...)
	out, err := m.sshClient.Run(cmd)

	log.SwitchLevel(err)("SshShell",
		log.Field("remoteAddr", m.RemoteAddr()),
		log.Field("cmd", cmd),
		log.Field("output", out),
		log.Field("error", err))

	return string(out), err
}

func (c *Module) LocalShell(format string, a ...interface{}) (string, error) {
	cmd := fmt.Sprintf(format, a...)
	out, err := exec.Command("bash", "-c", cmd).CombinedOutput()

	log.SwitchLevel(err)("LocalShell",
		log.Field("cmd", cmd),
		log.Field("output", out),
		log.Field("error", err))

	return string(out), err
}

func (c *Module) SshMountScript(name, remotePath string) error {
	script, ok := scripts.Get(name)
	if !ok {
		return fmt.Errorf("script '%s' not found", name)
	}

	_, err := c.SshShell("echo '%s' > %s", script, remotePath)
	log.SwitchLevel(err)("SshMountScript",
		log.Field("script", name),
		log.Field("remotePath", remotePath),
		log.Field("error", err))

	return err
}

func (m *Module) SshCreateDir(dir string) error {
	if dir == "" {
		return nil
	}
	_, err := m.SshShell("mkdir -p %s", dir)
	return err
}

func (m *Module) SshRemoveDir(dir string, root bool) error {
	if dir == "" {
		return nil
	}

	sudo := ""
	if root {
		sudo = "sudo"
	}
	_, err := m.SshShell("%s rm -rf %s", sudo, dir)
	return err
}

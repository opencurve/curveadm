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
 * Created Date: 2022-08-02
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package tools

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"text/template"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	TEMPLATE_SCP                             = `scp -P {{.port}} {{or .options ""}} {{.source}} {{.user}}@{{.host}}:{{.target}}`
	TEMPLATE_SSH_COMMAND                     = `ssh {{.user}}@{{.host}} -p {{.port}} {{or .options ""}} {{or .become ""}} {{.command}}`
	TEMPLATE_SSH_ATTACH                      = `ssh -tt {{.user}}@{{.host}} -p {{.port}} {{or .options ""}} {{or .become ""}} {{.command}}`
	TEMPLATE_COMMAND_EXEC_CONTAINER          = `{{.sudo}} docker exec -it {{.container_id}} /bin/bash -c "cd {{.home_dir}}; /bin/bash"`
	TEMPLATE_LOCAL_EXEC_CONTAINER            = `docker exec -it {{.container_id}} /bin/bash` // FIXME: merge it
	TEMPLATE_COMMAND_EXEC_CONTAINER_NOATTACH = `{{.sudo}} docker exec -t {{.container_id}} /bin/bash -c "{{.command}}"`
)

func prepareOptions(curveadm *cli.CurveAdm, host string, become bool, extra map[string]interface{}) (map[string]interface{}, error) {
	options := map[string]interface{}{}
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	config := hc.GetSSHConfig()
	options["user"] = config.User
	options["host"] = config.Host
	options["port"] = config.Port

	opts := []string{
		"-o StrictHostKeyChecking=no",
		//"-o UserKnownHostsFile=/dev/null",
	}
	if !config.ForwardAgent {
		opts = append(opts, fmt.Sprintf("-i %s", config.PrivateKeyPath))
	}
	if len(config.BecomeUser) > 0 && become {
		options["become"] = fmt.Sprintf("%s %s %s",
			config.BecomeMethod, config.BecomeFlags, config.BecomeUser)
	}

	for k, v := range extra {
		options[k] = v
	}

	options["options"] = strings.Join(opts, " ")
	return options, nil
}

func newCommand(curveadm *cli.CurveAdm, text string, options map[string]interface{}) (*exec.Cmd, error) {
	tmpl := template.Must(template.New(utils.MD5Sum(text)).Parse(text))
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, options); err != nil {
		return nil, errno.ERR_BUILD_TEMPLATE_FAILED.E(err)
	}
	command := buffer.String()
	items := strings.Split(command, " ")
	return exec.Command(items[0], items[1:]...), nil
}

func runCommand(curveadm *cli.CurveAdm, text string, options map[string]interface{}) error {
	cmd, err := newCommand(curveadm, text, options)
	if err != nil {
		return err
	}
	cmd.Stdout = curveadm.Out()
	cmd.Stderr = curveadm.Err()
	cmd.Stdin = curveadm.In()
	return cmd.Run()
}

func runCommandOutput(curveadm *cli.CurveAdm, text string, options map[string]interface{}) (string, error) {
	cmd, err := newCommand(curveadm, text, options)
	if err != nil {
		return "", err
	}
	out, err := cmd.CombinedOutput()
	return string(out), err
}

func ssh(curveadm *cli.CurveAdm, options map[string]interface{}) error {
	err := runCommand(curveadm, TEMPLATE_SSH_ATTACH, options)
	if err != nil && !strings.HasPrefix(err.Error(), "exit status") {
		return errno.ERR_CONNECT_REMOTE_HOST_WITH_INTERACT_BY_SSH_FAILED.E(err)
	}
	return nil
}

func scp(curveadm *cli.CurveAdm, options map[string]interface{}) error {
	// TODO: added error code
	_, err := runCommandOutput(curveadm, TEMPLATE_SCP, options)
	return err
}

func execute(curveadm *cli.CurveAdm, options map[string]interface{}) (string, error) {
	return runCommandOutput(curveadm, TEMPLATE_SSH_COMMAND, options)
}

func AttachRemoteHost(curveadm *cli.CurveAdm, host string, become bool) error {
	options, err := prepareOptions(curveadm, host, become,
		map[string]interface{}{"command": "/bin/bash"})
	if err != nil {
		return err
	}
	return ssh(curveadm, options)
}

func AttachRemoteContainer(curveadm *cli.CurveAdm, host, containerId, home string) error {
	data := map[string]interface{}{
		"sudo":         curveadm.Config().GetSudoAlias(),
		"container_id": containerId,
		"home_dir":     home,
	}
	tmpl := template.Must(template.New("command").Parse(TEMPLATE_COMMAND_EXEC_CONTAINER))
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, data); err != nil {
		return errno.ERR_BUILD_TEMPLATE_FAILED.E(err)
	}
	command := buffer.String()

	options, err := prepareOptions(curveadm, host, true,
		map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return ssh(curveadm, options)
}

func AttachLocalContainer(curveadm *cli.CurveAdm, containerId string) error {
	data := map[string]interface{}{
		"container_id": containerId,
	}
	tmpl := template.Must(template.New("command").Parse(TEMPLATE_LOCAL_EXEC_CONTAINER))
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, data); err != nil {
		return errno.ERR_BUILD_TEMPLATE_FAILED.E(err)
	}
	command := buffer.String()
	return runCommand(curveadm, command, map[string]interface{}{})
}

func ExecCmdInRemoteContainer(curveadm *cli.CurveAdm, host, containerId, cmd string) error {
	data := map[string]interface{}{
		"sudo":         curveadm.Config().GetSudoAlias(),
		"container_id": containerId,
		"command":      cmd,
	}
	tmpl := template.Must(template.New("command").Parse(TEMPLATE_COMMAND_EXEC_CONTAINER_NOATTACH))
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, data); err != nil {
		return errno.ERR_BUILD_TEMPLATE_FAILED.E(err)
	}
	command := buffer.String()

	options, err := prepareOptions(curveadm, host, true,
		map[string]interface{}{"command": command})
	if err != nil {
		return err
	}
	return ssh(curveadm, options)
}

func Scp(curveadm *cli.CurveAdm, host, source, target string) error {
	options, err := prepareOptions(curveadm, host, false,
		map[string]interface{}{
			"source": source,
			"target": target,
		})
	if err != nil {
		return err
	}
	return scp(curveadm, options)
}

func ExecuteRemoteCommand(curveadm *cli.CurveAdm, host, command string) (string, error) {
	options, err := prepareOptions(curveadm, host, true,
		map[string]interface{}{"command": command})
	if err != nil {
		return "", err
	}
	return execute(curveadm, options)
}

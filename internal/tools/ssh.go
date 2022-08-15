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
	"strings"
	"text/template"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	TEMPLATE_SSH_ATTACH             = `ssh -tt {{.user}}@{{.host}} -p {{.port}} {{or .options ""}} {{or .become ""}} {{.command}}`
	TEMPLATE_COMMAND_EXEC_CONTAINER = `{{.sudo}} docker exec -it {{.container_id}} /bin/bash -c "cd {{.home_dir}}; /bin/bash"`
)

func prepareOptions(curveadm *cli.CurveAdm, host, command string, become bool) (map[string]interface{}, error) {
	options := map[string]interface{}{}
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	config := hc.GetSSHConfig()
	options["user"] = config.User
	options["host"] = config.Host
	options["port"] = config.Port
	if !config.ForwardAgent {
		options["options"] = fmt.Sprintf("-i %s", config.PrivateKeyPath)
	}
	if len(config.BecomeUser) > 0 && become {
		options["become"] = fmt.Sprintf("%s %s %s",
			config.BecomeMethod, config.BecomeFlags, config.BecomeUser)
	}
	options["command"] = command
	return options, nil
}

func sshAttach(curveadm *cli.CurveAdm, options map[string]interface{}) error {
	tmpl := template.Must(template.New("ssh_attach").Parse(TEMPLATE_SSH_ATTACH))
	buffer := bytes.NewBufferString("")
	if err := tmpl.Execute(buffer, options); err != nil {
		return errno.ERR_BUILD_TEMPLATE_FAILED.E(err)
	}
	command := buffer.String()
	build.DEBUG(build.DEBUG_TOOL, build.Field{"command", command})

	cmd := utils.NewCommand(command)
	cmd.Stdout = curveadm.Out()
	cmd.Stderr = curveadm.Err()
	cmd.Stdin = curveadm.In()
	err := cmd.Run()
	if err != nil && !strings.HasPrefix(err.Error(), "exit status") {
		return errno.ERR_CONNECT_REMOTE_HOST_WITH_INTERACT_BY_SSH_FAILED.E(err)
	}
	return nil
}

func AttachRemoteHost(curveadm *cli.CurveAdm, host string, become bool) error {
	options, err := prepareOptions(curveadm, host, "/bin/bash", become)
	if err != nil {
		return err
	}
	return sshAttach(curveadm, options)
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

	options, err := prepareOptions(curveadm, host, command, true)
	if err != nil {
		return err
	}
	return sshAttach(curveadm, options)
}

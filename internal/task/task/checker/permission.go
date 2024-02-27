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
 * Created Date: 2022-07-15
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package checker

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	SIGNATURE_USER_NOT_FOUND               = "unknown user"
	SIGNATURE_PERMISSION_DENIED            = "permission denied"
	SIGNATURE_PERMISSION_WITH_PASSWORD     = "respect the privacy of others"
	SIGNATURE_COMMAND_NOT_FOUND            = "not found"
	SIGNATURE_DOCKER_DEAMON_IS_NOT_RUNNING = "is the docker daemon running"
)

func checkUser(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}

		if strings.Contains(*out, SIGNATURE_USER_NOT_FOUND) {
			return errno.ERR_USER_NOT_FOUND.S(*out)
		}
		return errno.ERR_UNKNOWN.S(*out)
	}
}

func checkHostname(hostname *string, success *bool) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_HOSTNAME_NOT_RESOLVED.
			F("hostname=%s", *hostname)
	}
}

func checkCreateDirectory(dc *topology.DeployConfig, path string, success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}

		*out = strings.ToLower(*out)
		if strings.Contains(*out, SIGNATURE_PERMISSION_DENIED) {
			return errno.ERR_CREATE_DIRECOTRY_PERMISSION_DENIED.
				F("host=%s, role=%s, directory=%s\n%s",
					dc.GetHost(), dc.GetRole(), path, *out)
		} else if strings.Contains(*out, SIGNATURE_PERMISSION_WITH_PASSWORD) {
			return errno.ERR_CREATE_DIRECOTRY_PERMISSION_DENIED.
				F("host=%s, role=%s, directory=%s (need password)",
					dc.GetHost(), dc.GetRole(), path)
		}
		return nil
	}
}

func CheckEngineInfo(host, engine string, success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}

		*out = strings.ToLower(*out)
		if strings.Contains(*out, SIGNATURE_COMMAND_NOT_FOUND) {
			return errno.ERR_CONTAINER_ENGINE_NOT_INSTALLED.
				F("host=%s\n%s", host, *out)
		} else if strings.Contains(*out, SIGNATURE_PERMISSION_DENIED) {
			return errno.ERR_EXECUTE_CONTAINER_ENGINE_COMMAND_PERMISSION_DENIED.
				F("host=%s\n%s", host, *out)
		} else if strings.Contains(*out, SIGNATURE_PERMISSION_WITH_PASSWORD) {
			return errno.ERR_EXECUTE_CONTAINER_ENGINE_COMMAND_PERMISSION_DENIED.
				F("host=%s (need password)", host)
		} else if strings.Contains(*out, SIGNATURE_DOCKER_DEAMON_IS_NOT_RUNNING) {
			return errno.ERR_DOCKER_DAEMON_IS_NOT_RUNNING.
				F("host=%s\n%s", host, *out)
		}
		return errno.ERR_UNKNOWN.S(*out)
	}
}

func NewCheckPermissionTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s", dc.GetHost(), dc.GetRole())
	t := task.NewTask("Check Permission <permission>", subname, hc.GetConnectConfig())

	// add step to task
	var out, hostname string
	var success bool
	dirs := getServiceDirectorys(dc)
	options := curveadm.ExecOptions()
	options.ExecWithSudo = false // the directory belong user

	// (1) check `become_user`
	t.AddStep(&step.Whoami{
		Success:     &success,
		Out:         &out,
		ExecOptions: options,
	})
	t.AddStep(&step.Lambda{
		Lambda: checkUser(&success, &out),
	})
	// (2) check whether hostname resolved
	t.AddStep(&step.Hostname{
		Out:         &hostname,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Ping{
		Destination: &hostname,
		Count:       1,
		Timeout:     1,
		Success:     &success,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkHostname(&hostname, &success),
	})
	// (3) check create directory permission
	for _, dir := range dirs {
		t.AddStep(&step.CreateDirectory{
			Paths:       []string{dir.Path},
			Success:     &success,
			Out:         &out,
			ExecOptions: options,
		})
		t.AddStep(&step.Lambda{
			Lambda: checkCreateDirectory(dc, dir.Path, &success, &out),
		})
	}
	// (4) check docker/podman engine command {exist, permission, running}
	t.AddStep(&step.EngineInfo{
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: CheckEngineInfo(dc.GetHost(), curveadm.ExecOptions().ExecWithEngine, &success, &out),
	})

	return t, nil
}

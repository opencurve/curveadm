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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	PERMISSIONS_600 = 384 // -rw------- (256 + 128 = 384)
)

func doNothing() step.LambdaType {
	return func(ctx *context.Context) error {
		return nil
	}
}

// we should check host again for maybe someone change the priavte key file
func checkHost(hc *hosts.HostConfig) step.LambdaType {
	return func(ctx *context.Context) error {
		privateKeyFile := hc.GetPrivateKeyFile()
		if !hc.GetForwardAgent() {
			if !utils.PathExist(privateKeyFile) {
				return errno.ERR_PRIVATE_KEY_FILE_NOT_EXIST.
					F("%s: no such file", privateKeyFile)
			} else if utils.GetFilePermissions(privateKeyFile) != PERMISSIONS_600 {
				return errno.ERR_PRIVATE_KEY_FILE_REQUIRE_600_PERMISSIONS.
					F("file=%s mode=%d", privateKeyFile, utils.GetFilePermissions(privateKeyFile))
			}
		}
		return nil
	}
}

func NewCheckSSHConnectTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	method := utils.Choose(hc.GetForwardAgent(), "forwardAgent", "privateKey")
	subname := fmt.Sprintf("host=%s method=%s", dc.GetHost(), method)
	t := task.NewTask("Check SSH Connect <ssh>", subname, hc.GetSSHConfig())

	// add step to task
	t.AddStep(&step.Lambda{
		Lambda: checkHost(hc),
	})
	t.AddStep(&step.Lambda{
		Lambda: doNothing(),
	})

	return t, nil
}

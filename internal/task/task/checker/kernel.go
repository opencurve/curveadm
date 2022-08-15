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
 * Created Date: 2022-07-14
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package checker

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	CHUNKSERVER_LEAST_KERNEL_VERSION = "3.15.0"
	REGEX_KERNEL_VAERSION            = "^(\\d+\\.\\d+\\.\\d+)-.+$"
)

func calcKernelVersion(version string) int {
	var num int
	items := strings.Split(version, ".")
	for _, item := range items {
		n, _ := strconv.Atoi(item)
		num = num*1000 + n
	}
	return num
}

func checkKernelVersion(out *string, dc *topology.DeployConfig) step.LambdaType {
	return func(ctx *context.Context) error {
		if !dc.GetEnableRenameAt2() {
			return nil
		}

		regex, err := regexp.Compile(REGEX_KERNEL_VAERSION)
		if err != nil {
			return errno.ERR_UNRECOGNIZED_KERNEL_VERSION.
				F("kernel version: %s", *out)
		}

		mu := regex.FindStringSubmatch(*out)
		if len(mu) == 0 {
			return errno.ERR_UNRECOGNIZED_KERNEL_VERSION.
				F("kernel version: %s", *out)
		}

		current := mu[1]
		if calcKernelVersion(current) < calcKernelVersion(CHUNKSERVER_LEAST_KERNEL_VERSION) {
			return errno.ERR_RENAMEAT_NOT_SUPPORTED_IN_CURRENT_KERNEL.
				F("kernel version: %s", *out)
		}
		return nil
	}
}

func checkKernelModule(name string, success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success == true {
			return nil
		}

		if name == comm.KERNERL_MODULE_NBD {
			return errno.ERR_KERNEL_NBD_MODULE_NOT_LOADED.S(*out)
		} else if name == comm.KERNERL_MODULE_FUSE {
			return errno.ERR_KERNEL_FUSE_MODULE_NOT_LOADED.S(*out)
		}
		return nil
	}
}

func NewCheckKernelVersionTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s role=%s require=(>=%s)",
		dc.GetHost(), dc.GetRole(), CHUNKSERVER_LEAST_KERNEL_VERSION)
	t := task.NewTask("Check Kernel Version <kernel>", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	t.AddStep(&step.UnixName{
		KernelRelease: true,
		Out:           &out,
		ExecOptions:   curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkKernelVersion(&out, dc),
	})

	return t, nil
}

func NewCheckKernelModuleTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	host := curveadm.MemStorage().Get(comm.KEY_CLIENT_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	name := curveadm.MemStorage().Get(comm.KEY_CHECK_KERNEL_MODULE_NAME).(string)
	subname := fmt.Sprintf("host=%s module=%s", host, name)
	t := task.NewTask("Check Kernel Module", subname, hc.GetSSHConfig())

	// add step to task
	var out string
	var success bool
	t.AddStep(&step.ModInfo{
		Name:        name,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkKernelModule(name, &success, &out),
	})
	return t, nil
}

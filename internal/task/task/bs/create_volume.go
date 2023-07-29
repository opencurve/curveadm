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
 * Created Date: 2022-07-31
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	FORMAT_TOOLS_CONF = `mdsAddr=%s
rootUserName=root
rootUserPassword=root_password
`
)

func checkVolumeStatus(out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*out) == 0 {
			return errno.ERR_VOLUME_CONTAINER_LOSED
		} else if !strings.HasPrefix(*out, "Up") {
			return errno.ERR_VOLUME_CONTAINER_ABNORMAL.
				F("status: %s", *out)
		}
		return nil
	}
}

func checkCreateStatus(out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *out == "SUCCESS" {
			return nil
		} else if *out == "EXIST" {
			return task.ERR_SKIP_TASK
		}
		return errno.ERR_CREATE_VOLUME_FAILED
	}
}

func runCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("command execution error: %w, stderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}

func checkDiskSizeStatus(out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		// Execute lsblk | grep nbd command
		output, err := runCommand("sh", "-c", "lsblk | grep nbd")
		if err != nil {
			return err
		}

		// Now 'output' contains the output of the 'lsblk | grep nbd' command
		// Parse the output to get the disk size information and store it in 'out' (assuming 'out' is a pointer to a string)
		*out = output

		return nil
	}
}

func NewCreateVolumeTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_MAP_OPTIONS).(MapOptions)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("hostname=%s image=%s", hc.GetHostname(), cc.GetContainerImage())
	t := task.NewTask("Create Volume", subname, hc.GetSSHConfig())

	// add step
	var out string
	containerName := volume2ContainerName(options.User, options.Volume)
	containerId := containerName
	toolsConf := fmt.Sprintf(FORMAT_TOOLS_CONF, cc.GetClusterMDSAddr())
	script := scripts.CREATE_VOLUME
	scriptPath := "/curvebs/nebd/sbin/create.sh"
	command := fmt.Sprintf("/bin/bash %s %s %s %d", scriptPath, options.User, options.Volume, options.Size)

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Quiet:       true,
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkVolumeStatus(&out),
	})
	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerName,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install create_volume.sh
		Content:           &script,
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     command,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkCreateStatus(&out),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkDiskSizeStatus(&out),
	})

	return t, nil
}

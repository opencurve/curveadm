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

package command

import (
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tasks/client"
	"github.com/opencurve/curveadm/internal/tasks/task"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type checkOptions struct {
	mountPoint string
}

func NewCheckCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options checkOptions

	cmd := &cobra.Command{
		Use:   "check MOUNT_POINT",
		Short: "Check mount status",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.mountPoint = args[0]
			return runCheck(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runCheck(curveadm *cli.CurveAdm, options checkOptions) error {
	mountPoint := strings.TrimSuffix(options.mountPoint, "/")

	if t, err := client.NewGetMountStatusTask(curveadm, mountPoint); err != nil {
		return err
	} else if err := task.ParallelExecute(1, []*task.Task{t}, task.Options{SilentSubBar: true}); err != nil {
		return err
	}

	v := curveadm.MemStorage().Get(client.KEY_MOUNT_STATUS)
	status := v.(client.MountStatus)
	curveadm.WriteOut("\n")
	curveadm.WriteOut("Mount Point : %s\n", status.MountPoint)
	curveadm.WriteOut("Container Id: %s\n", status.ContainerId)
	curveadm.WriteOut("Mount Status: %s\n", status.Status)
	return nil
}

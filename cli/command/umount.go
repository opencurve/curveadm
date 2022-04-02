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
	"github.com/opencurve/curveadm/internal/task/task/fs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type umountOptions struct {
	mountPoint string
}

func NewUmountCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options umountOptions

	cmd := &cobra.Command{
		Use:   "umount MOUNT_POINT",
		Short: "Umount filesystem",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.mountPoint = args[0]
			return runUmount(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runUmount(curveadm *cli.CurveAdm, options umountOptions) error {
	mountPoint := strings.TrimSuffix(options.mountPoint, "/")
	curveadm.MemStorage().Set(fs.KEY_MOUNT_POINT, mountPoint)
	err := tasks.ExecTasks(tasks.UMOUNT_FILESYSTEM, curveadm, nil)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

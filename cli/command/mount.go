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
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/fs"
	"github.com/opencurve/curveadm/internal/task/task/fs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	mountExample = `Examples:
  $ curveadm mount /s3_001 /path/to/mount -c client.yaml  # Mount CurveFS '/s3_001' to '/path/to/mount'`
)

type mountOptions struct {
	mountFSName string
	mountPoint  string
	filename    string
}

func NewMountCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options mountOptions

	cmd := &cobra.Command{
		Use:     "mount NAME_OF_CURVEFS MOUNT_POINT [OPTION]",
		Short:   "Mount filesystem",
		Args:    utils.ExactArgs(2),
		Example: mountExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.mountFSName = args[0]
			options.mountPoint = args[1]
			return runMount(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")

	return cmd
}

func runMount(curveadm *cli.CurveAdm, options mountOptions) error {
	cc, err := client.ParseClientConfig(options.filename)
	if err != nil {
		return err
	}

	// check mount point
	mountFSName := options.mountFSName
	mountPoint := strings.TrimSuffix(options.mountPoint, "/")
	err = utils.CheckMountPoint(mountPoint)
	if err != nil {
		return err
	}

	// check mount status
	curveadm.MemStorage().Set(fs.KEY_MOUNT_FSNAME, mountFSName)
	curveadm.MemStorage().Set(fs.KEY_MOUNT_POINT, mountPoint)
	err = tasks.ExecTasks(tasks.CHECK_MOUNT_STATUS, curveadm, cc)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	} else {
		v := curveadm.MemStorage().Get(fs.KEY_MOUNT_STATUS)
		status := v.(fs.MountStatus).Status
		if status != fs.STATUS_UNMOUNTED {
			return fmt.Errorf("path mounted, please run 'curveadm umount %s' first", mountPoint)
		}
	}

	// mount file system
	err = tasks.ExecTasks(tasks.MOUNT_FILESYSTEM, curveadm, cc)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}

	return nil
}

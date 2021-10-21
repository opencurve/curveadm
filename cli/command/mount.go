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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/client"
	"github.com/opencurve/curveadm/internal/tasks/task"
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
	mountPoint := options.mountPoint
	mountFSName := options.mountFSName

	if !utils.PathExist(mountPoint) {
		return fmt.Errorf("mount path '%s' not exist", mountPoint)
	} else if config, err := configure.ParseClientConfig(options.filename); err != nil {
		return err
	} else if t, err := client.NewMountFSTask(curveadm, mountPoint, mountFSName, config); err != nil {
		return err
	} else if err := task.ParallelExecute(1, []*task.Task{t}, task.Options{SilentSubBar: true}); err != nil {
		return err
	}

	return nil
}

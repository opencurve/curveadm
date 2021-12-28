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
 * Created Date: 2021-01-10
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	unmapExample = `Examples:
  $ curveadm unmap user:volume                          # Unmap volume with default configure file 
  $ curveadm unmap user:volume -c /path/to/client.yaml  # Unmap volume with specified configure file`
)

type unmapOptions struct {
	image    string
	filename string
}

func NewUnmapCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options unmapOptions

	cmd := &cobra.Command{
		Use:     "unmap USER:VOLUME [OPTION]",
		Short:   "Unmap nbd device",
		Args:    utils.ExactArgs(1),
		Example: unmapExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return runUnmap(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")

	return cmd
}

func runUnmap(curveadm *cli.CurveAdm, options unmapOptions) error {
	user, volume, err := parseImage(options.image)
	if err != nil {
		return err
	}

	cc, err := client.ParseClientConfig(options.filename)
	if err != nil {
		return err
	}

	// mount file system
	curveadm.MemStorage().Set(bs.KEY_MAP_OPTION, bs.MapOption{
		User:   user,
		Volume: volume,
	})
	err = tasks.ExecTasks(tasks.UNMAP_IMAGE, curveadm, cc)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

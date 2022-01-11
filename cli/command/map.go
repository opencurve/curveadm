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
 * Created Date: 2021-01-09
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	mapExample = `Examples:
  $ curveadm map user:volume --create                 # Map volume which created by automatic
  $ curveadm map user:volume --create --size=10GB     # Map volume which size is 10GB and created by automatic
  $ curveadm map user:volume -c /path/to/client.yaml  # Map volume with specified configure file`
)

type mapOptions struct {
	image    string
	filename string
	size     string
	create   bool
}

func NewMapCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options mapOptions

	cmd := &cobra.Command{
		Use:     "map USER:VOLUME [OPTION]",
		Short:   "Map a volume to nbd device",
		Args:    utils.ExactArgs(1),
		Example: mapExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return runMap(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")
	flags.BoolVarP(&options.create, "create", "", false, "")
	flags.StringVarP(&options.size, "size", "", "10GB", "Specify volume size")

	return cmd
}

func parseImage(image string) (user, volume string, err error) {
	items := strings.Split(image, ":")
	if len(items) != 2 || len(items[0]) == 0 || len(items[1]) == 0 {
		return "", "", fmt.Errorf("invalid image name")
	}
	return items[0], items[1], nil
}

func parseSize(size string) (int, error) {
	if !strings.HasSuffix(size, "GB") {
		return 0, fmt.Errorf("invalid size")
	}

	size = strings.TrimSuffix(size, "GB")
	return strconv.Atoi(size)
}

func runMap(curveadm *cli.CurveAdm, options mapOptions) error {
	user, volume, err := parseImage(options.image)
	if err != nil {
		return err
	}

	size, err := parseSize(options.size)
	if err != nil {
		return err
	}

	// config
	cc, err := client.ParseClientConfig(options.filename)
	if err != nil {
		return err
	}

	// mount file system
	curveadm.MemStorage().Set(bs.KEY_MAP_OPTION, bs.MapOption{
		User:   user,
		Volume: volume,
		Create: options.create,
		Size:   size,
	})
	if err = tasks.ExecTasks(tasks.START_NEBD_SERVICE, curveadm, cc); err == nil {
		err = tasks.ExecTasks(tasks.MAP_IMAGE, curveadm, cc)
	}

	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

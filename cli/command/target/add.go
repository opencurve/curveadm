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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package target

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	START_NEBD_SERVICE  = tasks.START_NEBD_SERVICE
	START_TARGET_DAEMON = tasks.START_TARGET_DAEMON
	ADD_TARGET          = tasks.ADD_TARGET
)

var (
	ADD_TARGET_STEPS = []int{
		START_NEBD_SERVICE,
		START_TARGET_DAEMON,
		ADD_TARGET,
	}
)

type addOptions struct {
	image    string
	filename string
	size     string
	create   bool
}

func NewAddCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options addOptions

	cmd := &cobra.Command{
		Use:   "add USER:VOLUME [OPTION]",
		Short: "Add a target of CurveBS",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return runAdd(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")
	flags.BoolVarP(&options.create, "create", "", false, "Create volume iff not exist")
	flags.StringVarP(&options.size, "size", "", "10GB", "Specify volume size")

	return cmd
}

func parseImage(image string) (user, volume string, err error) {
	items := strings.Split(image, ":")
	if len(items) != 2 || len(items[0]) == 0 || len(items[1]) == 0 {
		return "", "", fmt.Errorf("invalid volume format, please run --help to get example")
	}
	if !strings.HasPrefix(items[1], "/") {
		return "", "", fmt.Errorf("invalid volume format, image name must start with /")
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

func runAdd(curveadm *cli.CurveAdm, options addOptions) error {
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
	curveadm.MemStorage().Set(bs.KEY_ADD_TARGET_OPTION, bs.AddTargetOption{
		User:   user,
		Volume: volume,
		Create: options.create,
		Size:   size,
	})

	for _, step := range ADD_TARGET_STEPS {
		err := tasks.ExecTasks(step, curveadm, cc)
		if err != nil {
			return curveadm.NewPromptError(err, "")
		}
	}

	return nil
}

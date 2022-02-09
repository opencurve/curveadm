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
 * Created Date: 2022-02-09
 * Author: Jingli Chen (Wine93)
 */

package target

import (
	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type deleteOptions struct {
	tid      string
	filename string
}

func NewDeleteCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options deleteOptions

	cmd := &cobra.Command{
		Use:   "rm TID [OPTION]",
		Aliases: []string{"delete"},
		Short: "Delete a target of CurveBS",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.tid = args[0]
			return runDelete(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")

	return cmd
}

func runDelete(curveadm *cli.CurveAdm, options deleteOptions) error {
	// config
	cc, err := client.ParseClientConfig(options.filename)
	if err != nil {
		return err
	}

	curveadm.MemStorage().Set(bs.KEY_DELETE_TID, options.tid)
	err = tasks.ExecTasks(tasks.DELETE_TARGET, curveadm, cc)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}

	return nil
}

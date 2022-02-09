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
	task "github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	filename string
}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List targets",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")

	return cmd
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	// config
	cc, err := client.ParseClientConfig(options.filename)
	if err != nil {
		return err
	}

	err = tasks.ExecTasks(tasks.LIST_TARGETS, curveadm, cc)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}

	// display target
	targets := []task.Target{}
	m := curveadm.MemStorage().Map
	for _, v := range m {
		target := v.(*task.Target)
		targets = append(targets, *target)
	}

	output := tui.FormatTargets(targets)
	curveadm.WriteOut("\n")
	curveadm.WriteOut("%s", output)
	return nil
}

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
 * Created Date: 2021-12-27
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/format"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui/format"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	formatExample = `Examples:
  $ curveadm format -f /path/to/format.yaml           # Format chunkfile pool with specified configure file
  $ curveadm format --status -f /path/to/format.yaml  # Display formatting status`
)

type formatOptions struct {
	filename   string
	showStatus bool
}

func NewFormatCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options formatOptions

	cmd := &cobra.Command{
		Use:     "format",
		Short:   "Format chunkfile pool",
		Args:    cliutil.NoArgs,
		Example: formatExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFormat(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "formatting", "f", "format.yaml", "Specify the configure file for formatting chunkfile pool")
	flags.BoolVarP(&options.showStatus, "status", "", false, "Show formatting status")

	return cmd
}

func runFormat(curveadm *cli.CurveAdm, options formatOptions) error {
	fcs, err := format.ParseFormat(options.filename)
	if err != nil {
		return err
	}

	if !options.showStatus {
		err = tasks.ExecTasks(tasks.FORMAT_CHUNKFILE_POOL, curveadm, fcs)
	} else if err = tasks.ExecTasks(tasks.GET_FORMAT_STATUS, curveadm, fcs); err != nil {
		statuses := []bs.FormatStatus{}
		for _, v := range curveadm.MemStorage().Map {
			status := v.(bs.FormatStatus)
			statuses = append(statuses, status)
		}
		output := tui.FormatStatus(statuses)
		curveadm.WriteOut("%s\n", output)
	}

	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

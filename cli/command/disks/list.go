/*
 *  Copyright (c) 2023 NetEase Inc.
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
 * Created Date: 2023-02-24
 * Author: Lijin Xiong (lijin.xiong@zstack.io)
 */

package disks

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	host string
}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions
	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List disk information",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "*", "List disk of host")

	return cmd
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	var err error
	var diskRecords []storage.Disk

	if options.host == "*" {
		diskRecords = curveadm.DiskRecords()
	} else {
		if diskRecords, err = curveadm.Storage().GetDisk(common.DISK_FILTER_HOST,
			options.host); err != nil {
			return err
		}
	}

	output := tui.FormatDisks(diskRecords)
	curveadm.WriteOut(output)
	return nil
}

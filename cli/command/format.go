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
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type formatOptions struct {
	filename   string
	showStatus bool
}

func NewFormatCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options formatOptions

	cmd := &cobra.Command{
		Use:   "format",
		Short: "Format chunkfile pool for chunkserver service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFormat(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "disk", "f", "format.yaml", "Specify the path of disk list file")
	flags.BoolVarP(&options.showStatus, "status", "", false, "Show format status")

	return cmd
}

func runFormat(curveadm *cli.CurveAdm, options formatOptions) error {
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Role: topology.ROLE_CHUNKSERVER,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return fmt.Errorf("service not found")
	}

	disks, err := format.ParseDiskList(options.filename)
	if err != nil {
		return err
	} else if len(disks) == 0 {
		return nil
	}

	if err := tasks.ExecTasks(tasks.RESTART_SERVICE, curveadm, dcs); err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

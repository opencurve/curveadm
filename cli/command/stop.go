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
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type stopOptions struct {
	id   string
	role string
	host string
}

func NewStopCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options stopOptions

	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStop(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.id, "id", "", "*", "Specify service id")
	flags.StringVarP(&options.role, "role", "", "*", "Specify service role")
	flags.StringVarP(&options.host, "host", "", "*", "Specify service host")

	return cmd
}

func runStop(curveadm *cli.CurveAdm, options stopOptions) error {
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})

	if len(dcs) == 0 {
		return fmt.Errorf("service not found")
	}

	// stop service
	curveadm.WriteOut("Warning: Stop all service now, client IO will be hang!\n")
	if pass := tui.ConfirmYes("Do you want to continue? [YES/No]: "); !pass {
		curveadm.WriteOut("Stop canceled\n")
		return nil
	}

	if err := tasks.ExecTasks(tasks.STOP_SERVICE, curveadm, dcs); err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return nil
}

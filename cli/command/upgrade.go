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
 * Created Date: 2022-01-16
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var UPGRADE_STEPS = []int{
	tasks.STOP_SERVICE,
	tasks.CLEAN_SERVICE,
	tasks.PULL_IMAGE,
	tasks.CREATE_CONTAINER,
	tasks.SYNC_CONFIG,
	tasks.START_SERVICE,
}

type upgradeOptions struct {
	id    string
	role  string
	host  string
	force bool
}

func NewUpgradeCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options upgradeOptions

	cmd := &cobra.Command{
		Use:   "upgrade",
		Short: "Upgrade service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.id, "id", "", "*", "Specify service id")
	flags.StringVarP(&options.role, "role", "", "*", "Specify service role")
	flags.StringVarP(&options.host, "host", "", "*", "Specify service host")
	flags.BoolVarP(&options.force, "force", "f", false, "Never prompt")

	return cmd
}

func runUpgrade(curveadm *cli.CurveAdm, options upgradeOptions) error {
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

	curveadm.WriteOut("Upgrade %d services one by one\n", len(dcs))
	for i, dc := range dcs {
		curveadm.WriteOut("\n")
		curveadm.WriteOut("Upgrade %d/%d service: \n", i+1, len(dcs))
		curveadm.WriteOut("  + host=%s  role=%s  image=%s\n", dc.GetHost(), dc.GetRole(), dc.GetContainerImage())
		if !options.force && !tui.ConfirmYes("Do you want to continue?") {
			curveadm.WriteOut("Upgrade abort\n")
			break
		}

		for _, step := range UPGRADE_STEPS {
			if step == tasks.CLEAN_SERVICE {
				curveadm.MemStorage().Set(task.KEY_CLEAN_ITEMS, []string{"container"})
				curveadm.MemStorage().Set(task.KEY_RECYCLE, true)
			}
			err := tasks.ExecTasks(step, curveadm, dc)
			if err != nil {
				return err
			}
		}

		curveadm.WriteOut("Upgrade %d/%d sucess\n", i+1, len(dcs))
	}

	return nil
}

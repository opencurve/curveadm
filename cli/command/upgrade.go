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

// __SIGN_BY_WINE93__

package command

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	UPGRADE_PLAYBOOK_STEPS = []int{
		// TODO(P0): we can skip it for upgrade one service more than once
		playbook.PULL_IMAGE,
		playbook.STOP_SERVICE,
		playbook.CLEAN_SERVICE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SERVICE,
	}
)

type upgradeOptions struct {
	id    string
	role  string
	host  string
	force bool
}

func NewUpgradeCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options upgradeOptions

	cmd := &cobra.Command{
		Use:   "upgrade [OPTIONS]",
		Short: "Upgrade service",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkCommonOptions(curveadm, options.id, options.role, options.host)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpgrade(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify service id")
	flags.StringVar(&options.role, "role", "*", "Specify service role")
	flags.StringVar(&options.host, "host", "*", "Specify service host")
	flags.BoolVarP(&options.force, "force", "f", false, "Never prompt")

	return cmd
}

func genUpgradePlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options upgradeOptions) (*playbook.Playbook, error) {
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := UPGRADE_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
			Options: map[string]interface{}{
				comm.KEY_CLEAN_ITEMS:      []string{comm.CLEAN_ITEM_CONTAINER},
				comm.KEY_CLEAN_BY_RECYCLE: true,
			},
		})
	}
	return pb, nil
}

func displayTitle(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, options upgradeOptions) {
	total := len(dcs)
	if options.force {
		curveadm.WriteOutln(color.YellowString("Upgrade %d services at once", total))
	} else {
		curveadm.WriteOutln(color.YellowString("Upgrade %d services one by one", total))
	}
	curveadm.WriteOutln(color.YellowString("Upgrade services: %s", serviceStats(dcs)))
}

func upgradeAtOnce(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, options upgradeOptions) error {
	// 1) display upgrade title
	displayTitle(curveadm, dcs, options)

	// 2) confirm by user
	if pass := tui.ConfirmYes(tui.DEFAULT_CONFIRM_PROMPT); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("upgrade service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 3) generate upgrade playbook
	pb, err := genUpgradePlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 4) run playbook
	err = pb.Run()
	if err != nil {
		return err
	}

	// 5) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Upgrade %d services success :)", len(dcs)))
	return nil
}

func upgradeOneByOne(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, options upgradeOptions) error {
	// 1) display upgrade title
	displayTitle(curveadm, dcs, options)

	// 2) upgrade service one by one
	total := len(dcs)
	for i, dc := range dcs {
		// 2.1) confirm by user
		curveadm.WriteOutln("")
		curveadm.WriteOutln("Upgrade %s service:", color.BlueString("%d/%d", i+1, total))
		curveadm.WriteOutln("  + host=%s  role=%s  image=%s", dc.GetHost(), dc.GetRole(), dc.GetContainerImage())
		if pass := tui.ConfirmYes(tui.DEFAULT_CONFIRM_PROMPT); !pass {
			curveadm.WriteOut(tui.PromptCancelOpetation("upgrade service"))
			return errno.ERR_CANCEL_OPERATION
		}

		// 2.2) generate upgrade playbook
		pb, err := genUpgradePlaybook(curveadm, []*topology.DeployConfig{dc}, options)
		if err != nil {
			return err
		}

		// 2.3) run playbook
		err = pb.Run()
		if err != nil {
			return err
		}

		// 2.4) print success prompt
		curveadm.WriteOutln("")
		curveadm.WriteOutln(color.GreenString("Upgrade %d/%d sucess :)"), i+1, total)
	}
	return nil
}

func runUpgrade(curveadm *cli.CurveAdm, options upgradeOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) filter deploy config
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return errno.ERR_NO_SERVICES_MATCHED
	}

	// 3.1) upgrade service at once
	if options.force {
		return upgradeAtOnce(curveadm, dcs, options)
	}

	// 3.2) OR upgrade service one by one
	return upgradeOneByOne(curveadm, dcs, options)
}

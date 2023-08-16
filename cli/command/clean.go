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

// __SIGN_BY_WINE93__

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	utils "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	CLEAN_EXAMPLE = `Examples:
  $ curveadm clean                               # Clean everything for all services
  $ curveadm clean --only='log,data'             # Clean log and data for all services
  $ curveadm clean --role=etcd --only=container  # Clean container for etcd services`
)

var (
	CLEAN_PLAYBOOK_STEPS = []int{
		playbook.CLEAN_SERVICE,
	}

	CLEAN_ITEMS = []string{
		comm.CLEAN_ITEM_LOG,
		comm.CLEAN_ITEM_DATA,
		comm.CLEAN_ITEM_CONTAINER,
	}
)

type cleanOptions struct {
	id             string
	role           string
	host           string
	only           []string
	withoutRecycle bool
}

func checkCleanOptions(curveadm *cli.CurveAdm, options cleanOptions) error {
	supported := utils.Slice2Map(CLEAN_ITEMS)
	for _, item := range options.only {
		if !supported[item] {
			return errno.ERR_UNSUPPORT_CLEAN_ITEM.
				F("clean item: %s", item)
		}
	}
	return checkCommonOptions(curveadm, options.id, options.role, options.host)
}

func NewCleanCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options cleanOptions

	cmd := &cobra.Command{
		Use:     "clean [OPTIONS]",
		Short:   "Clean service's environment",
		Args:    cliutil.NoArgs,
		Example: CLEAN_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkCleanOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify service id")
	flags.StringVar(&options.role, "role", "*", "Specify service role")
	flags.StringVar(&options.host, "host", "*", "Specify service host")
	flags.StringSliceVarP(&options.only, "only", "o", CLEAN_ITEMS, "Specify clean item")
	flags.BoolVar(&options.withoutRecycle, "no-recycle", false, "Remove data directory directly instead of recycle chunks")

	return cmd
}

func genCleanPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options cleanOptions) (*playbook.Playbook, error) {
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := CLEAN_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
			Options: map[string]interface{}{
				comm.KEY_CLEAN_ITEMS:      options.only,
				comm.KEY_CLEAN_BY_RECYCLE: !options.withoutRecycle,
			},
		})
	}
	return pb, nil
}

func runClean(curveadm *cli.CurveAdm, options cleanOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) generate clean playbook
	pb, err := genCleanPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptCleanService(options.role, options.host, options.only)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("clean service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	return pb.Run()
}

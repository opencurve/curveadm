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
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	START_PLAYBOOK_STEPS = []int{
		playbook.START_SERVICE,
	}
)

type startOptions struct {
	id   string
	role string
	host string
}

func checkCommonOptions(curveadm *cli.CurveAdm, id, role, host string) error {
	items := []struct {
		key      string
		callback func(string) error
	}{
		{id, curveadm.CheckId},
		{role, curveadm.CheckRole},
		{host, curveadm.CheckHost},
	}

	for _, item := range items {
		if item.key == "*" {
			continue
		}
		err := item.callback(item.key)
		if err != nil {
			return err
		}
	}
	return nil
}

func NewStartCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options startOptions

	cmd := &cobra.Command{
		Use:   "start [OPTIONS]",
		Short: "Start service",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkCommonOptions(curveadm, options.id, options.role, options.host)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify service id")
	flags.StringVar(&options.role, "role", "*", "Specify service role")
	flags.StringVar(&options.host, "host", "*", "Specify service host")

	return cmd
}

func genStartPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options startOptions) (*playbook.Playbook, error) {
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := START_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
		})
	}
	return pb, nil
}

func runStart(curveadm *cli.CurveAdm, options startOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) generate start playbook
	pb, err := genStartPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptStartService(options.id, options.role, options.host)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("start service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	return pb.Run()
}

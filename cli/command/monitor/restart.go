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
* Project: Curveadm
* Created Date: 2023-04-28
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	RESTART_PLAYBOOK_STEPS = []int{
		playbook.RESTART_MONITOR_SERVICE,
	}
)

type restartOptions struct {
	id   string
	role string
	host string
}

func NewRestartCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options restartOptions
	cmd := &cobra.Command{
		Use:   "restart [OPTIONS]",
		Short: "Restart monitor service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRestart(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify monitor service id")
	flags.StringVar(&options.role, "role", "*", "Specify monitor service role")
	flags.StringVar(&options.host, "host", "*", "Specify monitor service host")

	return cmd
}

func genRestartPlaybook(curveadm *cli.CurveAdm,
	mcs []*configure.MonitorConfig,
	options restartOptions) (*playbook.Playbook, error) {
	mcs = configure.FilterMonitorConfig(curveadm, mcs, configure.FilterMonitorOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(mcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := RESTART_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: mcs,
		})
	}
	return pb, nil
}

func runRestart(curveadm *cli.CurveAdm, options restartOptions) error {
	// 1) parse monitor configure
	mcs, err := parseMonitorConfig(curveadm)
	if err != nil {
		return err
	}

	// 2) generate restart playbook
	pb, err := genRestartPlaybook(curveadm, mcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptRestartService(options.id, options.role, options.host)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("restart monitor service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	return pb.Run()
}

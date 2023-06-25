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
* Created Date: 2023-04-27
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	CLEAN_EXAMPLE = `Examples:
  $ curveadm monitor clean                                  # Clean everything for monitor
  $ curveadm monitor clean --only='data'                    # Clean data for monitor
  $ curveadm monitor clean --role=grafana --only=container  # Clean container for grafana service`
)

var (
	CLEAN_PLAYBOOK_STEPS = []int{
		playbook.CLEAN_MONITOR,
	}

	CLEAN_ITEMS = []string{
		comm.CLEAN_ITEM_DATA,
		comm.CLEAN_ITEM_CONTAINER,
	}
)

type cleanOptions struct {
	id   string
	role string
	host string
	only []string
}

func NewCleanCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options cleanOptions

	cmd := &cobra.Command{
		Use:     "clean [OPTIONS]",
		Short:   "Clean monitor's environment",
		Args:    cliutil.NoArgs,
		Example: CLEAN_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.id, "id", "*", "Specify monitor service id")
	flags.StringVar(&options.role, "role", "*", "Specify monitor service role")
	flags.StringVar(&options.host, "host", "*", "Specify monitor service host")
	flags.StringSliceVarP(&options.only, "only", "o", CLEAN_ITEMS, "Specify clean item")
	return cmd
}

func genCleanPlaybook(curveadm *cli.CurveAdm,
	mcs []*configure.MonitorConfig,
	options cleanOptions) (*playbook.Playbook, error) {
	mcs = configure.FilterMonitorConfig(curveadm, mcs, configure.FilterMonitorOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(mcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}
	steps := CLEAN_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: mcs,
			Options: map[string]interface{}{
				comm.KEY_CLEAN_ITEMS: options.only,
			},
		})
	}
	return pb, nil
}

func runClean(curveadm *cli.CurveAdm, options cleanOptions) error {
	// 1) parse monitor config
	mcs, err := parseMonitorConfig(curveadm)
	if err != nil {
		return err
	}

	// 2) generate clean playbook
	pb, err := genCleanPlaybook(curveadm, mcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptCleanService(options.role, options.host, options.only)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("clean monitor service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	err = pb.Run()
	if err != nil {
		return err
	}
	return nil
}

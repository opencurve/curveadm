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
 * Created Date: 2022-02-09
 * Author: Jingli Chen (Wine93)
 */

package target

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	LIST_PLAYBOOK_STEPS = []int{
		playbook.LIST_TARGETS,
	}
)

type listOptions struct {
	host string
}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List targets",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")

	return cmd
}

func genListPlaybook(curveadm *cli.CurveAdm, options listOptions) (*playbook.Playbook, error) {
	steps := LIST_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: nil,
			Options: map[string]interface{}{
				comm.KEY_TARGET_OPTIONS: bs.TargetOption{
					Host: options.host,
				},
			},
		})
	}
	return pb, nil
}

func displayTargets(curveadm *cli.CurveAdm) {
	targets := []bs.Target{}
	value := curveadm.MemStorage().Get(comm.KEY_ALL_TARGETS)
	if value != nil {
		m := value.(map[string]*bs.Target)
		for _, target := range m {
			targets = append(targets, *target)
		}
	}

	output := tui.FormatTargets(targets)
	curveadm.WriteOutln("")
	curveadm.WriteOut(output)
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	// 1) generate list playbook
	pb, err := genListPlaybook(curveadm, options)
	if err != nil {
		return err
	}

	// 2) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 3) print targets
	displayTargets(curveadm)
	return nil
}

/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-06-23
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/storage"
	pg "github.com/opencurve/curveadm/internal/task/task/playground"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct{}

var GET_PLAYGROUND_STATUS_PLAYBOOK_STEPS = []int{
	playbook.GET_PLAYGROUND_STATUS,
}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List playgrounds",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func genListPlaybook(curveadm *cli.CurveAdm,
	playgrounds []storage.Playground) (*playbook.Playbook, error) {
	configs := []interface{}{}
	for _, playground := range playgrounds {
		configs = append(configs, playground)
	}
	steps := GET_PLAYGROUND_STATUS_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: configs,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: true,
			},
		})
	}
	return pb, nil
}

func displayPlaygrounds(curveadm *cli.CurveAdm) {
	statuses := []pg.PlaygroundStatus{}
	value := curveadm.MemStorage().Get(comm.KEY_ALL_PLAYGROUNDS_STATUS)
	if value != nil {
		m := value.(map[string]pg.PlaygroundStatus)
		for _, status := range m {
			statuses = append(statuses, status)
		}
	}

	output := tui.FormatPlayground(statuses)
	curveadm.WriteOutln("")
	curveadm.WriteOut(output)
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	// 1) get playgrounds
	playgrounds, err := curveadm.Storage().GetPlaygrounds("%")
	if err != nil {
		return errno.ERR_GET_ALL_PLAYGROUND_FAILED.E(err)
	}

	// 2) gen list playground
	pb, err := genListPlaybook(curveadm, playgrounds)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print playgrounds
	displayPlaygrounds(curveadm)
	return nil
}

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
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	name string
}

var REMOVE_PLAYGROUND_PLAYBOOK_STEPS = []int{
	playbook.REMOVE_PLAYGROUND,
}

func NewRemoveCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options removeOptions

	cmd := &cobra.Command{
		Use:     "rm PLAYGROUND",
		Aliases: []string{"delete"},
		Short:   "Remove playground",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runRemove(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func genRemovePlaybook(curveadm *cli.CurveAdm,
	options removeOptions) (*playbook.Playbook, error) {
	steps := REMOVE_PLAYGROUND_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type: step,
			Configs: &configure.PlaygroundConfig{
				Name: options.name,
			},
		})
	}
	return pb, nil
}

func runRemove(curveadm *cli.CurveAdm, options removeOptions) error {
	// 1) get playground
	name := options.name
	playgrounds, err := curveadm.Storage().GetPlaygrounds(name)
	if err != nil {
		return errno.ERR_GET_PLAYGROUND_BY_NAME_FAILED.E(err)
	} else if len(playgrounds) == 0 {
		return errno.ERR_PLAYGROUND_NOT_FOUND.
			F("playground=%s", name)
	}

	// 2) generate remove playground
	pb, err := genRemovePlaybook(curveadm, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln("Playground '%s' removed.", name)
	return nil
}

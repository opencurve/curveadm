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
 * Created Date: 2022-07-26
 * Author: Jingli Chen (Wine93)
 */

package client

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/storage"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/client"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	GET_STATUS_PLAYBOOK_STEPS = []int{
		playbook.INIT_CLIENT_STATUS,
		playbook.GET_CLIENT_STATUS,
	}
)

type statusOptions struct {
	verbose bool
}

func NewStatusCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options statusOptions

	cmd := &cobra.Command{
		Use:   "status [OPTIONS]",
		Short: "Display client status",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output for status")

	return cmd
}

func genStatusPlaybook(curveadm *cli.CurveAdm,
	clients []storage.Client,
	options statusOptions) (*playbook.Playbook, error) {
	config := []interface{}{}
	for _, client := range clients {
		config = append(config, client)
	}

	steps := GET_STATUS_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: config,
			Options: map[string]interface{}{
				comm.KEY_CLIENT_STATUS_VERBOSE: options.verbose,
			},
			ExecOptions: playbook.ExecOptions{
				SilentSubBar:  true,
				SilentMainBar: step == playbook.INIT_CLIENT_STATUS,
				SkipError:     true,
			},
		})
	}
	return pb, nil
}

func displayStatus(curveadm *cli.CurveAdm, clients []storage.Client, options statusOptions) {
	statuses := []task.ClientStatus{}
	v := curveadm.MemStorage().Get(comm.KEY_ALL_CLIENT_STATUS)
	if v != nil {
		m := v.(map[string]task.ClientStatus)
		for _, status := range m {
			statuses = append(statuses, status)
		}
	}

	output := tui.FormatStatus(statuses, options.verbose)
	if len(clients) > 0 {
		curveadm.WriteOutln("")
	}
	curveadm.WriteOut(output)
}

func runStatus(curveadm *cli.CurveAdm, options statusOptions) error {
	// 1) get all clients
	clients, err := curveadm.Storage().GetClients()
	if err != nil {
		return errno.ERR_GET_ALL_CLIENTS_FAILED.E(err)
	}

	// 2) generate get status playbook
	pb, err := genStatusPlaybook(curveadm, clients, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()

	// 4) display service status
	displayStatus(curveadm, clients, options)
	return err
}

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
 * Created Date: 2021-11-26
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/storage"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	SUPPORT_UPLOAD_URL_FORMAT = "http://curveadm.aspirer.wang:19301/upload?path=%s"
)

var (
	SUPPORT_PLAYBOOK_STEPS = []int{
		playbook.INIT_SUPPORT,
		//playbook.COLLECT_REPORT,
		playbook.COLLECT_CURVEADM,
		playbook.COLLECT_SERVICE,
	}
)

type supportOptions struct {
	ids []string
}

func NewSupportCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options supportOptions

	cmd := &cobra.Command{
		Use:   "support [OPTIONS]",
		Short: "Get support from Curve team",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runSupport(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&options.ids, "client", "c", []string{}, "Specify client id")

	return cmd
}

func getClients(curveadm *cli.CurveAdm,
	options supportOptions) ([]storage.Client, error) {
	out := []storage.Client{}
	for _, id := range options.ids {
		clients, err := curveadm.Storage().GetClient(id)
		if err != nil {
			return nil, errno.ERR_GET_CLIENT_BY_ID_FAILED.E(err)
		} else if len(clients) == 0 {
			return nil, errno.ERR_CLIENT_ID_NOT_FOUND.
				F("client id: %s", id)
		}
		out = append(out, clients[0])
	}
	return out, nil
}

func genSupportPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options supportOptions) (*playbook.Playbook, error) {
	clients, err := getClients(curveadm, options)
	if err != nil {
		return nil, err
	}
	cconfig := []interface{}{}
	for _, client := range clients {
		cconfig = append(cconfig, client)
	}

	steps := SUPPORT_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	if len(clients) > 0 {
		steps = append(steps, playbook.COLLECT_CLIENT)
	}
	for _, step := range steps {
		config := dcs
		switch step {
		case playbook.INIT_SUPPORT:
			config = config[:1]
		case playbook.COLLECT_REPORT:
			config = config[:1]
		case playbook.COLLECT_CURVEADM:
			config = config[:1]
		}

		if step == playbook.COLLECT_CLIENT {
			pb.AddStep(&playbook.PlaybookStep{
				Type:    step,
				Configs: cconfig,
			})
		} else {
			pb.AddStep(&playbook.PlaybookStep{
				Type:    step,
				Configs: config,
			})
		}
	}
	return pb, nil
}

func runSupport(curveadm *cli.CurveAdm, options supportOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) generate secret
	secret := utils.RandString(32)
	curveadm.MemStorage().Set(comm.KEY_SECRET, secret)
	curveadm.MemStorage().Set(comm.KEY_ALL_CLIENT_IDS, options.ids)
	curveadm.MemStorage().Set(comm.KEY_SUPPORT_UPLOAD_URL_FORMAT, SUPPORT_UPLOAD_URL_FORMAT)

	// 3) generate support playbook
	pb, err := genSupportPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 4) confirm by user
	if pass := tui.ConfirmYes(tui.PromptCollectService()); !pass {
		return nil
	}

	// 5) run playbook
	err = pb.Run()
	if err != nil {
		return err
	}

	// 6) print secret
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Upload success, please tell the below secret key to curve team :)"))
	curveadm.WriteOut(color.GreenString("secret key: "))
	curveadm.WriteOutln(color.YellowString(secret))
	return nil
}

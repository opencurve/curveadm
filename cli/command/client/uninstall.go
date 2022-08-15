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
 * Created Date: 2022-08-08
 * Author: Jingli Chen (Wine93)
 */

package client

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	UNINSTALL_CURVE_CLEINT_PLAYBOOK_STEPS = []int{
		playbook.DETECT_OS_RELEASE,
		playbook.UNINSTALL_CLIENT,
	}
)

type uninstallOptions struct {
	kind string
	host string
}

func checkUninstallOptions(curveadm *cli.CurveAdm, options uninstallOptions) error {
	kind := options.kind
	if kind != topology.KIND_CURVEBS && kind != topology.KIND_CURVEFS {
		return errno.ERR_UNSUPPORT_CLIENT_KIND.
			F("kind: %s", kind)
	}
	return nil
}

func NewUninstallCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options uninstallOptions

	cmd := &cobra.Command{
		Use:   "uninstall KIND [OPTIONS]",
		Short: "Uninstall CurveBS/CurveFS client",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.kind = args[0]
			return checkUninstallOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUninstall(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "local", "Specify uninstall target host")

	return cmd
}

func genUninstallPlaybook(curveadm *cli.CurveAdm,
	v interface{}, options uninstallOptions) (*playbook.Playbook, error) {
	steps := UNINSTALL_CURVE_CLEINT_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: nil,
			Options: map[string]interface{}{
				comm.KEY_CLIENT_HOST: options.host,
				comm.KEY_CLIENT_KIND: options.kind,
			},
		})
	}
	return pb, nil
}

func runUninstall(curveadm *cli.CurveAdm, options uninstallOptions) error {
	// 1) generate map playbook
	pb, err := genUninstallPlaybook(curveadm, nil, options)
	if err != nil {
		return err
	}

	// 2) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("UnInstall %s %s success ^_^"),
		options.host, options.kind)
	return nil
}

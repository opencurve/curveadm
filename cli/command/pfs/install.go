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

package pfs

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	utils "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	INSTALL_PFS_PLAYBOOK_STEPS = []int{
		playbook.DETECT_OS_RELEASE,
		playbook.INSTALL_POLARFS,
	}
)

type installOptions struct {
	host     string
	filename string
}

func checkInstallOptions(curveadm *cli.CurveAdm, options installOptions) error {
	if !utils.PathExist(options.filename) {
		return errno.ERR_CLIENT_CONFIGURE_FILE_NOT_EXIST.
			F("%s: no such file", utils.AbsPath(options.filename))
	}
	return nil
}

func NewInstallCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options installOptions

	cmd := &cobra.Command{
		Use:   "install [OPTIONS]",
		Short: "Install PolarFS",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkInstallOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "local", "Specify install target host")
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")

	return cmd
}

func genInstallPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options installOptions) (*playbook.Playbook, error) {
	steps := INSTALL_PFS_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: ccs,
			Options: map[string]interface{}{
				comm.KEY_POLARFS_HOST: options.host,
			},
		})
	}
	return pb, nil
}

func runInstall(curveadm *cli.CurveAdm, options installOptions) error {
	// 1) parse client configure
	cc, err := configure.ParseClientConfig(options.filename)
	if err != nil {
		return err
	} else if cc.GetKind() != topology.KIND_CURVEBS {
		return errno.ERR_REQUIRE_CURVEBS_KIND_CLIENT_CONFIGURE_FILE.
			F("kind: %s", cc.GetKind())
	}

	// TODO(P0): get local ip failed
	// TODO(P0): log dir
	// 2) generate map playbook
	pb, err := genInstallPlaybook(curveadm, []*configure.ClientConfig{cc}, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Install polarfs to %s success ^_^"), options.host)
	return nil
}

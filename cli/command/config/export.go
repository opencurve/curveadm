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
 * Project: CurveAdm
 * Created Date: 2023-11-9
 * Author: Jiang Jun (youarefree123)
 */

// __SIGN_BY_YOUAREFREE123__

package config

import (
	"path"
	"strings"

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
	GET_EXPORT_PLAYBOOK_STEPS = []int{
		playbook.EXPORT_TOOLSV2_CONF,
	}
)

type exportOptions struct {
	output string
}

func checkExportOptions(curveadm *cli.CurveAdm, options exportOptions) error {
	if !strings.HasPrefix(options.output, "/") {
		return errno.ERR_EXPORT_TOOLSV2_CONF_REQUIRE_ABSOLUTE_PATH.
			F("/path/to/curve.yaml: %s", options.output)
	}
	return nil
}

func NewExportCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options exportOptions

	cmd := &cobra.Command{
		Use:   "export [OPTIONS]",
		Short: "Export curve.yaml",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkExportOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExport(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.output, "path", "p", path.Join(curveadm.PluginDir(), "curve.yaml"), "Path where the exported YAML is stored")
	return cmd
}

func genExportPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options exportOptions) (*playbook.Playbook, error) {

	steps := GET_EXPORT_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs[:1],
			Options: map[string]interface{}{
				comm.KEY_TOOLSV2_CONF_PATH: options.output,
			},
		})
	}

	return pb, nil
}

func runExport(curveadm *cli.CurveAdm, options exportOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) generate get export playbook
	pb, err := genExportPlaybook(curveadm, dcs, options)
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
	curveadm.WriteOutln(color.GreenString("Export curve.yaml to %s success ^_^"), options.output)
	return err
}

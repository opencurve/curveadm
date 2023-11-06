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
 * Created Date: 2023-12-25
 * Author: Xinyu Zhuo (0fatal)
 */

package copy

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
	COPY_TOOL_PLAYBOOK_STEPS = []int{
		playbook.COPY_TOOL,
	}
)

type copyOptions struct {
	host     string
	path     string
	confPath string
}

func NewCopyTool2Command(curveadm *cli.CurveAdm) *cobra.Command {
	var options copyOptions

	cmd := &cobra.Command{
		Use:   "tool [OPTIONS]",
		Short: "Copy tool v2 on the specified host",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCopyTool(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")
	flags.StringVar(&options.path, "path", "/usr/local/bin/curve", "Specify target copy path of tool v2")
	flags.StringVar(&options.confPath, "confPath", "/etc/curve/curve.yaml", "Specify target config path of tool v2")

	return cmd
}

func genCopyToolPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options copyOptions,
) (*playbook.Playbook, error) {
	configs := curveadm.FilterDeployConfig(dcs, topology.FilterOption{Id: "*", Role: "*", Host: options.host})[:1]
	if len(configs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}
	steps := COPY_TOOL_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: configs,
			Options: map[string]interface{}{
				comm.KEY_COPY_PATH:      options.path,
				comm.KEY_COPY_CONF_PATH: options.confPath,
			},
		})
	}
	return pb, nil
}

func runCopyTool(curveadm *cli.CurveAdm, options copyOptions) error {
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	pb, err := genCopyToolPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	err = pb.Run()
	if err != nil {
		return err
	}

	curveadm.WriteOutln(color.GreenString("Copy %s to %s success."),
		"curve tool v2", options.host)
	return nil
}

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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type statusOptions struct {
	id     string
	role   string
	host   string
	vebose bool
}

func NewStatusCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options statusOptions

	cmd := &cobra.Command{
		Use:   "status [OPTIONS]",
		Short: "Display service status",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.id, "id", "", "*", "Specify service id")
	flags.StringVarP(&options.role, "role", "", "*", "Specify service role")
	flags.StringVarP(&options.host, "host", "", "*", "Specify service host")
	flags.BoolVarP(&options.vebose, "verbose", "v", false, "Verbose output for status")

	return cmd
}

func getClusterMdsAddr(dcs []*configure.DeployConfig) string {
	if len(dcs) == 0 {
		return "-"
	} else if value, err := dcs[0].GetVariables().Get("cluster_mds_addr"); err == nil {
		return value
	}
	return "-"
}

func runStatus(curveadm *cli.CurveAdm, options statusOptions) error {
	dcs, err := configure.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	dcs = configure.FilterDeployConfig(dcs, configure.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})

	if err := tasks.ExecParallelTasks(tasks.GET_SERVICE_STATUS, curveadm, dcs); err != nil {
		return curveadm.NewPromptError(err, "")
	}

	// display status
	statuses := []tasks.ServiceStatus{}
	m := curveadm.MemStorage().Map
	for _, v := range m {
		status := v.(tasks.ServiceStatus)
		statuses = append(statuses, status)
	}

	output := tui.FormatStatus(statuses, options.vebose)
	curveadm.WriteOut("\n")
	curveadm.WriteOut("cluster name    : %s\n", curveadm.ClusterName())
	curveadm.WriteOut("cluster mds addr: %s\n", getClusterMdsAddr(dcs))
	curveadm.WriteOut("\n")
	curveadm.WriteOut("%s", output)
	return nil
}

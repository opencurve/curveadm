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

// __SIGN_BY_WINE93__

package command

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/service"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	GET_STATUS_PLAYBOOK_STEPS = []int{
		playbook.INIT_SERVIE_STATUS,
		playbook.GET_SERVICE_STATUS,
	}
)

type statusOptions struct {
	id            string
	role          string
	host          string
	verbose       bool
	showInstances bool
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
	flags.StringVar(&options.id, "id", "*", "Specify service id")
	flags.StringVar(&options.role, "role", "*", "Specify service role")
	flags.StringVar(&options.host, "host", "*", "Specify service host")
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output for status")
	flags.BoolVarP(&options.showInstances, "show-instances", "s", false, "Display service num")

	return cmd
}

func getClusterMdsAddr(dcs []*topology.DeployConfig) string {
	value, err := dcs[0].GetVariables().Get("cluster_mds_addr")
	if err != nil {
		return "-"
	}
	return value
}

func getClusterMdsLeader(statuses []task.ServiceStatus) string {
	leaders := []string{}
	for _, status := range statuses {
		if !status.IsLeader {
			continue
		}
		dc := status.Config
		leader := fmt.Sprintf("%s:%d / %s",
			dc.GetListenIp(), dc.GetListenPort(), status.Id)
		leaders = append(leaders, leader)
	}
	if len(leaders) > 0 {
		return strings.Join(leaders, ", ")
	}
	return color.RedString("<no leader>")
}

func displayStatus(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, options statusOptions) {
	statuses := []task.ServiceStatus{}
	value := curveadm.MemStorage().Get(comm.KEY_ALL_SERVICE_STATUS)
	if value != nil {
		m := value.(map[string]task.ServiceStatus)
		for _, status := range m {
			statuses = append(statuses, status)
		}
	}

	output := tui.FormatStatus(statuses, options.verbose, options.showInstances)
	curveadm.WriteOutln("")
	curveadm.WriteOutln("cluster name      : %s", curveadm.ClusterName())
	curveadm.WriteOutln("cluster kind      : %s", dcs[0].GetKind())
	curveadm.WriteOutln("cluster mds addr  : %s", getClusterMdsAddr(dcs))
	curveadm.WriteOutln("cluster mds leader: %s", getClusterMdsLeader(statuses))
	curveadm.WriteOutln("")
	curveadm.WriteOut("%s", output)
}

func genStatusPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options statusOptions) (*playbook.Playbook, error) {
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := GET_STATUS_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
			ExecOptions: playbook.ExecOptions{
				Concurrency:   100,
				SilentSubBar:  true,
				SilentMainBar: step == playbook.INIT_SERVIE_STATUS,
				SkipError:     true,
			},
		})
	}
	return pb, nil
}

func runStatus(curveadm *cli.CurveAdm, options statusOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) generate get status playbook
	pb, err := genStatusPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()

	// 4) display service status
	displayStatus(curveadm, dcs, options)
	return err
}

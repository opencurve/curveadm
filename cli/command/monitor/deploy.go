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
* Project: Curveadm
* Created Date: 2023-04-17
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tasks"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	DEPLOY_EXAMPLE = `Examples:
	$ curveadm monitor deploy -c monitor.yaml    # deploy monitor for current cluster`
)

var (
	MONITOR_DEPLOY_STEPS = []int{
		playbook.PULL_MONITOR_IMAGE,
		playbook.CREATE_MONITOR_CONTAINER,
		playbook.SYNC_MONITOR_CONFIG,
		playbook.CLEAN_CONFIG_CONTAINER,
		playbook.START_MONITOR_SERVICE,
	}
)

type deployOptions struct {
	filename string
}

/*
 * Deploy Steps:
 *   1) pull images(curvebs, node_exporter, prometheus, grafana)
 *   2) create container
 *   3) sync config
 *   4) start container
 *     4.1) start node_exporter container
 *     4.2) start prometheus container
 *     4.3) start grafana container
 */
func NewDeployCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options deployOptions

	cmd := &cobra.Command{
		Use:     "deploy [OPTIONS]",
		Short:   "Deploy monitor for current cluster",
		Args:    cliutil.NoArgs,
		Example: DEPLOY_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "monitor.yaml", "Specify monitor configuration file")
	return cmd
}

func genDeployPlaybook(curveadm *cli.CurveAdm,
	mcs []*configure.MonitorConfig) (*playbook.Playbook, error) {
	steps := MONITOR_DEPLOY_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		if step == playbook.CLEAN_CONFIG_CONTAINER {
			pb.AddStep(&playbook.PlaybookStep{
				Type:    step,
				Configs: mcs,
				ExecOptions: tasks.ExecOptions{
					SilentMainBar: true,
					SilentSubBar:  true,
				},
			})
			continue
		}
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: mcs,
		})
	}
	return pb, nil
}

func parseTopology(curveadm *cli.CurveAdm) ([]string, []string, []*topology.DeployConfig, error) {
	dcs, err := curveadm.ParseTopology()
	if err != nil || len(dcs) == 0 {
		return nil, nil, nil, err
	}
	hosts := []string{}
	hostIps := []string{}
	thostMap := make(map[string]bool)
	thostIpMap := make(map[string]bool)
	for _, dc := range dcs {
		thostMap[dc.GetHost()] = true
		thostIpMap[dc.GetListenIp()] = true
	}
	for key := range thostMap {
		hosts = append(hosts, key)
	}
	for key := range thostIpMap {
		hostIps = append(hostIps, key)
	}
	return hosts, hostIps, dcs, nil
}

func runDeploy(curveadm *cli.CurveAdm, options deployOptions) error {
	// 1) parse cluster topology and get services' hosts
	hosts, hostIps, dcs, err := parseTopology(curveadm)
	if err != nil {
		return err
	}

	// 2) parse monitor configure
	mcs, err := configure.ParseMonitorConfig(curveadm, options.filename, "", hosts, hostIps, dcs)
	if err != nil {
		return err
	}

	// 3) save monitor data
	data, err := utils.ReadFile(options.filename)
	if err != nil {
		return errno.ERR_READ_MONITOR_FILE_FAILED.E(err)
	}
	err = curveadm.Storage().ReplaceMonitor(storage.Monitor{
		ClusterId: curveadm.ClusterId(),
		Monitor:   data,
	})
	if err != nil {
		return errno.ERR_REPLACE_MONITOR_FAILED.E(err)
	}

	// 4) generate deploy playbook
	pb, err := genDeployPlaybook(curveadm, mcs)
	if err != nil {
		return err
	}

	// 5) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 6) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Deploy monitor success ^_^"))
	return nil
}

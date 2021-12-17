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
	"fmt"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	PULL_IMAGE int = iota
	CREATE_CONTAINER
	SYNC_CONFIG
	START_ETCD
	START_MDS
	START_METASEREVR
	CREATE_TOPOLOGY
)

type deployOptions struct{}

func NewDeployCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options deployOptions

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy cluster",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func filterDeployConfig(dcs []*configure.DeployConfig, role string) []*configure.DeployConfig {
	options := configure.FilterOption{Id: "*", Role: role, Host: "*"}
	return configure.FilterDeployConfig(dcs, options)
}

func displayTitle(curveadm *cli.CurveAdm, dcs []*configure.DeployConfig) {
	netcd := 0
	nmds := 0
	nmetaserver := 0
	for _, dc := range dcs {
		if dc.GetRole() == configure.ROLE_ETCD {
			netcd += 1
		} else if dc.GetRole() == configure.ROLE_MDS {
			nmds += 1
		} else if dc.GetRole() == configure.ROLE_METASERVER {
			nmetaserver += 1
		}
	}

	curveadm.WriteOut("Cluster Name    : %s\n", curveadm.ClusterName())
	curveadm.WriteOut("Cluster Services: etcd*%d, mds*%d, metaserver*%d\n", netcd, nmds, nmetaserver)
	curveadm.WriteOut("\n")
}

/*
 * Deploy Steps:
 *   1) pull image
 *   2) create container
 *   3) sync config
 *   4) start container
 *     4.1) start etcd container
 *     4.2) start mds container
 *     4.3) create topology
 *     4.4) start metaserver
 */
func runDeploy(curveadm *cli.CurveAdm, options deployOptions) error {
	deployConfigs, err := configure.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	// display title
	displayTitle(curveadm, deployConfigs)

	// exec task one by one
	taskSeq := []int{
		PULL_IMAGE,
		CREATE_CONTAINER,
		SYNC_CONFIG,
		START_ETCD,
		START_MDS,
		CREATE_TOPOLOGY,
		START_METASEREVR,
	}
	for _, v := range taskSeq {
		taskType := tasks.UNKNOWN
		dcs := deployConfigs
		switch v {
		case PULL_IMAGE:
			taskType = tasks.PULL_IMAGE
		case CREATE_CONTAINER:
			taskType = tasks.CREATE_CONTAINER
		case SYNC_CONFIG:
			taskType = tasks.SYNC_CONFIG
		case START_ETCD:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(dcs, configure.ROLE_ETCD)
		case START_MDS:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(dcs, configure.ROLE_MDS)
		case CREATE_TOPOLOGY:
			taskType = tasks.CREATE_TOPOLOGY
			dcs = filterDeployConfig(dcs, configure.ROLE_MDS)[:1]
		case START_METASEREVR:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(dcs, configure.ROLE_METASERVER)
		}

		if len(dcs) == 0 {
			return fmt.Errorf("there is no service specified in topology, " +
				"please use 'curveadm config commit' to update topology")
		} else if err := tasks.ExecTasks(taskType, curveadm, dcs); err != nil {
			return curveadm.NewPromptError(err, "")
		} else {
			curveadm.WriteOut("\n")
		}
	}

	curveadm.WriteOut(color.GreenString("Cluster '%s' successfully deployed :)\n"), curveadm.ClusterName())
	return nil
}

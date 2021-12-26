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
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errors"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	// task type
	PULL_IMAGE int = iota
	CREATE_CONTAINER
	SYNC_CONFIG
	START_ETCD
	START_MDS
	START_METASEREVR
	START_CHUNKSERVER
	CREATE_PHYSICAL_POOL
	CREATE_LOGICAL_POOL
)

var (
	CURVEBS_STEPS = []int{
		PULL_IMAGE,
		CREATE_CONTAINER,
		SYNC_CONFIG,
		START_ETCD,
		START_MDS,
		CREATE_PHYSICAL_POOL,
		START_CHUNKSERVER,
		CREATE_LOGICAL_POOL,
	}

	CURVEFS_STEPS = []int{
		PULL_IMAGE,
		CREATE_CONTAINER,
		SYNC_CONFIG,
		START_ETCD,
		START_MDS,
		START_METASEREVR,
		CREATE_LOGICAL_POOL,
	}
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

func filterDeployConfig(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, role string) []*topology.DeployConfig {
	options := topology.FilterOption{Id: "*", Role: role, Host: "*"}
	return curveadm.FilterDeployConfig(dcs, options)
}

func displayTitle(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) {
	netcd := 0
	nmds := 0
	nmetaserver := 0
	nchunkserevr := 0
	for _, dc := range dcs {
		role := dc.GetRole()
		switch role {
		case topology.ROLE_ETCD:
			netcd += 1
		case topology.ROLE_MDS:
			nmds += 1
		case topology.ROLE_CHUNKSERVER:
			nchunkserevr += 1
		case topology.ROLE_METASERVER:
			nmetaserver += 1
		}
	}

	var serviceStats string
	kind := dcs[0].GetKind()
	if kind == topology.KIND_CURVEBS {
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, chunkserver*%d", netcd, nmds, nchunkserevr)
	} else { // KIND_CURVEFS
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, metaserver*%d", netcd, nmds, nmetaserver)
	}

	curveadm.WriteOut("Cluster Name    : %s\n", curveadm.ClusterName())
	curveadm.WriteOut("Cluster Kind    : %s\n", kind)
	curveadm.WriteOut("Cluster Services: %s\n", serviceStats)
	curveadm.WriteOut("\n")
}

func execDeployTask(curveadm *cli.CurveAdm, deployConfigs []*topology.DeployConfig, steps []int) error {
	for _, step := range steps {
		taskType := tasks.UNKNOWN
		dcs := deployConfigs
		switch step {
		case PULL_IMAGE:
			taskType = tasks.PULL_IMAGE
		case CREATE_CONTAINER:
			taskType = tasks.CREATE_CONTAINER
		case SYNC_CONFIG:
			taskType = tasks.SYNC_CONFIG
		case START_ETCD:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case START_MDS:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
		case START_CHUNKSERVER:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_CHUNKSERVER)
		case START_METASEREVR:
			taskType = tasks.START_SERVICE
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_METASERVER)
		case CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_PHYSICAL_POOL)
			taskType = tasks.CREATE_POOL
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		case CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_LOGICAL_POOL)
			taskType = tasks.CREATE_POOL
			dcs = filterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		}

		if len(dcs) == 0 {
			return errors.ERR_CONFIGURE_NO_SERVICE
		}

		err := tasks.ExecTasks(taskType, curveadm, dcs)
		if err != nil {
			return err
		}

		curveadm.WriteOut("\n")
	}
	return nil
}

/*
 * Deploy Steps:
 *   1) pull image
 *   2) create container
 *   3) sync config
 *   4) start container
 *     4.1) start etcd container
 *     4.2) start mds container
 *     4.4) start chunkserver(curvebs) / metaserver(curvefs)
 *   5) create topology
 */
func runDeploy(curveadm *cli.CurveAdm, options deployOptions) error {
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	} else if len(dcs) == 0 {
		return errors.ERR_CONFIGURE_NO_SERVICE
	}

	// display title
	displayTitle(curveadm, dcs)

	// exec deploy task one by one
	kind := dcs[0].GetKind()
	if kind == topology.KIND_CURVEBS {
		err = execDeployTask(curveadm, dcs, CURVEBS_STEPS)
	} else {
		err = execDeployTask(curveadm, dcs, CURVEFS_STEPS)
	}

	if err == nil {
		curveadm.WriteOut(color.GreenString("Cluster '%s' successfully deployed ^_^.\n"), curveadm.ClusterName())
	} else if err != nil {
		return curveadm.NewPromptError(err, "")
	}

	return err
}

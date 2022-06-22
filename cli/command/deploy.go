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
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errors"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	CURVEBS_DEPLOY_STEPS = []int{
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_ETCD,
		comm.START_MDS,
		comm.CREATE_PHYSICAL_POOL,
		comm.START_CHUNKSERVER,
		comm.CREATE_LOGICAL_POOL,
		comm.START_SNAPSHOTCLONE,
		comm.BALANCE_LEADER,
	}

	CURVEFS_DEPLOY_STEPS = []int{
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_ETCD,
		comm.START_MDS,
		comm.CREATE_LOGICAL_POOL,
		comm.START_METASEREVR,
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

func genDeploySteps(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) []comm.DeployStep {
	kind := dcs[0].GetKind()
	steps := CURVEFS_DEPLOY_STEPS
	if kind == topology.KIND_CURVEBS {
		steps = CURVEBS_DEPLOY_STEPS
	}

	dss := []comm.DeployStep{}
	for _, step := range steps {
		ds := comm.DeployStep{Type: step}
		switch step {
		case comm.START_ETCD:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case comm.START_MDS:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
		case comm.START_CHUNKSERVER:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_CHUNKSERVER)
		case comm.START_SNAPSHOTCLONE:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_SNAPSHOTCLONE)
		case comm.START_METASEREVR:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_METASERVER)
		case comm.CREATE_PHYSICAL_POOL:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		case comm.CREATE_LOGICAL_POOL:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		case comm.BALANCE_LEADER:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		default:
			ds.DeployConfigs = dcs
		}
		dss = append(dss, ds)
	}
	return dss
}

func displayDeployTitle(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) {
	netcd := 0
	nmds := 0
	nchunkserevr := 0
	nsnapshotclone := 0
	nmetaserver := 0
	for _, dc := range dcs {
		role := dc.GetRole()
		switch role {
		case topology.ROLE_ETCD:
			netcd += 1
		case topology.ROLE_MDS:
			nmds += 1
		case topology.ROLE_CHUNKSERVER:
			nchunkserevr += 1
		case topology.ROLE_SNAPSHOTCLONE:
			nsnapshotclone += 1
		case topology.ROLE_METASERVER:
			nmetaserver += 1
		}
	}

	var serviceStats string
	kind := dcs[0].GetKind()
	if kind == topology.KIND_CURVEBS {
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, chunkserver*%d, snapshotclone*%d",
			netcd, nmds, nchunkserevr, nsnapshotclone)
	} else { // KIND_CURVEFS
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, metaserver*%d",
			netcd, nmds, nmetaserver)
	}

	curveadm.WriteOut("Cluster Name    : %s\n", curveadm.ClusterName())
	curveadm.WriteOut("Cluster Kind    : %s\n", kind)
	curveadm.WriteOut("Cluster Services: %s\n", serviceStats)
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
 *     4.3) create physical pool(curvebs)
 *     4.3) start chunkserver(curvebs) / metaserver(curvefs) container
 *     4.4) start snapshotserver(curvebs) container
 *   5) create logical pool
 *   6) balance leader rapidly
 */
func runDeploy(curveadm *cli.CurveAdm, options deployOptions) error {
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	} else if len(dcs) == 0 {
		return errors.ERR_CONFIGURE_NO_SERVICE
	}

	// display title
	displayDeployTitle(curveadm, dcs)

	// exec deploy task one by one
	steps := genDeploySteps(curveadm, dcs)
	err = comm.ExecDeploy(curveadm, steps)
	if err == nil {
		curveadm.WriteOut(color.GreenString("Cluster '%s' successfully deployed ^_^.\n"), curveadm.ClusterName())
	} else if err != nil {
		return curveadm.NewPromptError(err, "")
	}

	return err
}

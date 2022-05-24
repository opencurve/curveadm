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
 * Created Date: 2022-05-20
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"errors"
	"fmt"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// etcd
	SCALE_OUT_ETCD_STEPS = []int{
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_ETCD,
	}

	// mds
	SCALE_OUT_MDS_STEPS = []int{
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_MDS,
	}

	// snapshotclone (curvebs)
	SCALE_OUT_SNAPSHOTCLONE_STEPS = []int{
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_SNAPSHOTCLONE,
	}

	// chunkserevr (curvebs)
	SCALE_OUT_CHUNKSERVER_STEPS = []int{
		comm.BACKUP_ETCD_DATA,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.CREATE_PHYSICAL_POOL,
		comm.START_CHUNKSERVER,
		comm.CREATE_LOGICAL_POOL,
	}

	// metaserver (curvefs)
	SCALE_OUT_METASERVER_STEPS = []int{
		comm.BACKUP_ETCD_DATA,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_METASEREVR,
		comm.CREATE_LOGICAL_POOL,
	}
)

type scaleOutOptions struct {
	filename string
}

func NewScaleOutCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options scaleOutOptions

	cmd := &cobra.Command{
		Use:   "scale-out TOPOLOGY",
		Short: "Scale out cluster",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filename = args[0]
			return runScaleOut(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func checkScaleOutTopology(oldData, newData string) (error, bool) {
	diffs, err := comm.DiffTopology(oldData, newData)
	if errors.Is(err, comm.ERR_EMPTY_TOPOLOGY) {
		return fmt.Errorf("cluster topology is empty"), false
	} else if errors.Is(err, comm.ERR_NO_SERVICE) {
		return fmt.Errorf("you can't scale out empty cluster"), false
	} else if err != nil {
		return err, false
	}

	dcs4add, dcs4del, dcs4change := comm.ParseDiff(diffs)
	if len(dcs4del) != 0 {
		return fmt.Errorf("you can't delete service in scale-out"), false
	} else if len(dcs4add) == 0 {
		return fmt.Errorf("no new services for scale out"), false
	} else if !comm.IsSameRole(dcs4add) {
		return fmt.Errorf("you can only scale out same role services every time"), false
	}
	return nil, len(dcs4change) != 0
}

func genScaleOutSteps(curveadm *cli.CurveAdm, oldData, newData string) ([]comm.DeployStep, error) {
	diffs, _ := comm.DiffTopology(oldData, newData) // ignore error
	dcs4add, _, _ := comm.ParseDiff(diffs)
	dcs, _ := topology.ParseTopology(curveadm.ClusterTopologyData())

	var steps []int
	role := dcs4add[0].GetRole()
	switch role {
	case topology.ROLE_ETCD:
		steps = SCALE_OUT_ETCD_STEPS
	case topology.ROLE_MDS:
		steps = SCALE_OUT_MDS_STEPS
	case topology.ROLE_SNAPSHOTCLONE:
		steps = SCALE_OUT_SNAPSHOTCLONE_STEPS
	case topology.ROLE_CHUNKSERVER:
		steps = SCALE_OUT_CHUNKSERVER_STEPS
	case topology.ROLE_METASERVER:
		steps = SCALE_OUT_METASERVER_STEPS
	default:
		return nil, fmt.Errorf("unknown role '%s'", role)
	}

	dss := []comm.DeployStep{}
	for _, step := range steps {
		ds := comm.DeployStep{Type: step}
		switch step {
		case comm.BACKUP_ETCD_DATA:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case comm.CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_SCALE_OUT_CLUSTER, dcs4add)
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		case comm.CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_SCALE_OUT_CLUSTER, dcs4add)
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		default:
			ds.DeployConfigs = dcs4add
		}
		dss = append(dss, ds)
	}
	return dss, nil
}

func displayScaleOutTitle(curveadm *cli.CurveAdm) {
	dcs := curveadm.MemStorage().Get(task.KEY_SCALE_OUT_CLUSTER).([]*topology.DeployConfig)
	role := dcs[0].GetRole()
	curveadm.WriteOutln("NOTICE: cluster '%s' is about to scale out:", curveadm.ClusterName())
	curveadm.WriteOutln("  - Scale-out services: %s*%d", role, len(dcs))
}

func runScaleOut(curveadm *cli.CurveAdm, options scaleOutOptions) error {
	// 1. show topology difference
	oldData := curveadm.ClusterTopologyData()
	newData, err := utils.ReadFile(options.filename)
	if err != nil {
		return err
	}
	curveadm.WriteOutln(utils.Diff(oldData, newData))

	// 2. validate topology difference
	err, warning := checkScaleOutTopology(oldData, newData)
	if err != nil {
		return err
	}

	// 3. generate scale-out deploy steps
	steps, err := genScaleOutSteps(curveadm, oldData, newData)
	if err != nil {
		return err
	}

	// 4. execute scale-out steps one by one
	displayScaleOutTitle(curveadm)
	if pass := tui.ConfirmYes(tui.PromptScaleOut(warning)); !pass {
		curveadm.WriteOutln(tui.PromptCancelOpetation("scale-out"))
		return nil
	} else if err := comm.ExecDeploy(curveadm, steps); err != nil {
		return curveadm.NewPromptError(err, "")
	} else if err := curveadm.Storage().SetClusterTopology(curveadm.ClusterId(), newData); err != nil {
		return err
	}
	curveadm.WriteOut(color.GreenString("Cluster '%s' successfully scaled out\n"), curveadm.ClusterName())
	return nil
}

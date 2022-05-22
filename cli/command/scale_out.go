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
		Use:   "scale-out",
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

func checkDiff4ScaleOut(diffs []topology.TopologyDiff, err4diff error) (dcs []*topology.DeployConfig, err error, warning bool) {
	if errors.Is(err4diff, comm.ERR_EMPTY_TOPOLOGY) {
		err = fmt.Errorf("cluster topology is empty")
		return
	} else if errors.Is(err4diff, comm.ERR_NO_SERVICE) {
		err = fmt.Errorf("you can't scale out empty cluster")
		return
	}

	for _, diff := range diffs {
		diffType := diff.DiffType
		if diffType == topology.DIFF_ADD {
			dcs = append(dcs, diff.DeployConfig)
		} else if diffType == topology.DIFF_DELETE {
			err = fmt.Errorf("you can't delete service in scale-out")
			return
		} else if diffType == topology.DIFF_CHANGE {
			warning = true
		}
	}

	if len(dcs) == 0 {
		err = fmt.Errorf("No new services for scale out")
		return
	}

	role := dcs[0].GetRole()
	for _, dc := range dcs {
		if dc.GetRole() != role {
			err = fmt.Errorf("You can only scale out same role services every time")
			return
		}
	}

	return
}

func steps2scale(curveadm *cli.CurveAdm, dcs, dcs2scale []*topology.DeployConfig) ([]comm.Step, error) {
	var steps []int
	role := dcs2scale[0].GetRole()
	switch role {
	case topology.ROLE_ETCD:
		steps = SCALE_OUT_ETCD_STEPS
	case topology.ROLE_MDS:
		steps = SCALE_OUT_MDS_STEPS
	case topology.ROLE_CHUNKSERVER:
		steps = SCALE_OUT_CHUNKSERVER_STEPS
	case topology.ROLE_METASERVER:
		steps = SCALE_OUT_METASERVER_STEPS
	default:
		return nil, fmt.Errorf("unknown role '%s'", role)
	}

	ss := []comm.Step{}
	for _, step := range steps {
		s := comm.Step{Type: step}
		switch step {
		case comm.BACKUP_ETCD_DATA:
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case comm.CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_SCALE_OUT, dcs2scale)
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		case comm.CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_SCALE_OUT, dcs2scale)
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)[:1]
		default:
			s.DeployConfigs = dcs2scale
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func runScaleOut(curveadm *cli.CurveAdm, options scaleOutOptions) error {
	oldData := curveadm.ClusterTopologyData()
	newData, err := utils.ReadFile(options.filename)
	if err != nil {
		return err
	}

	curveadm.Out().Write([]byte(utils.Diff(oldData, newData)))

	diffs, err := comm.DiffTopology(oldData, newData)
	if err != nil {
		return err
	}

	dcs2scale, err, warning := checkDiff4ScaleOut(diffs, err)
	if err != nil {
		return err
	}

	dcs, err := topology.ParseTopology(oldData)
	if err != nil {
		return err
	}

	steps, err := steps2scale(curveadm, dcs, dcs2scale)
	if err != nil {
		return err
	}

	if pass := tui.ConfirmYes(tui.PromptScaleOut(warning)); !pass {
		curveadm.WriteOut("scale-out canceled")
		return nil
	}

	err = comm.ExecDeploy(curveadm, steps)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	} else if err := curveadm.Storage().SetClusterTopology(curveadm.ClusterId(), newData); err != nil {
		return err
	}
	curveadm.WriteOut(color.GreenString("Cluster '%s' successfully scaled out\n"), curveadm.ClusterName())
	return nil
}

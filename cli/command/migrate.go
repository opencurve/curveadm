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
	"github.com/opencurve/curveadm/internal/configure/pool"
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// etcd
	MIGRATE_ETCD_STEPS = []int{
		comm.STOP_SERVICE,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_ETCD,
	}

	// mds
	MIGRATE_MDS_STEPS = []int{
		comm.STOP_SERVICE,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_MDS,
	}

	// chunkserevr (curvebs)
	MIGRATE_CHUNKSERVER_STEPS = []int{
		comm.BACKUP_ETCD_DATA,
		comm.STOP_SERVICE,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.CREATE_PHYSICAL_POOL,
		comm.START_CHUNKSERVER,
		comm.CREATE_LOGICAL_POOL,
	}

	// metaserver (curvefs)
	MIGRATE_METASERVER_STEPS = []int{
		comm.BACKUP_ETCD_DATA,
		comm.STOP_SERVICE,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_METASEREVR,
		comm.CREATE_LOGICAL_POOL,
	}
)

type migrateOptions struct {
	filename string
}

func NewMigrateCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options migrateOptions

	cmd := &cobra.Command{
		Use:   "migrate TOPOLOGY",
		Short: "Migrate services",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filename = args[0]
			return runMigrate(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func isSameRole(dcs []*topology.DeployConfig) bool {
	role := dcs[0].GetRole()
	for _, dc := range dcs {
		if dc.GetRole() != role {
			return false
		}
	}
	return true
}

func checkDiff4Migrate(diffs []topology.TopologyDiff, err4diff error) (migrates []*pool.MigrateServer, err error, warning bool) {
	if errors.Is(err4diff, comm.ERR_EMPTY_TOPOLOGY) {
		err = fmt.Errorf("cluster topology is empty")
		return
	} else if errors.Is(err4diff, comm.ERR_NO_SERVICE) {
		err = fmt.Errorf("you can't migrate empty cluster")
		return
	}

	dcs2del, dcs2add := []*topology.DeployConfig{}, []*topology.DeployConfig{}
	for _, diff := range diffs {
		diffType := diff.DiffType
		if diffType == topology.DIFF_ADD {
			dcs2add = append(dcs2add, diff.DeployConfig)
		} else if diffType == topology.DIFF_DELETE {
			dcs2del = append(dcs2del, diff.DeployConfig)
		} else if diffType == topology.DIFF_CHANGE {
			warning = true
		}
	}

	if len(dcs2del) != len(dcs2add) {
		err = fmt.Errorf("You can only migrate same host service ervery time")
		return
	} else if len(dcs2del) == 0 {
		err = fmt.Errorf("No service for migrating")
		return
	} else if !isSameRole(dcs2del) || !isSameRole(dcs2add) ||
		dcs2del[0].GetRole() != dcs2add[0].GetRole() {
		err = fmt.Errorf("You can only migrate same role services every time")
		return
	} else if len(dcs2del) != dcs2del[0].GetReplica() {
		err = fmt.Errorf("You can only migrate whole host services every time")
		return
	}

	for i := 0; i < len(dcs2del); i++ {
		migrates = append(migrates, &pool.MigrateServer{dcs2del[i], dcs2add[i]})
	}
	return
}

func selectActiveMDS(dcs, dcs2del, dcs2add []*topology.DeployConfig) []*topology.DeployConfig {
	m := map[string]bool{}
	for _, dc := range dcs2del {
		m[dc.GetName()] = true
	}

	out := []*topology.DeployConfig{}
	for _, dc := range dcs {
		if m[dc.GetName()] {
			continue
		}
		out = append(out, dc)
	}

	if len(out) > 0 {
		return out
	}
	return dcs2add
}

func steps2migrate(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, migrates []*pool.MigrateServer) ([]comm.Step, error) {
	dcs2del, dcs2add := []*topology.DeployConfig{}, []*topology.DeployConfig{}
	for _, migrate := range migrates {
		dcs2del = append(dcs2del, migrate.From)
		dcs2add = append(dcs2add, migrate.To)
	}

	var steps []int
	role := dcs2del[0].GetRole()
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
		case comm.STOP_SERVICE:
			s.DeployConfigs = dcs2del
		case comm.CLEAN_SERVICE_CONTAINER:
			s.DeployConfigs = dcs2del
		case comm.BACKUP_ETCD_DATA:
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case comm.CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_MIGRATE_SERVERS, migrates)
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
			s.DeployConfigs = selectActiveMDS(s.DeployConfigs, dcs2del, dcs2add)[:1]
		case comm.CREATE_LOGICAL_POOL:
			s.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
			s.DeployConfigs = selectActiveMDS(s.DeployConfigs, dcs2del, dcs2add)[:1]
		default:
			s.DeployConfigs = dcs2add
		}
		ss = append(ss, s)
	}
	return ss, nil
}

func runMigrate(curveadm *cli.CurveAdm, options migrateOptions) error {
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

	migrates, err, warning := checkDiff4Migrate(diffs, err)
	if err != nil {
		return err
	}

	dcs, err := topology.ParseTopology(oldData)
	if err != nil {
		return err
	}

	steps, err := steps2migrate(curveadm, dcs, migrates)
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
	curveadm.WriteOut(color.GreenString("Cluster '%s' successfully migrated\n"), curveadm.ClusterName())
	return nil
}

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
		comm.STOP_ETCD,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_ETCD,
	}

	// mds
	MIGRATE_MDS_STEPS = []int{
		comm.STOP_MDS,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_MDS,
	}

	// snapshotclone
	MIGRATE_SNAPSHOTCLONE_STEPS = []int{
		comm.STOP_SNAPSHOTCLONE,
		comm.CLEAN_SERVICE_CONTAINER,
		comm.PULL_IMAGE,
		comm.CREATE_CONTAINER,
		comm.SYNC_CONFIG,
		comm.START_SNAPSHOTCLONE,
	}

	// chunkserevr (curvebs)
	MIGRATE_CHUNKSERVER_STEPS = []int{
		comm.BACKUP_ETCD_DATA,
		comm.STOP_CHUNKSERVER,
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
		comm.STOP_METASEREVR,
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

func checkMigrateTopology(oldData, newData string) (error, bool) {
	diffs, err := comm.DiffTopology(oldData, newData)
	if errors.Is(err, comm.ERR_EMPTY_TOPOLOGY) {
		return fmt.Errorf("cluster topology is empty"), false
	} else if errors.Is(err, comm.ERR_NO_SERVICE) {
		return fmt.Errorf("you can't scale out empty cluster"), false
	} else if err != nil {
		return err, false
	}

	dcs4add, dcs4del, dcs4change := comm.ParseDiff(diffs)
	if len(dcs4del) != len(dcs4add) {
		return fmt.Errorf("you can only migrate same host services ervery time"), false
	} else if len(dcs4add) == 0 {
		return fmt.Errorf("no service for migrating"), false
	} else if !comm.IsSameRole(dcs4add) || !comm.IsSameRole(dcs4add) ||
		dcs4del[0].GetRole() != dcs4add[0].GetRole() {
		return fmt.Errorf("you can only migrate same role services every time"), false
	} else if len(dcs4del) != dcs4del[0].GetReplica() {
		return fmt.Errorf("you can only migrate whole host services every time"), false
	}
	return nil, len(dcs4change) != 0
}

func selectActiveMDS(dcs, dcs4del, dcs4add []*topology.DeployConfig) []*topology.DeployConfig {
	deleted := map[string]bool{}
	for _, dc := range dcs4del {
		deleted[dc.GetName()] = true
	}

	out := []*topology.DeployConfig{}
	for _, dc := range dcs {
		if deleted[dc.GetName()] { // already deleted
			continue
		}
		out = append(out, dc)
	}
	return append(out, dcs4add...)
}

func genMigrateSteps(curveadm *cli.CurveAdm, oldData, newData string) ([]comm.DeployStep, error) {
	diffs, _ := comm.DiffTopology(oldData, newData) // ignore error
	dcs4add, dcs4del, _ := comm.ParseDiff(diffs)
	dcs, _ := topology.ParseTopology(curveadm.ClusterTopologyData())
	comm.SortDeployConfigs(dcs4add)
	comm.SortDeployConfigs(dcs4del)

	migrates := []*pool.MigrateServer{}
	for i := 0; i < len(dcs4del); i++ {
		migrates = append(migrates, &pool.MigrateServer{dcs4del[i], dcs4add[i]})
	}

	var steps []int
	role := dcs4del[0].GetRole()
	switch role {
	case topology.ROLE_ETCD:
		steps = MIGRATE_ETCD_STEPS
	case topology.ROLE_MDS:
		steps = MIGRATE_MDS_STEPS
	case topology.ROLE_SNAPSHOTCLONE:
		steps = MIGRATE_SNAPSHOTCLONE_STEPS
	case topology.ROLE_CHUNKSERVER:
		steps = MIGRATE_CHUNKSERVER_STEPS
	case topology.ROLE_METASERVER:
		steps = MIGRATE_METASERVER_STEPS
	default:
		return nil, fmt.Errorf("unknown role '%s'", role)
	}

	dss := []comm.DeployStep{}
	for _, step := range steps {
		ds := comm.DeployStep{Type: step}
		switch step {
		case comm.STOP_ETCD, comm.STOP_MDS, comm.STOP_SNAPSHOTCLONE, comm.STOP_CHUNKSERVER, comm.STOP_METASEREVR:
			ds.DeployConfigs = dcs4del
		case comm.CLEAN_SERVICE_CONTAINER:
			ds.DeployConfigs = dcs4del
		case comm.BACKUP_ETCD_DATA:
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_ETCD)
		case comm.CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_MIGRATE_SERVERS, migrates)
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
			ds.DeployConfigs = selectActiveMDS(ds.DeployConfigs, dcs4del, dcs4add)[:1]
		case comm.CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_MIGRATE_SERVERS, migrates)
			ds.DeployConfigs = comm.FilterDeployConfig(curveadm, dcs, topology.ROLE_MDS)
			ds.DeployConfigs = selectActiveMDS(ds.DeployConfigs, dcs4del, dcs4add)[:1]
		default: // PULL_IMAGE, CREATE_CONATINER, SYNC_CONFIG, START_{ETCD,MDS,SNAPSHOTCLONE,CHUNKSERVER,METASERVER}
			ds.DeployConfigs = dcs4add
		}
		dss = append(dss, ds)
	}
	return dss, nil
}

func displayMigrateTitle(curveadm *cli.CurveAdm) {
	migrates := curveadm.MemStorage().Get(task.KEY_MIGRATE_SERVERS).([]*pool.MigrateServer)
	from := migrates[0].From
	to := migrates[0].To
	curveadm.WriteOutln("NOTICE: cluster '%s' is about to migrate services:", curveadm.ClusterName())
	curveadm.WriteOutln("  - Migrate services: %s*%d", from.GetRole(), len(migrates))
	curveadm.WriteOutln("  - Migrate host: from %s to %s", from.GetHost(), to.GetHost())
}

func runMigrate(curveadm *cli.CurveAdm, options migrateOptions) error {
	// 1. show topology difference
	oldData := curveadm.ClusterTopologyData()
	newData, err := utils.ReadFile(options.filename)
	if err != nil {
		return err
	}
	curveadm.WriteOutln(utils.Diff(oldData, newData))

	// 2. validate topology difference
	err, warning := checkMigrateTopology(oldData, newData)
	if err != nil {
		return err
	}

	// 3. generate scale-out deploy steps
	steps, err := genMigrateSteps(curveadm, oldData, newData)
	if err != nil {
		return err
	}

	// 4. execute scale-out steps one by one
	displayMigrateTitle(curveadm)
	if pass := tui.ConfirmYes(tui.PromptMigrate(warning)); !pass {
		curveadm.WriteOutln(tui.PromptCancelOpetation("migrate"))
		return nil
	} else if err := comm.ExecDeploy(curveadm, steps); err != nil {
		return curveadm.NewPromptError(err, "")
	} else if err := curveadm.Storage().SetClusterTopology(curveadm.ClusterId(), newData); err != nil {
		return err
	}
	curveadm.WriteOut(color.GreenString("Services successfully migrated\n"))
	return nil
}

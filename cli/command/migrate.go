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
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/common"
	tuicomm "github.com/opencurve/curveadm/internal/tui/common"

	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	MIGRATE_ETCD_STEPS = []int{
		playbook.ADD_ETCD_MEMBER,
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.AMEND_ETCD_CONFIG,
		playbook.START_ETCD,
		playbook.AMEND_MDS_CONFIG, // add a etcd endpoint
		playbook.RESTART_SERVICE,  // restart all mds then add a etcd endpoint in mds.conf
		playbook.REMOVE_ETCD_MEMBER,
		playbook.STOP_SERVICE,
		playbook.CLEAN_SERVICE, // only container
		// playbook.AMEND_MDS_CONFIG, // remove a etcd endpoint
		// playbook.RESTART_SERVICE, // restart all mds then remove a etcd endpoint in mds.conf
		// playbook.RELOAD_METASERVER
		playbook.UPDATE_TOPOLOGY,
	}

	// mds
	MIGRATE_MDS_STEPS = []int{
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_MDS,
		playbook.STOP_SERVICE,
		playbook.CLEAN_SERVICE, // only container
		playbook.UPDATE_TOPOLOGY,
	}

	// snapshotclone
	MIGRATE_SNAPSHOTCLONE_STEPS = []int{
		playbook.STOP_SERVICE,
		playbook.CLEAN_SERVICE, // only container
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SNAPSHOTCLONE,
		playbook.UPDATE_TOPOLOGY,
	}

	// chunkserevr (curvebs)
	MIGRATE_CHUNKSERVER_STEPS = []int{
		playbook.BACKUP_ETCD_DATA,
		playbook.CREATE_PHYSICAL_POOL, // add machine that migrate to
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_CHUNKSERVER,
		playbook.MARK_SERVER_PENGDDING, // start migrate to new server
	}

	// metaserver (curvefs)
	MIGRATE_METASERVER_STEPS = []int{
		playbook.BACKUP_ETCD_DATA,
		playbook.CREATE_LOGICAL_POOL,
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_METASERVER,
		playbook.STOP_SERVICE, // start migrate to new server
	}

	MIGRATE_ROLE_STEPS = map[string][]int{
		topology.ROLE_ETCD:          MIGRATE_ETCD_STEPS,
		topology.ROLE_MDS:           MIGRATE_MDS_STEPS,
		topology.ROLE_CHUNKSERVER:   MIGRATE_CHUNKSERVER_STEPS,
		topology.ROLE_SNAPSHOTCLONE: MIGRATE_SNAPSHOTCLONE_STEPS,
		topology.ROLE_METASERVER:    MIGRATE_METASERVER_STEPS,
	}

	MIGRATE_POST_CLEAN_STEPS = []int{
		playbook.STOP_SERVICE,         // bs
		playbook.CLEAN_SERVICE,        // bs, fs
		playbook.CREATE_PHYSICAL_POOL, // only for chunkserver, remove server that migrate from
		playbook.CREATE_LOGICAL_POOL,  // only for metaserver, remove server that migrate from
		playbook.UPDATE_TOPOLOGY,      // bs, fs
	}
)

type migrateOptions struct {
	filename        string
	poolset         string
	poolsetDiskType string
	clean           bool
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

	flags := cmd.Flags()
	flags.StringVar(&options.poolset, "poolset", "default", "Specify the poolset")
	flags.StringVar(&options.poolsetDiskType, "poolset-disktype", "ssd", "Specify the disk type of physical pool")
	flags.BoolVar(&options.clean, "clean", false, "Clean migrated environment for chunkserver or metaserver")
	return cmd
}

// NOTE: you can only migrate same role whole host services ervey time
func checkMigrateTopology(curveadm *cli.CurveAdm, data string) error {
	diffs, err := curveadm.DiffTopology(curveadm.ClusterTopologyData(), data)
	if err != nil {
		return err
	}

	m := map[int][]*topology.DeployConfig{}
	for _, diff := range diffs {
		key := diff.DiffType
		m[key] = append(m[key], diff.DeployConfig)
	}

	dcs2add := m[topology.DIFF_ADD]
	dcs2del := m[topology.DIFF_DELETE]
	if len(dcs2add) > len(dcs2del) {
		return errno.ERR_ADD_SERVICE_WHILE_MIGRATING_IS_DENIED
	} else if len(dcs2add) < len(dcs2del) {
		return errno.ERR_DELETE_SERVICE_WHILE_MIGRATING_IS_DENIED
	}

	if len(dcs2add) == 0 {
		return errno.ERR_NO_SERVICES_FOR_MIGRATING
	}
	if !curveadm.IsSameRole(dcs2add) ||
		!curveadm.IsSameRole(dcs2del) ||
		dcs2add[0].GetRole() != dcs2del[0].GetRole() {
		return errno.ERR_REQUIRE_SAME_ROLE_SERVICES_FOR_MIGRATING
	}
	if len(dcs2del) != dcs2del[0].GetInstances() {
		return errno.ERR_REQUIRE_WHOLE_HOST_SERVICES_FOR_MIGRATING
	}

	return nil
}

func getMigrates(curveadm *cli.CurveAdm, data string) []*configure.MigrateServer {
	diffs, _ := diffTopology(curveadm, data)
	dcs2add := diffs[topology.DIFF_ADD]
	dcs2del := diffs[topology.DIFF_DELETE]
	configure.SortDeployConfigs(dcs2add)
	configure.SortDeployConfigs(dcs2del)

	migrates := []*configure.MigrateServer{}
	for i := 0; i < len(dcs2add); i++ {
		migrates = append(migrates, &configure.MigrateServer{
			From: dcs2del[i],
			To:   dcs2add[i],
		})
	}

	return migrates
}

func genMigratePlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig, options migrateOptions, data string) (*playbook.Playbook, error) {
	diffs, _ := diffTopology(curveadm, data)
	dcs2add := diffs[topology.DIFF_ADD]
	dcs2del := diffs[topology.DIFF_DELETE]
	migrates := getMigrates(curveadm, data)
	role := migrates[0].From.GetRole()
	steps := MIGRATE_ROLE_STEPS[role]
	etcdDCs := curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_ETCD)

	// post clean
	if options.clean {
		steps = MIGRATE_POST_CLEAN_STEPS
		if migrates[0].From.GetKind() == common.KIND_CURVEBS {
			steps = append(steps[:3], steps[4:]...)
		} else {
			steps = append(steps[1:2], steps[3:]...)
		}
	}

	poolset := configure.Poolset{
		Name: options.poolset,
		Type: options.poolsetDiskType,
	}

	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		// configs
		config := dcs2add
		switch step {
		case playbook.STOP_SERVICE,
			playbook.CLEAN_SERVICE,
			playbook.ADD_ETCD_MEMBER:
			config = dcs2del
		case playbook.BACKUP_ETCD_DATA:
			config = curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_ETCD)
		case playbook.AMEND_MDS_CONFIG,
			playbook.RESTART_SERVICE:
			config = curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_MDS)
		case
			playbook.CREATE_PHYSICAL_POOL,
			playbook.CREATE_LOGICAL_POOL,
			playbook.MARK_SERVER_PENGDDING:
			config = curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_MDS)[:1]
		}

		// options
		optionsKV := map[string]interface{}{}
		switch step {
		case playbook.CLEAN_SERVICE:
			optionsKV[comm.KEY_CLEAN_ITEMS] = []string{comm.CLEAN_ITEM_CONTAINER}
			optionsKV[comm.KEY_CLEAN_BY_RECYCLE] = true
			optionsKV[comm.KEY_REMOVE_MIGRATED_SERVER] = true
		case playbook.CREATE_PHYSICAL_POOL:
			optionsKV[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_PHYSICAL
			optionsKV[comm.KEY_MIGRATE_SERVERS] = migrates
			optionsKV[comm.KEY_POOLSET] = poolset
		case playbook.CREATE_LOGICAL_POOL:
			optionsKV[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_LOGICAL
			optionsKV[comm.KEY_MIGRATE_SERVERS] = migrates
			optionsKV[comm.KEY_NEW_TOPOLOGY_DATA] = data
			optionsKV[comm.KEY_IF_UPDATE_TOPOLOG] = false
			optionsKV[comm.KEY_POOLSET] = poolset
		case playbook.UPDATE_TOPOLOGY:
			optionsKV[comm.KEY_NEW_TOPOLOGY_DATA] = data
		case playbook.ADD_ETCD_MEMBER,
			playbook.AMEND_ETCD_CONFIG,
			playbook.AMEND_MDS_CONFIG:
			optionsKV[comm.KEY_MIGRATE_SERVERS] = migrates
			optionsKV[comm.KEY_CLUSTER_DCS] = etcdDCs
		}

		pb.AddStep(&playbook.PlaybookStep{
			Type:        step,
			Configs:     config,
			Options:     optionsKV,
			ExecOptions: playbook.ExecOptions{
				// SilentSubBar: step == playbook.UPDATE_TOPOLOGY,
			},
		})
	}
	return pb, nil
}

func displayMigrateTitle(curveadm *cli.CurveAdm, data string) {
	migrates := getMigrates(curveadm, data)
	from := migrates[0].From
	to := migrates[0].To
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.YellowString("NOTICE: cluster '%s' is about to migrate services:", curveadm.ClusterName()))
	curveadm.WriteOutln(color.YellowString("  - Migrate services: %s*%d", from.GetRole(), len(migrates)))
	curveadm.WriteOutln(color.YellowString("  - Migrate host: from %s to %s", from.GetHost(), to.GetHost()))
}

func runMigrate(curveadm *cli.CurveAdm, options migrateOptions) error {
	// TODO(P0): added prechek for target host
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) read topology from file
	data, err := readTopology(curveadm,
		options.filename,
		options.clean,
	)
	if err != nil {
		return err
	}

	// 3) check topology
	err = checkMigrateTopology(curveadm, data)
	if err != nil {
		return err
	}

	if !options.clean {
		// 4) display title
		displayMigrateTitle(curveadm, data)

		// 5) confirm by user
		if pass := tuicomm.ConfirmYes(tuicomm.DEFAULT_CONFIRM_PROMPT); !pass {
			curveadm.WriteOutln(tuicomm.PromptCancelOpetation("migrate service"))
			return errno.ERR_CANCEL_OPERATION
		}
	}

	// 6) generate migrate playbook
	pb, err := genMigratePlaybook(curveadm, dcs, options, data)
	if err != nil {
		return err
	}

	// 8) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 9) print success prompt
	if options.clean {
		return nil
	}
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Services successfully migrateed ^_^."))
	// TODO(P1): warning iff there is changed configs
	// tui.PromptMigrate()
	return nil
}

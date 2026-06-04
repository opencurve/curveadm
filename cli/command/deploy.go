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
	"time"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	utils "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	CLEAN_PRECHECK_ENVIRONMENT = playbook.CLEAN_PRECHECK_ENVIRONMENT
	PULL_IMAGE                 = playbook.PULL_IMAGE
	CREATE_CONTAINER           = playbook.CREATE_CONTAINER
	SYNC_CONFIG                = playbook.SYNC_CONFIG
	START_ETCD                 = playbook.START_ETCD
	ENABLE_ETCD_AUTH           = playbook.ENABLE_ETCD_AUTH
	START_MDS                  = playbook.START_MDS
	CREATE_PHYSICAL_POOL       = playbook.CREATE_PHYSICAL_POOL
	START_CHUNKSERVER          = playbook.START_CHUNKSERVER
	CREATE_LOGICAL_POOL        = playbook.CREATE_LOGICAL_POOL
	START_SNAPSHOTCLONE        = playbook.START_SNAPSHOTCLONE
	START_METASERVER           = playbook.START_METASERVER
	BALANCE_LEADER             = playbook.BALANCE_LEADER

	ROLE_ETCD          = topology.ROLE_ETCD
	ROLE_MDS           = topology.ROLE_MDS
	ROLE_CHUNKSERVER   = topology.ROLE_CHUNKSERVER
	ROLE_SNAPSHOTCLONE = topology.ROLE_SNAPSHOTCLONE
	ROLE_METASERVER    = topology.ROLE_METASERVER
)

var (
	CURVEBS_DEPLOY_STEPS = []int{
		CLEAN_PRECHECK_ENVIRONMENT,
		PULL_IMAGE,
		CREATE_CONTAINER,
		SYNC_CONFIG,
		START_ETCD,
		ENABLE_ETCD_AUTH,
		START_MDS,
		CREATE_PHYSICAL_POOL,
		START_CHUNKSERVER,
		CREATE_LOGICAL_POOL,
		START_SNAPSHOTCLONE,
		BALANCE_LEADER,
	}

	CURVEFS_DEPLOY_STEPS = []int{
		CLEAN_PRECHECK_ENVIRONMENT,
		PULL_IMAGE,
		CREATE_CONTAINER,
		SYNC_CONFIG,
		START_ETCD,
		ENABLE_ETCD_AUTH,
		START_MDS,
		CREATE_LOGICAL_POOL,
		START_METASERVER,
	}

	DEPLOY_FILTER_ROLE = map[int]string{
		START_ETCD:           ROLE_ETCD,
		ENABLE_ETCD_AUTH:     ROLE_ETCD,
		START_MDS:            ROLE_MDS,
		START_CHUNKSERVER:    ROLE_CHUNKSERVER,
		START_SNAPSHOTCLONE:  ROLE_SNAPSHOTCLONE,
		START_METASERVER:     ROLE_METASERVER,
		CREATE_PHYSICAL_POOL: ROLE_MDS,
		CREATE_LOGICAL_POOL:  ROLE_MDS,
		BALANCE_LEADER:       ROLE_MDS,
	}

	DEPLOY_LIMIT_SERVICE = map[int]int{
		CREATE_PHYSICAL_POOL: 1,
		CREATE_LOGICAL_POOL:  1,
		BALANCE_LEADER:       1,
		ENABLE_ETCD_AUTH:     1,
	}

	CAN_SKIP_ROLES = []string{
		ROLE_SNAPSHOTCLONE,
	}
)

type deployOptions struct {
	skip            []string
	insecure        bool
	poolset         string
	poolsetDiskType string
}

func checkDeployOptions(options deployOptions) error {
	supported := utils.Slice2Map(CAN_SKIP_ROLES)
	for _, role := range options.skip {
		if !supported[role] {
			return errno.ERR_UNSUPPORT_SKIPPED_SERVICE_ROLE.
				F("skip role: %s", role)
		}
	}
	return nil
}

func NewDeployCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options deployOptions

	cmd := &cobra.Command{
		Use:   "deploy [OPTIONS]",
		Short: "Deploy cluster",
		Args:  cliutil.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkDeployOptions(options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringSliceVar(&options.skip, "skip", []string{}, "Specify skipped service roles")
	flags.BoolVarP(&options.insecure, "insecure", "k", false, "Deploy without precheck")
	flags.StringVar(&options.poolset, "poolset", "default", "Specify the poolset name")
	flags.StringVar(&options.poolsetDiskType, "poolset-disktype", "ssd", "Specify the disk type of physical pool")

	return cmd
}

func skipServiceRole(deployConfigs []*topology.DeployConfig, options deployOptions) []*topology.DeployConfig {
	skipped := utils.Slice2Map(options.skip)
	dcs := []*topology.DeployConfig{}
	for _, dc := range deployConfigs {
		if skipped[dc.GetRole()] {
			continue
		}
		dcs = append(dcs, dc)
	}
	return dcs
}

func skipDeploySteps(dcs []*topology.DeployConfig, deploySteps []int, options deployOptions) []int {
	steps := []int{}
	skipped := utils.Slice2Map(options.skip)
	for _, step := range deploySteps {
		if (step == START_SNAPSHOTCLONE && skipped[ROLE_SNAPSHOTCLONE]) ||
			(step == ENABLE_ETCD_AUTH && len(dcs) > 0 && !dcs[0].GetEtcdAuthEnable()) {
			continue
		}
		steps = append(steps, step)
	}
	return steps
}

func precheckBeforeDeploy(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options deployOptions) error {
	// 1) skip precheck
	if options.insecure {
		return nil
	}

	// 2) generate precheck playbook
	pb, err := genPrecheckPlaybook(curveadm, dcs, precheckOptions{
		skipSnapshotClone: utils.Slice2Map(options.skip)[ROLE_SNAPSHOTCLONE],
	})
	if err != nil {
		return err
	}

	// 3) run playbook
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) printf success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Congratulations!!! all precheck passed :)"))
	curveadm.WriteOut(color.GreenString("Now we start to deploy cluster, sleep 3 seconds..."))
	time.Sleep(time.Duration(3) * time.Second)
	curveadm.WriteOutln("\n")
	return nil
}

func calcNumOfChunkserver(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) int {
	services := curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_CHUNKSERVER)
	return len(services)
}

func genDeployPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options deployOptions) (*playbook.Playbook, error) {
	var steps []int
	kind := dcs[0].GetKind()
	if kind == topology.KIND_CURVEBS {
		steps = CURVEBS_DEPLOY_STEPS
	} else {
		steps = CURVEFS_DEPLOY_STEPS
	}
	steps = skipDeploySteps(dcs, steps, options)
	poolset := configure.Poolset{
		Name: options.poolset,
		Type: options.poolsetDiskType,
	}
	diskType := options.poolsetDiskType

	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		// configs
		config := dcs
		if len(DEPLOY_FILTER_ROLE[step]) > 0 {
			role := DEPLOY_FILTER_ROLE[step]
			config = curveadm.FilterDeployConfigByRole(config, role)
		}
		n := len(config)
		if DEPLOY_LIMIT_SERVICE[step] > 0 {
			n = DEPLOY_LIMIT_SERVICE[step]
			config = config[:n]
		}

		// options
		options := map[string]interface{}{}
		if step == CREATE_PHYSICAL_POOL {
			options[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_PHYSICAL
			options[comm.KEY_POOLSET] = poolset
			options[comm.KEY_NUMBER_OF_CHUNKSERVER] = calcNumOfChunkserver(curveadm, dcs)
		} else if step == CREATE_LOGICAL_POOL {
			options[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_LOGICAL
			options[comm.POOLSET] = poolset
			options[comm.POOLSET_DISK_TYPE] = diskType
			options[comm.KEY_NUMBER_OF_CHUNKSERVER] = calcNumOfChunkserver(curveadm, dcs)
		}

		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: config,
			Options: options,
		})
	}
	return pb, nil
}

func statistics(dcs []*topology.DeployConfig) map[string]int {
	count := map[string]int{}
	for _, dc := range dcs {
		count[dc.GetRole()]++
	}
	return count
}

func serviceStats(dcs []*topology.DeployConfig) string {
	count := statistics(dcs)
	netcd := count[topology.ROLE_ETCD]
	nmds := count[topology.ROLE_MDS]
	nchunkserevr := count[topology.ROLE_METASERVER]
	nsnapshotclone := count[topology.ROLE_SNAPSHOTCLONE]
	nmetaserver := count[topology.ROLE_METASERVER]

	var serviceStats string
	kind := dcs[0].GetKind()
	if kind == topology.KIND_CURVEBS { // KIND_CURVEBS
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, chunkserver*%d, snapshotclone*%d",
			netcd, nmds, nchunkserevr, nsnapshotclone)
	} else { // KIND_CURVEFS
		serviceStats = fmt.Sprintf("etcd*%d, mds*%d, metaserver*%d",
			netcd, nmds, nmetaserver)
	}

	return serviceStats
}

func displayDeployTitle(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig) {
	curveadm.WriteOutln("Cluster Name    : %s", curveadm.ClusterName())
	curveadm.WriteOutln("Cluster Kind    : %s", dcs[0].GetKind())
	curveadm.WriteOutln("Cluster Services: %s", serviceStats(dcs))
	curveadm.WriteOutln("")
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
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) skip service role
	dcs = skipServiceRole(dcs, options)

	// 3) precheck before deploy
	err = precheckBeforeDeploy(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 4) generate deploy playbook
	pb, err := genDeployPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 5) display title
	displayDeployTitle(curveadm, dcs)

	// 6) run playground
	if err = pb.Run(); err != nil {
		return err
	}

	// 7) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Cluster '%s' successfully deployed ^_^."), curveadm.ClusterName())
	return nil
}

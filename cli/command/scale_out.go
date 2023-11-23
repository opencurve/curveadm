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
	"time"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	utils "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	// etcd
	SCALE_OUT_ETCD_STEPS = []int{
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SERVICE,
		playbook.UPDATE_TOPOLOGY,
	}

	// mds
	SCALE_OUT_MDS_STEPS = []int{
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SERVICE,
		playbook.UPDATE_TOPOLOGY,
	}

	// snapshotclone (curvebs)
	SCALE_OUT_SNAPSHOTCLONE_STEPS = []int{
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SERVICE,
		playbook.UPDATE_TOPOLOGY,
	}

	// chunkserevr (curvebs)
	SCALE_OUT_CHUNKSERVER_STEPS = []int{
		playbook.BACKUP_ETCD_DATA,
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.CREATE_PHYSICAL_POOL,
		playbook.START_SERVICE,
		playbook.CREATE_LOGICAL_POOL,
	}

	// metaserver (curvefs)
	SCALE_OUT_METASERVER_STEPS = []int{
		playbook.BACKUP_ETCD_DATA,
		playbook.PULL_IMAGE,
		playbook.CREATE_CONTAINER,
		playbook.SYNC_CONFIG,
		playbook.START_SERVICE,
		playbook.CREATE_LOGICAL_POOL,
	}

	SCALE_OUT_ROLE_STEPS = map[string][]int{
		topology.ROLE_ETCD:          SCALE_OUT_ETCD_STEPS,
		topology.ROLE_MDS:           SCALE_OUT_MDS_STEPS,
		topology.ROLE_CHUNKSERVER:   SCALE_OUT_CHUNKSERVER_STEPS,
		topology.ROLE_SNAPSHOTCLONE: SCALE_OUT_SNAPSHOTCLONE_STEPS,
		topology.ROLE_METASERVER:    SCALE_OUT_METASERVER_STEPS,
	}

	SCALE_OUT_SCALE_OUT_FILTER_ROLE = map[int]string{
		playbook.BACKUP_ETCD_DATA:     ROLE_ETCD,
		playbook.START_ETCD:           ROLE_ETCD,
		playbook.START_MDS:            ROLE_MDS,
		playbook.START_CHUNKSERVER:    ROLE_CHUNKSERVER,
		playbook.START_SNAPSHOTCLONE:  ROLE_SNAPSHOTCLONE,
		playbook.START_METASERVER:     ROLE_METASERVER,
		playbook.CREATE_PHYSICAL_POOL: ROLE_MDS,
		playbook.CREATE_LOGICAL_POOL:  ROLE_MDS,
		playbook.BALANCE_LEADER:       ROLE_MDS,
	}

	LIMIT_SERVICE = map[int]int{
		playbook.CREATE_PHYSICAL_POOL: 1,
		playbook.CREATE_LOGICAL_POOL:  1,
		playbook.BALANCE_LEADER:       1,
	}
)

type scaleOutOptions struct {
	insecure        bool
	filename        string
	poolset         string
	poolsetDiskType string
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

	flags := cmd.Flags()
	flags.BoolVarP(&options.insecure, "insecure", "k", false,
		"Scale out cluster without precheck")
	flags.StringVar(&options.poolset, "poolset", "default", "Specify the poolset name")
	flags.StringVar(&options.poolsetDiskType, "poolset-disktype", "ssd", "Specify the disk type of physical pool")

	return cmd
}

func readTopology(curveadm *cli.CurveAdm, filename string) (string, error) {
	if !utils.PathExist(filename) {
		return "", errno.ERR_TOPOLOGY_FILE_NOT_FOUND.
			F("%s: no such file", utils.AbsPath(filename))
	}

	data, err := utils.ReadFile(filename)
	if err != nil {
		return "", errno.ERR_READ_TOPOLOGY_FILE_FAILED.E(err)
	}

	oldData := curveadm.ClusterTopologyData()
	curveadm.WriteOut("%s", utils.Diff(oldData, data))
	return data, nil
}

func diffTopology(curveadm *cli.CurveAdm, data string) (map[int][]*topology.DeployConfig, error) {
	diffs, err := curveadm.DiffTopology(curveadm.ClusterTopologyData(), data)
	if err != nil {
		return nil, err
	}

	m := map[int][]*topology.DeployConfig{}
	for _, diff := range diffs {
		key := diff.DiffType
		m[key] = append(m[key], diff.DeployConfig)
	}
	return m, nil
}

func getHostNum(dcs []*topology.DeployConfig) int {
	num := 0
	exist := map[string]bool{}
	for _, dc := range dcs {
		parentId := dc.GetParentId()
		if _, ok := exist[parentId]; !ok {
			num++
			exist[parentId] = true
		}
	}
	return num
}

func checkScaleOutTopology(curveadm *cli.CurveAdm, data string) error {
	diffs, err := diffTopology(curveadm, data)
	if err != nil {
		return err
	}

	dcs2del := diffs[topology.DIFF_DELETE]
	if len(dcs2del) > 0 {
		return errno.ERR_DELETE_SERVICE_WHILE_SCALE_OUT_CLUSTER_IS_DENIED
	}
	dcs2add := diffs[topology.DIFF_ADD]
	if len(dcs2add) == 0 {
		return errno.ERR_NO_SERVICES_FOR_SCALE_OUT_CLUSTER
	}
	if !curveadm.IsSameRole(dcs2add) {
		return errno.ERR_REQUIRE_SAME_ROLE_SERVICES_FOR_SCALE_OUT_CLUSTER
	}

	role := dcs2add[0].GetRole()
	num := getHostNum(dcs2add)
	switch role {
	case topology.ROLE_CHUNKSERVER:
		if num < 3 {
			return errno.ERR_CHUNKSERVER_REQUIRES_3_HOSTS_WHILE_SCALE_OUT.
				F("host num: %d", num)
		}
	case topology.ROLE_METASERVER:
		if num < 3 {
			return errno.ERR_METASERVER_REQUIRES_3_HOSTS_WHILE_SCALE_OUT.
				F("host num: %d", num)
		}
	}

	return nil
}

func genScaleOutPrecheckPlaybook(curveadm *cli.CurveAdm, data string) (*playbook.Playbook, error) {
	dcsAll, _ := curveadm.ParseTopologyData(data)
	kind := dcsAll[0].GetKind()
	steps := CURVEFS_PRECHECK_STEPS
	if kind == topology.KIND_CURVEBS {
		steps = CURVEBS_PRECHECK_STEPS
	}
	diffs, _ := diffTopology(curveadm, data)
	dcs2scaleOut := diffs[topology.DIFF_ADD]

	// add playbook step
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		configs := dcs2scaleOut
		switch step {
		case playbook.CHECK_TOPOLOGY:
			configs = dcsAll[:1] // any deploy config
		case playbook.CHECK_KERNEL_VERSION:
			configs = curveadm.FilterDeployConfigByRole(configs, ROLE_CHUNKSERVER)
		case playbook.CHECK_NETWORK_FIREWALL:
			// TODO(P0): enable check network firewall
			// TODO(P1): chech wether other services can connect to the scaled out services
			configs = configs[:0]
		case playbook.GET_HOST_DATE:
			configs = dcsAll
		case playbook.CHECK_HOST_DATE:
			configs = configs[:1]
		case playbook.CHECK_CHUNKFILE_POOL:
			configs = curveadm.FilterDeployConfigByRole(configs, ROLE_CHUNKSERVER)
		}

		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: configs,
			Options: map[string]interface{}{
				comm.KEY_ALL_DEPLOY_CONFIGS:       dcsAll,
				comm.KEY_CHECK_WITH_WEAK:          true,
				comm.KEY_CHECK_SKIP_SNAPSHOECLONE: true,
			},
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: step == playbook.CHECK_HOST_DATE,
			},
		})
	}

	// add playbook post steps
	steps = PRECHECK_POST_STEPS
	for _, step := range steps {
		pb.AddPostStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs2scaleOut,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: true,
			},
		})
	}

	return pb, nil
}

func precheckBeforeScaleOut(curveadm *cli.CurveAdm, options scaleOutOptions, data string) error {
	// 1) skip precheck
	if options.insecure {
		return nil
	}

	// 2) generate precheck playbook
	pb, err := genScaleOutPrecheckPlaybook(curveadm, data)
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
	curveadm.WriteOut(color.GreenString("Now we start to scale out cluster, sleep 3 seconds..."))
	time.Sleep(time.Duration(3) * time.Second)
	curveadm.WriteOutln("\n")
	return nil
}

func genScaleOutPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	data string,
	options scaleOutOptions) (*playbook.Playbook, error) {
	diffs, _ := diffTopology(curveadm, data)
	dcs2scaleOut := diffs[topology.DIFF_ADD]
	role := dcs2scaleOut[0].GetRole()
	steps := SCALE_OUT_ROLE_STEPS[role]
	poolset := configure.Poolset{Name: options.poolset, Type: options.poolsetDiskType}

	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		// configs
		config := dcs2scaleOut
		switch step {
		case playbook.BACKUP_ETCD_DATA:
			config = curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_ETCD)
		case CREATE_PHYSICAL_POOL,
			CREATE_LOGICAL_POOL:
			config = curveadm.FilterDeployConfigByRole(dcs, topology.ROLE_MDS)[:1]
		}

		// options
		options := map[string]interface{}{}
		switch step {
		case CREATE_PHYSICAL_POOL:
			options[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_PHYSICAL
			options[comm.KEY_SCALE_OUT_CLUSTER] = dcs2scaleOut
			options[comm.KEY_NEW_TOPOLOGY_DATA] = data
			options[comm.KEY_NUMBER_OF_CHUNKSERVER] = calcNumOfChunkserver(curveadm, dcs) +
				calcNumOfChunkserver(curveadm, dcs2scaleOut)
			options[comm.KEY_POOLSET] = poolset
		case CREATE_LOGICAL_POOL:
			options[comm.KEY_CREATE_POOL_TYPE] = comm.POOL_TYPE_LOGICAL
			options[comm.KEY_SCALE_OUT_CLUSTER] = dcs2scaleOut
			options[comm.KEY_NEW_TOPOLOGY_DATA] = data
			options[comm.KEY_NUMBER_OF_CHUNKSERVER] = calcNumOfChunkserver(curveadm, dcs) +
				calcNumOfChunkserver(curveadm, dcs2scaleOut)
			options[comm.KEY_POOLSET] = poolset
		case playbook.UPDATE_TOPOLOGY:
			options[comm.KEY_NEW_TOPOLOGY_DATA] = data
		}

		// exec options
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: config,
			Options: options,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: step == playbook.UPDATE_TOPOLOGY,
			},
		})
	}
	return pb, nil
}

func displayScaleOutTitle(curveadm *cli.CurveAdm, data string) {
	diffs, _ := diffTopology(curveadm, data)
	dcs := diffs[topology.DIFF_ADD]
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.YellowString("NOTICE: cluster '%s' is about to scale out:",
		curveadm.ClusterName()))
	curveadm.WriteOutln(color.YellowString("  - Scale out services: %s*%d",
		dcs[0].GetRole(), len(dcs)))
}

func runScaleOut(curveadm *cli.CurveAdm, options scaleOutOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) read topology from file
	data, err := readTopology(curveadm, options.filename)
	if err != nil {
		return err
	}

	// 3) check topology
	err = checkScaleOutTopology(curveadm, data)
	if err != nil {
		return err
	}

	// 4) display title
	displayScaleOutTitle(curveadm, data)

	// 5) confirm by user
	if pass := tui.ConfirmYes(tui.DEFAULT_CONFIRM_PROMPT); !pass {
		curveadm.WriteOutln(tui.PromptCancelOpetation("scale-out"))
		return nil
	}

	// 6) precheck before deploy
	err = precheckBeforeScaleOut(curveadm, options, data)
	if err != nil {
		return err
	}

	// 7) generate scale-out playbook
	pb, err := genScaleOutPlaybook(curveadm, dcs, data, options)
	if err != nil {
		return err
	}

	// 8) run playground
	if err = pb.Run(); err != nil {
		return err
	}

	// 9) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Cluster '%s' successfully scaled out ^_^."),
		curveadm.ClusterName())
	// TODO(P1): warning iff there is changed configs
	// tui.PromptScaleOut()
	return nil
}

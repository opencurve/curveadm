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

// __SIGN_BY_WINE93__

package cluster

import (
	"github.com/google/uuid"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/utils"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"github.com/spf13/cobra"
)

const (
	ADD_EXAMPLE = `Examples:
  $ curveadm add my-cluster                            # Add a cluster named 'my-cluster'
  $ curveadm add my-cluster -m "deploy for test"       # Add a cluster with description
  $ curveadm add my-cluster -f /path/to/topology.yaml  # Add a cluster with specified topology
  $ curveadm add my-cluster --type develop             # Add a cluster with specified type (develop,production,test)`
)

var (
	CHECK_TOPOLOGY_PLAYBOOK_STEPS = []int{
		playbook.CHECK_TOPOLOGY,
	}
	SUPPORTED_DEPLOY_TYPES = []string{
		"production",
		"test",
		"develop",
	}
)

type addOptions struct {
	name        string
	description string
	filename    string
	deployType  string
}

func NewAddCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options addOptions

	cmd := &cobra.Command{
		Use:     "add CLUSTER [OPTIONS]",
		Short:   "Add cluster",
		Args:    utils.ExactArgs(1),
		Example: ADD_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkAddOptions(cmd)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runAdd(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.description, "description", "m", "", "Description for cluster")
	flags.StringVarP(&options.filename, "topology", "f", "", "Specify the path of topology file")
	flags.StringVar(&options.deployType, "type", "develop", "Specify the type of cluster")
	return cmd
}

func readTopology(filename string) (string, error) {
	if len(filename) == 0 {
		return "", nil
	} else if !utils.PathExist(filename) {
		return "", errno.ERR_TOPOLOGY_FILE_NOT_FOUND.
			F("%s: no such file", utils.AbsPath(filename))
	}

	data, err := utils.ReadFile(filename)
	if err != nil {
		return "", errno.ERR_READ_TOPOLOGY_FILE_FAILED.E(err)
	}
	return data, nil
}

func genCheckTopologyPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options addOptions) (*playbook.Playbook, error) {
	steps := CHECK_TOPOLOGY_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: nil,
			Options: map[string]interface{}{
				comm.KEY_ALL_DEPLOY_CONFIGS:       dcs,
				comm.KEY_CHECK_SKIP_SNAPSHOECLONE: false,
				comm.KEY_CHECK_WITH_WEAK:          true,
			},
			ExecOptions: playbook.ExecOptions{
				Concurrency:   100,
				SilentSubBar:  true,
				SilentMainBar: true,
				SkipError:     false,
			},
		})
	}
	return pb, nil
}

func checkTopology(curveadm *cli.CurveAdm, data string, options addOptions) error {
	if len(options.filename) == 0 {
		return nil
	}

	dcs, err := curveadm.ParseTopologyData(data)
	if err != nil {
		return err
	}

	pb, err := genCheckTopologyPlaybook(curveadm, dcs, options)
	if err != nil {
		return err
	}
	return pb.Run()
}

func checkAddOptions(cmd *cobra.Command) error {
	deployType, err := cmd.Flags().GetString("type")
	if err != nil {
		return err
	}
	for _, t := range SUPPORTED_DEPLOY_TYPES {
		if deployType == t {
			return nil
		}
	}
	return errno.ERR_UNSUPPORT_DEPLOY_TYPE.F("deploy type: %s", deployType)
}

func runAdd(curveadm *cli.CurveAdm, options addOptions) error {
	// 1) check wether cluster already exist
	name := options.name
	storage := curveadm.Storage()
	clusters, err := storage.GetClusters(name)
	if err != nil {
		log.Error("Get clusters failed",
			log.Field("cluster name", name),
			log.Field("error", err))
		return errno.ERR_GET_ALL_CLUSTERS_FAILED.E(err)
	} else if len(clusters) > 0 {
		return errno.ERR_CLUSTER_ALREADY_EXIST.
			F("cluster name: %s", name)
	}

	// 2) read topology iff specified and validte it
	data, err := readTopology(options.filename)
	if err != nil {
		return err
	}

	// 3) check topology
	err = checkTopology(curveadm, data, options)
	if err != nil {
		return err
	}

	// 4) insert cluster (with topology) into database
	uuid := uuid.NewString()
	err = storage.InsertCluster(name, uuid, options.description, data, options.deployType)
	if err != nil {
		return errno.ERR_INSERT_CLUSTER_FAILED.E(err)
	}

	// 5) print success prompt
	curveadm.WriteOutln("Added cluster '%s'", name)
	return nil
}

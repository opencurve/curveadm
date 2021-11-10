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

package cluster

import (
	"fmt"

	"github.com/opencurve/curveadm/internal/utils"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/spf13/cobra"
)

var (
	addExample = `Examples:
  $ curveadm add my-cluster                            # Add a cluster named 'my-cluster'
  $ curveadm add my-cluster -m "deploy for test"       # Add a cluster with description
  $ curveadm add my-cluster -f /path/to/topology.yaml  # Add a cluster with specified topology`
)

type addOptions struct {
	name        string
	descriotion string
	filename    string
}

func NewAddCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options addOptions

	cmd := &cobra.Command{
		Use:     "add CLUSTER [OPTIONS]",
		Short:   "Add cluster",
		Args:    utils.ExactArgs(1),
		Example: addExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runAdd(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.descriotion, "description", "m", "", "Description for cluster")
	flags.StringVarP(&options.filename, "topology", "f", "", "Specify the path of topology file")

	return cmd
}

func readTopology(filename string) (string, error) {
	if filename == "" {
		return "", nil
	}
	return utils.ReadFile(filename)
}

func runAdd(curveadm *cli.CurveAdm, options addOptions) error {
	name := options.name
	storage := curveadm.Storage()
	clusters, err := storage.GetClusters(name)
	if err != nil {
		log.Error("GetClusters", log.Field("error", err))
		return err
	} else if len(clusters) > 0 {
		return fmt.Errorf("cluster %s already exist", name)
	}

	// read topology from file
	data, err := readTopology(options.filename)
	if err != nil {
		return err
	}
	return storage.InsertCluster(name, options.descriotion, data)
}

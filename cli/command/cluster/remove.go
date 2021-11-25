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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	clusterName string
}

func NewRemoveCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options removeOptions

	cmd := &cobra.Command{
		Use:     "rm CLUSTER",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove cluster",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.clusterName = args[0]
			return runRemove(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runRemove(curveadm *cli.CurveAdm, options removeOptions) error {
	storage := curveadm.Storage()
	clusterName := options.clusterName
	clusters, err := storage.GetClusters(clusterName) // Get all cluster
	if err != nil {
		log.Error("GetClusters", log.Field("error", err))
		return err
	} else if len(clusters) == 0 {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	// TODO(@Wine93): check all service has removed before delete cluster
	if pass := common.ConfirmYes("Do you want to continue? [y/N]: "); !pass {
		return nil
	} else if err := curveadm.Storage().DeleteCluster(clusterName); err != nil {
		return err
	}

	curveadm.WriteOut("Deleted cluster '%s'\n", clusterName)
	return nil
}

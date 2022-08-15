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
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	clusterName string
	force       bool
}

func NewRemoveCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options removeOptions

	cmd := &cobra.Command{
		Use:     "rm CLUSTER [OPTIONS]",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove cluster",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.clusterName = args[0]
			return runRemove(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.force, "force", "f", false, "Remove cluster by force")

	return cmd
}

func checkAllServicesRemoved(curveadm *cli.CurveAdm, options removeOptions, clusterId int) error {
	if options.force {
		return nil
	}

	services, err := curveadm.Storage().GetServices(clusterId)
	if err != nil {
		return errno.ERR_GET_ALL_SERVICES_CONTAINER_ID_FAILED.E(err)
	}

	for _, service := range services {
		if len(service.ContainerId) != 0 &&
			service.ContainerId != comm.CLEANED_CONTAINER_ID {
			return errno.ERR_ACTIVE_SERVICE_IN_CLUSTER.
				F("service id: %s", service.Id)
		}
	}
	return nil
}

func runRemove(curveadm *cli.CurveAdm, options removeOptions) error {
	// 1) get cluster by name
	storage := curveadm.Storage()
	clusterName := options.clusterName
	clusters, err := storage.GetClusters(clusterName) // Get all clusters
	if err != nil {
		log.Error("Get cluster failed",
			log.Field("error", err))
		return errno.ERR_GET_ALL_CLUSTERS_FAILED.E(err)
	} else if len(clusters) == 0 {
		return errno.ERR_CLUSTER_NOT_FOUND.
			F("cluster name: %s", clusterName)
	}

	// 2) remove cluster
	//   2.1): check wether all services removed (ignore by force)
	//   2.2): confirm by user
	//   2.3): delete cluster in database
	if err := checkAllServicesRemoved(curveadm, options, clusters[0].Id); err != nil {
		return err
	} else if pass := tui.ConfirmYes(tui.PromptRemoveCluster(clusterName)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("remove cluster"))
		return errno.ERR_CANCEL_OPERATION
	} else if err := curveadm.Storage().DeleteCluster(clusterName); err != nil {
		return errno.ERR_DELETE_CLUSTER_FAILED.E(err)
	}

	// 3) print success prompt
	curveadm.WriteOutln("Deleted cluster '%s'", clusterName)
	return nil
}

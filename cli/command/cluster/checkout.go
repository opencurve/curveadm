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

	"github.com/opencurve/curveadm/pkg/log"

	cliutil "github.com/opencurve/curveadm/internal/utils"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/spf13/cobra"
)

type checkoutOptions struct {
	clusterName string
}

func NewCheckoutCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options checkoutOptions

	cmd := &cobra.Command{
		Use:   "checkout CLUSTER",
		Short: "Switch cluster",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.clusterName = args[0]
			return runCheckout(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runCheckout(curveadm *cli.CurveAdm, options checkoutOptions) error {
	clusterName := options.clusterName
	storage := curveadm.Storage()
	clusters, err := storage.GetClusters(clusterName)
	if err != nil {
		log.Error("GetClusters", log.Field("error", err))
		return err
	} else if len(clusters) == 0 {
		return fmt.Errorf("cluster '%s' not found", clusterName)
	}

	if err = storage.CheckoutCluster(clusterName); err != nil {
		return err
	}

	curveadm.WriteOut("Switched to cluster '%s'\n", clusterName)
	return nil
}

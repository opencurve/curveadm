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

package config

import (
	"encoding/json"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/pool"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type showOptions struct {
	showPool bool
}

func NewShowCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options showOptions

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show cluster topology",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runShow(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.showPool, "pool", "p", false, "Show cluster pool information")

	return cmd
}

func runShow(curveadm *cli.CurveAdm, options showOptions) error {
	if !options.showPool {
		curveadm.WriteOut("%s", curveadm.ClusterTopologyData())
		return nil
	}

	// show cluster pool information
	pool := pool.CurveClusterTopo{}
	err := json.Unmarshal([]byte(curveadm.ClusterPoolData()), &pool)
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(pool, "", "    ")
	if err != nil {
		return err
	}

	curveadm.WriteOut("%s\n", string(bytes))
	return nil
}

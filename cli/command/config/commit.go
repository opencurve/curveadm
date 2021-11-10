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
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	setExample = `Examples:
  $ curveadm config commit /path/to/topology.yaml  # Commit cluster topology`
)

type commitOptions struct {
	filename string
	slient   bool
}

func NewCommitCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options commitOptions

	cmd := &cobra.Command{
		Use:     "commit TOPOLOGY",
		Short:   "Commit cluster topology",
		Args:    utils.ExactArgs(1),
		Example: setExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filename = args[0]
			return runCommit(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.slient, "slient", "s", false, "Slient output for config commit")

	return cmd
}

func validateTopology(oldData, newData string) error {
	if oldData == "" {
		return nil
	} else if dcs, err := configure.ParseTopology(oldData); err != nil {
		return err
	} else if len(dcs) == 0 {
		return nil
	}

	diffs, err := configure.DiffTopology(oldData, newData)
	if err != nil {
		return err
	}

	for _, diff := range diffs {
		switch diff.DiffType {
		case configure.DIFF_DELETE:
			return fmt.Errorf("you can't delete service in config setting")
		case configure.DIFF_ADD:
			return fmt.Errorf("you can't add service in config setting")
		}
	}

	return nil
}

func runCommit(curveadm *cli.CurveAdm, options commitOptions) error {
	oldData := curveadm.ClusterTopologyData()
	newData, err := utils.ReadFile(options.filename)
	if err != nil {
		return err
	}

	if !options.slient {
		diff := utils.Diff(oldData, newData)
		curveadm.WriteOut("%s\n", diff)
	}

	if err := validateTopology(oldData, newData); err != nil {
		return err
	} else if pass := common.ConfirmYes("Do you want to continue? [y/N]: "); !pass {
		return nil
	} else if err := curveadm.Storage().SetClusterTopology(curveadm.ClusterId(), newData); err != nil {
		return err
	}

	curveadm.WriteOut("Cluster '%s' topology updated\n", curveadm.ClusterName())
	return nil
}

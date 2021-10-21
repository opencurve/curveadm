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
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	diffExample = `Examples:
  $ curveadm config diff /path/to/topology.yaml  # Display difference for topology`
)

type diffOptions struct {
	filename string
}

func NewDiffCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options diffOptions

	cmd := &cobra.Command{
		Use:     "diff TOPOLOGY",
		Short:   "Display difference for topology",
		Args:    utils.ExactArgs(1),
		Example: diffExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filename = args[0]
			return runDiff(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runDiff(curveadm *cli.CurveAdm, options diffOptions) error {
	data1 := curveadm.ClusterTopologyData()
	data2, err := utils.ReadFile(options.filename)
	if err != nil {
		return err
	}

	diff := utils.Diff(data1, data2)
	curveadm.Out().Write([]byte(diff))
	return nil
}

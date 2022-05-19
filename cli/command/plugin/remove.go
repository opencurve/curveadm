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
 * Created Date: 2022-05-17
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"github.com/opencurve/curveadm/cli/cli"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	name string
}

func NewRemoveCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options removeOptions

	cmd := &cobra.Command{
		Use:     "rm PLUGIN",
		Aliases: []string{"remove", "delete"},
		Short:   "Remove plugin",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runRemove(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runRemove(curveadm *cli.CurveAdm, options removeOptions) error {
	err := curveadm.PluginManager().Remove(options.name)
	if err == nil {
		curveadm.WriteOut("Plugin '%s' uninstalled\n", options.name)
	}
	return err
}

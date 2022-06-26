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
 * Created Date: 2022-06-23
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/log"
	"github.com/spf13/cobra"
)

type listOptions struct{}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "ls",
		Aliases: []string{"list"},
		Short:   "List playgrounds",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	playgrounds, err := curveadm.Storage().GetPlaygrounds("%")
	if err != nil {
		log.Error("GetPlayground", log.Field("error", err))
		return err
	}

	output := tui.FormatPlayground(playgrounds)
	curveadm.WriteOut(output)
	return nil
}

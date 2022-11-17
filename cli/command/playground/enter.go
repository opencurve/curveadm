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
 * Created Date: 2022-06-24
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type enterOptions struct {
	id string
}

func NewEnterCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options enterOptions

	cmd := &cobra.Command{
		Use:   "enter ID",
		Short: "Enter playground container",
		Args:  utils.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.id = args[0]
			return runEnter(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runEnter(curveadm *cli.CurveAdm, options enterOptions) error {
	// 1) get playground
	id := options.id
	playgrounds, err := curveadm.Storage().GetPlaygroundById(id)
	if err != nil {
		return errno.ERR_GET_PLAYGROUND_BY_NAME_FAILED.E(err)
	} else if len(playgrounds) == 0 {
		return errno.ERR_PLAYGROUND_NOT_FOUND.
			F("id=%s", id)
	}

	// 2) attch local container
	return tools.AttachLocalContainer(curveadm, playgrounds[0].Name)
}

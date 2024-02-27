/*
*  Copyright (c) 2023 NetEase Inc.
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
* Created Date: 2023-12-13
* Author: liuminjian
 */

package daemon

import (
	"github.com/opencurve/curveadm/cli/cli"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

func NewDaemonCommand(curveadm *cli.CurveAdm) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage curveadm daemon service",
		Args:  cliutil.NoArgs,
		RunE:  cliutil.ShowHelp(curveadm.Err()),
	}

	cmd.AddCommand(
		NewStartCommand(curveadm),
		NewStopCommand(curveadm),
	)
	return cmd
}

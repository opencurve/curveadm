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
	"github.com/spf13/cobra"
)

const (
	START_EXAMPLR = `Examples:
  $ curveadm daemon start   # Start daemon service`
)

type startOptions struct {
	filename string
}

func NewStartCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options startOptions
	pigeon := curveadm.GetPigeon()

	cmd := &cobra.Command{
		Use:     "start [OPTIONS]",
		Short:   "Start daemon service",
		Example: START_EXAMPLR,
		RunE: func(cmd *cobra.Command, args []string) error {
			return pigeon.Start(options.filename)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", pigeon.DefaultConfFile(),
		"Specify pigeon configure file")

	return cmd
}

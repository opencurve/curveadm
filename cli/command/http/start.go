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
* Project: Curveadm
* Created Date: 2023-03-31
* Author: wanghai (SeanHai)
 */

package http

import (
	"fmt"
	"os/exec"
	"path"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/spf13/cobra"
)

const (
	START_EXAMPLR = `Examples:
  $ curveadm http start  # Start an http service to receive requests`
	EXEC_PATH = "http/pigeon"
)

func NewStartCommand(curveadm *cli.CurveAdm) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "start [OPTIONS]",
		Short:   "Start http service",
		Example: START_EXAMPLR,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(curveadm)
		},
		DisableFlagsInUseLine: true,
	}
	return cmd
}

func runStart(curveadm *cli.CurveAdm) error {
	path := path.Join(curveadm.RootDir(), EXEC_PATH)
	cmd := exec.Command(path, "start")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s", output)
	}
	return nil
}

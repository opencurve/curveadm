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

// __SIGN_BY_WINE93__

package hosts

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/errno"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	COMMIT_EXAMPLE = `Examples:
  $ curveadm hosts commit /path/to/hosts.yaml  # Commit hosts`
)

type commitOptions struct {
	filename string
	slient   bool
}

func NewCommitCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options commitOptions

	cmd := &cobra.Command{
		Use:     "commit HOSTS [OPTIONS]",
		Short:   "Commit hosts",
		Args:    utils.ExactArgs(1),
		Example: COMMIT_EXAMPLE,
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

func readAndCheckHosts(curveadm *cli.CurveAdm, options commitOptions) (string, error) {
	// 1) read hosts from file
	if !utils.PathExist(options.filename) {
		return "", errno.ERR_HOSTS_FILE_NOT_FOUND.
			F("%s: no such file", utils.AbsPath(options.filename))
	}
	data, err := utils.ReadFile(options.filename)
	if err != nil {
		return data, errno.ERR_READ_HOSTS_FILE_FAILED.E(err)
	}

	// 2) display difference
	oldData := curveadm.Hosts()
	if !options.slient {
		diff := utils.Diff(oldData, data)
		curveadm.WriteOutln(diff)
	}

	// 3) check hosts data
	_, err = hosts.ParseHosts(data)
	return data, err
}

func runCommit(curveadm *cli.CurveAdm, options commitOptions) error {
	// 1) read and check hosts
	data, err := readAndCheckHosts(curveadm, options)
	if err != nil {
		return err
	}

	// 2) confirm by user
	pass := tui.ConfirmYes("Do you want to continue?")
	if !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("commit hosts"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 3) update hosts in database
	err = curveadm.Storage().SetHosts(data)
	if err != nil {
		return errno.ERR_UPDATE_HOSTS_FAILED.E(err)
	}

	// 4) print success prompt
	curveadm.WriteOutln(color.GreenString("Hosts updated"))
	return nil
}

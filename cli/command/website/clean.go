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
* Created Date: 2023-05-08
* Author: wanghai (SeanHai)
 */

package website

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	CLEAN_EXAMPLE = `Examples:
  $ curveadm website clean                                  # Clean everything for website
  $ curveadm website clean --only='data'                    # Clean data for website`
)

var (
	CLEAN_PLAYBOOK_STEPS = []int{
		playbook.CLEAN_WEBSITE,
	}

	CLEAN_ITEMS = []string{
		comm.CLEAN_ITEM_DATA,
		comm.CLEAN_ITEM_LOG,
		comm.CLEAN_ITEM_CONTAINER,
	}
)

type cleanOptions struct {
	filename string
	only     []string
}

func NewCleanCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options cleanOptions

	cmd := &cobra.Command{
		Use:     "clean [OPTIONS]",
		Short:   "Clean website environment",
		Args:    cliutil.NoArgs,
		Example: CLEAN_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "website.yaml", "Specify website configuration file")
	flags.StringSliceVarP(&options.only, "only", "o", CLEAN_ITEMS, "Specify clean item")
	return cmd
}

func genCleanPlaybook(curveadm *cli.CurveAdm,
	wcs []*configure.WebsiteConfig,
	options cleanOptions) (*playbook.Playbook, error) {
	if len(wcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}
	steps := CLEAN_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: wcs,
			Options: map[string]interface{}{
				comm.KEY_CLEAN_ITEMS: options.only,
			},
		})
	}
	return pb, nil
}

func runClean(curveadm *cli.CurveAdm, options cleanOptions) error {
	// 1) parse website config
	wcs, err := configure.ParseWebsiteConfig(options.filename)
	if err != nil {
		return err
	}

	// 2) generate clean playbook
	pb, err := genCleanPlaybook(curveadm, wcs, options)
	if err != nil {
		return err
	}

	// 3) confirm by user
	if pass := tui.ConfirmYes(tui.PromptCleanService(configure.ROLE_WEBSITE, wcs[0].GetHost(), options.only)); !pass {
		curveadm.WriteOut(tui.PromptCancelOpetation("clean website service"))
		return errno.ERR_CANCEL_OPERATION
	}

	// 4) run playground
	err = pb.Run()
	if err != nil {
		return err
	}
	return nil
}

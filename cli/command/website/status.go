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
	"github.com/opencurve/curveadm/internal/task/task/website"
	tui "github.com/opencurve/curveadm/internal/tui/service"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	STATUS_PLAYBOOK_STEPS = []int{
		playbook.INIT_WEBSITE_STATUS,
		playbook.GET_WEBSITE_STATUS,
	}
)

type statusOptions struct {
	filename string
	verbose  bool
}

func NewStatusCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options statusOptions
	cmd := &cobra.Command{
		Use:   "status [OPTIONS]",
		Short: "Display website status",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "website.yaml", "Specify website configuration file")
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output for status")
	return cmd
}

func genStatusPlaybook(curveadm *cli.CurveAdm,
	wcs []*configure.WebsiteConfig) (*playbook.Playbook, error) {
	if len(wcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := STATUS_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: wcs,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar:  true,
				SilentMainBar: step == playbook.INIT_WEBSITE_STATUS,
				SkipError:     true,
			},
		})
	}
	return pb, nil
}

func displayStatus(curveadm *cli.CurveAdm, wcs []*configure.WebsiteConfig, options statusOptions) {
	statuses := []website.WebsiteStatus{}
	value := curveadm.MemStorage().Get(comm.KEY_WEBSITE_STATUS)
	if value != nil {
		m := value.(map[string]website.WebsiteStatus)
		for _, status := range m {
			statuses = append(statuses, status)
		}
	}

	output := tui.FormatWebsiteStatus(statuses, options.verbose)
	curveadm.WriteOutln("")
	curveadm.WriteOut("%s", output)
}

func runStatus(curveadm *cli.CurveAdm, options statusOptions) error {
	// 1) parse website config
	wcs, err := configure.ParseWebsiteConfig(options.filename)
	if err != nil {
		return err
	}

	// 2) generate get status playbook
	pb, err := genStatusPlaybook(curveadm, wcs)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()

	// 4) display service status
	displayStatus(curveadm, wcs, options)
	return err

}

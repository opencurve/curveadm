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
* Created Date: 2023-05-06
* Author: wanghai (SeanHai)
 */

package website

import (
	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	DEPLOY_EXAMPLE = `Examples:
	$ curveadm website deploy -c website.yaml    # deploy website`
)

var (
	DEPLOY_PALYBOOK_STEPS = []int{
		playbook.PULL_WEBSITE_IMAGE,
		playbook.CREATE_WEBSITE_CONTAINER,
		playbook.SYNC_WEBSITE_CONFIG,
		playbook.START_WEBSITE_SERVICE,
	}
)

type deployOptions struct {
	filename string
}

func NewDeployCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options deployOptions

	cmd := &cobra.Command{
		Use:     "deploy [OPTIONS]",
		Short:   "Deploy website",
		Args:    cliutil.NoArgs,
		Example: DEPLOY_EXAMPLE,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDeploy(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "website.yaml", "Specify website configuration file")
	return cmd
}

func genDeployPlaybook(curveadm *cli.CurveAdm,
	wcs []*configure.WebsiteConfig) (*playbook.Playbook, error) {
	steps := DEPLOY_PALYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: wcs,
		})
	}
	return pb, nil
}

func runDeploy(curveadm *cli.CurveAdm, options deployOptions) error {
	// 1) parse website configure
	wcs, err := configure.ParseWebsiteConfig(options.filename)
	if err != nil {
		return err
	}

	// 2) generate deploy playbook
	pb, err := genDeployPlaybook(curveadm, wcs)
	if err != nil {
		return err
	}

	// 5) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 6) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Deploy website success ^_^"))
	return nil
}

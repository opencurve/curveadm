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
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	WEBSITE_START_STEPS = []int{
		playbook.START_WEBSITE_SERVICE,
	}
)

type startOptions struct {
	filename string
}

func NewStartCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options startOptions
	cmd := &cobra.Command{
		Use:   "start [OPTIONS]",
		Short: "Start website service",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStart(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}
	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "conf", "c", "website.yaml", "Specify website configuration file")
	return cmd
}

func genStartPlaybook(curveadm *cli.CurveAdm,
	wcs []*configure.WebsiteConfig) (*playbook.Playbook, error) {
	if len(wcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := WEBSITE_START_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: wcs,
		})
	}
	return pb, nil
}

func runStart(curveadm *cli.CurveAdm, options startOptions) error {
	// 1) parse website configure
	wcs, err := configure.ParseWebsiteConfig(options.filename)
	if err != nil {
		return err
	}

	// 2) generate start playbook
	pb, err := genStartPlaybook(curveadm, wcs)
	if err != nil {
		return err
	}

	// 3) run playground
	return pb.Run()
}

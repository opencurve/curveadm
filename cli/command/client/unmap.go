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
 * Created Date: 2022-01-10
 * Author: Jingli Chen (Wine93)
 */

package client

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	UNMAP_EXAMPLE = `Examples:
  $ curveadm unmap user:volume --host machine1  # Unmap volume`
)

var (
	UNMAP_PLAYBOOK_STEPS = []int{
		playbook.UNMAP_IMAGE,
	}
)

type unmapOptions struct {
	host  string
	image string
}

func checkUnmapOptions(curveadm *cli.CurveAdm, options unmapOptions) error {
	_, _, err := ParseImage(options.image)
	return err
}

func NewUnmapCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options unmapOptions

	cmd := &cobra.Command{
		Use:     "unmap USER:VOLUME [OPTIONS]",
		Short:   "Unmap nbd device",
		Args:    utils.ExactArgs(1),
		Example: UNMAP_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return checkUnmapOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return runUnmap(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")

	return cmd
}

func genUnmapPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options unmapOptions) (*playbook.Playbook, error) {
	user, name, _ := ParseImage(options.image)
	steps := UNMAP_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: nil,
			Options: map[string]interface{}{
				comm.KEY_MAP_OPTIONS: bs.MapOptions{
					Host:   options.host,
					User:   user,
					Volume: name,
				},
			},
		})
	}
	return pb, nil
}

func runUnmap(curveadm *cli.CurveAdm, options unmapOptions) error {
	// 1) generate unmap playbook
	pb, err := genUnmapPlaybook(curveadm, nil, options)
	if err != nil {
		return err
	}

	// 2) run playground
	return pb.Run()
}

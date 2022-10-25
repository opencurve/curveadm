/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-07-20
 * Author: Jingli Chen (Wine93)
 */

package hosts

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	utils "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	verbose bool
	labels  []string
}

func NewListCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options listOptions

	cmd := &cobra.Command{
		Use:     "ls [OPTIONS]",
		Aliases: []string{"list"},
		Short:   "List hosts",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.BoolVarP(&options.verbose, "verbose", "v", false, "Verbose output for hosts")
	flags.StringSliceVarP(&options.labels, "labels", "l", []string{}, "Specify the host labels")

	return cmd
}

func filter(data string, labels []string) ([]*hosts.HostConfig, error) {
	hcs, err := hosts.ParseHosts(data)
	if err != nil {
		return nil, err
	}

	out := []*hosts.HostConfig{}
	m := utils.Slice2Map(labels)
	for _, hc := range hcs {
		if len(m) == 0 {
			out = append(out, hc)
			continue
		}
		for _, label := range hc.GetLabels() {
			if _, ok := m[label]; ok {
				out = append(out, hc)
				break
			}
		}
	}

	return out, nil
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	var hcs []*hosts.HostConfig
	var err error
	hosts := curveadm.Hosts()
	if len(hosts) > 0 {
		hcs, err = filter(hosts, options.labels) // filter hosts
		if err != nil {
			return err
		}
	}

	output := tui.FormatHosts(hcs, options.verbose)
	curveadm.WriteOut(output)
	return nil
}

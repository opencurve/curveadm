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
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/tui"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type listOptions struct {
	verbose bool
	labels  string
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
	flags.StringVarP(&options.labels, "labels", "l", "", "Specify the host labels")

	return cmd
}

/*
 * pattern         description
 * ---             ---
 * label2:label2   multiple labels: all hosts belong to label <label1> plus all hosts belong to <label2>
 * label2:!label2  excluding labels: all hosts belong to label <label1> except those belong to label <label2>
 * label2:&label2  intersection labels: any hosts belong to label <label1> that are also belong to label <label2>
 */
func parsePattern(labels []string) (include, exclude, intersect map[string]bool) {
	include = map[string]bool{}
	exclude = map[string]bool{}
	intersect = map[string]bool{}
	for _, label := range labels {
		if len(label) == 0 {
			continue
		}

		switch label[0] {
		case '!':
			exclude[label[1:]] = true
		case '&':
			intersect[label[1:]] = true
		default:
			include[label] = true
		}
	}
	return
}

// return true if dropped
func excludeOne(hc *hosts.HostConfig, exclude map[string]bool) bool {
	if len(exclude) == 0 {
		return false
	}

	for _, label := range hc.GetLabels() {
		if exclude[label] {
			return true
		}
	}
	return false
}

// return true if selected
func includeOne(hc *hosts.HostConfig, include map[string]bool) bool {
	if len(include) == 0 {
		return true
	}

	for _, label := range hc.GetLabels() {
		if include[label] {
			return true
		}
	}
	return false
}

// return true if selected
func intersectOne(hc *hosts.HostConfig, intersect map[string]bool) bool {
	if len(intersect) == 0 {
		return true
	}

	exist := map[string]bool{}
	for _, label := range hc.GetLabels() {
		if intersect[label] {
			exist[label] = true
		}
	}
	return len(exist) == len(intersect)
}

func filter(data string, labels []string) ([]*hosts.HostConfig, error) {
	hcs, err := hosts.ParseHosts(data)
	if err != nil {
		return nil, err
	}
	if len(labels) == 0 {
		return hcs, nil
	}

	out := []*hosts.HostConfig{}
	include, exclude, intersect := parsePattern(labels)
	for _, hc := range hcs {
		if excludeOne(hc, exclude) {
			continue
		} else if !includeOne(hc, include) {
			continue
		} else if !intersectOne(hc, intersect) {
			continue
		}
		out = append(out, hc)
	}
	return out, nil
}

func runList(curveadm *cli.CurveAdm, options listOptions) error {
	var hcs []*hosts.HostConfig
	var err error
	data := curveadm.Hosts()
	if len(data) > 0 {
		labels := strings.Split(options.labels, ":")
		hcs, err = filter(data, labels) // filter hosts
		if err != nil {
			return err
		}
	}

	output := tui.FormatHosts(hcs, options.verbose)
	curveadm.WriteOut(output)
	return nil
}

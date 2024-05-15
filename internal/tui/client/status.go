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
 * Created Date: 2022-07-31
 * Author: Jingli Chen (Wine93)
 */

package service

import (
	"sort"

	"github.com/fatih/color"
	comm "github.com/opencurve/curveadm/internal/common"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func statusDecorate(status string) string {
	switch status {
	case comm.CLIENT_STATUS_LOSED, comm.CLIENT_STATUS_UNKNOWN:
		return color.RedString(status)
	}
	return status
}

func sortStatuses(statuses []task.ClientStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		s1, s2 := statuses[i], statuses[j]
		if s1.Kind == s2.Kind {
			return s1.Host < s2.Host
		}
		return s1.Kind < s2.Kind
	})
}

func FormatStatus(statuses []task.ClientStatus, verbose bool) string {
	lines := [][]interface{}{}

	// title
	title := []string{
		"Id",
		"Kind",
		"Host",
		"Container Id",
		"Status",
		"Address",
		"Aux Info",
	}
	if verbose {
		title = append(title, "Config Dumpfile")
	}
	first, second := tui.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	// status
	sortStatuses(statuses)
	for _, status := range statuses {
		// line
		line := []interface{}{
			status.Id,
			status.Kind,
			status.Host,
			tui.TrimContainerId(status.ContainerId),
			tui.DecorateMessage{Message: status.Status, Decorate: statusDecorate},
			status.Address,
			status.AuxInfo,
		}
		if verbose {
			line = append(line, status.CfgPath)
		}

		// lines
		lines = append(lines, line)
	}

	output := tui.FixedFormat(lines, 2)
	return output
}

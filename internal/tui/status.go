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

package tui

import (
	"sort"

	task "github.com/opencurve/curveadm/internal/task/task/common"

	"github.com/fatih/color"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

// status: One of created, restarting, running, removing, paused, exited, or dead
func statusDecorate(status string) string {
	if status == "Cleaned" {
		return color.BlueString(status)
	} else if status == "Losed" {
		return color.RedString(status)
	}

	return status
}

func sortStatues(statuses []task.ServiceStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		s1, s2 := statuses[i], statuses[j]
		if s1.Role == s2.Role {
			if s1.Id == s2.Id {
				return s1.Host < s2.Host
			}
			return s1.Id < s2.Id
		}
		return s1.Role < s2.Role
	})
}

func FormatStatus(statuses []task.ServiceStatus, vebose bool) string {
	lines := [][]interface{}{}

	// title
	title := []string{"Id", "Role", "Host", "Container Id", "Status"}
	if vebose {
		title = append(title, "Log Dir")
		title = append(title, "Data Dir")
	}
	first, second := tui.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	// status
	sortStatues(statuses)
	for _, status := range statuses {
		line := []interface{}{
			status.Id,
			status.Role,
			status.Host,
			status.ContainerId,
			tui.DecorateMessage{Message: status.Status, Decorate: statusDecorate},
		}

		if vebose {
			line = append(line, status.LogDir)
			line = append(line, status.DataDir)
		}

		lines = append(lines, line)
	}

	output := tui.FixedFormat(lines, 2)
	return output
}

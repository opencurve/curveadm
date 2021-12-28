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
 * Created Date: 2021-12-29
 * Author: Jingli Chen (Wine93)
 */

package format

import (
	"sort"

	"github.com/opencurve/curveadm/internal/task/task/bs"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

func sortStatues(statuses []bs.FormatStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		s1, s2 := statuses[i], statuses[j]
		if s1.Host == s2.Host {
			return s1.Device < s2.Device
		}
		return s1.Host < s2.Host
	})
}

func FormatStatus(statuses []bs.FormatStatus) string {
	lines := [][]interface{}{}

	// title
	title := []string{"Host", "Device", "MountPoint", "Formatted", "Status"}
	first, second := tui.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	// status
	sortStatues(statuses)
	for _, status := range statuses {
		line := []interface{}{
			status.Host,
			status.Device,
			status.MountPoint,
			status.Formatted,
			status.Status,
		}
		lines = append(lines, line)
	}

	output := tui.FixedFormat(lines, 2)
	return output
}

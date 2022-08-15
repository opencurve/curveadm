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
 * Created Date: 2021-06-24
 * Author: Jingli Chen (Wine93)
 */

package tui

import (
	"sort"

	"github.com/fatih/color"
	comm "github.com/opencurve/curveadm/internal/common"
	pg "github.com/opencurve/curveadm/internal/task/task/playground"
	"github.com/opencurve/curveadm/internal/tui/common"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func playgroundStatusDecorate(status string) string {
	switch status {
	case comm.PLAYGROUDN_STATUS_LOSED:
		return color.RedString(status)
	}
	return status
}

func sortStatues(statuses []pg.PlaygroundStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		s1, s2 := statuses[i], statuses[j]
		return s1.Id < s2.Id
	})
}

func FormatPlayground(statuses []pg.PlaygroundStatus) string {
	lines := [][]interface{}{}
	title := []string{"Id", "Name", "Create Time", "Status"}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	sortStatues(statuses)
	for _, status := range statuses {
		lines = append(lines, []interface{}{
			status.Id,
			status.Name,
			status.CreateTime,
			tuicommon.DecorateMessage{Message: status.Status, Decorate: playgroundStatusDecorate},
		})
	}

	output := common.FixedFormat(lines, 2)
	return output
}

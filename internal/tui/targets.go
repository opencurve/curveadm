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
 * Created Date: 2022-02-09
 * Author: Jingli Chen (Wine93)
 */

package tui

import (
	"sort"

	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/tui/common"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func sortTargets(targets []step.Target) {
	sort.Slice(targets, func(i, j int) bool {
		t1, t2 := targets[i], targets[j]
		return t1.Tid < t2.Tid
	})
}

func FormatTargets(targets []step.Target) string {
	lines := [][]interface{}{}
	title := []string{"Tid", "Host", "Target Name", "Store", "Portal"}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	sortTargets(targets)
	for _, target := range targets {
		lines = append(lines, []interface{}{
			target.Tid,
			target.Host,
			target.Name,
			target.Store,
			target.Portal,
		})
	}

	output := common.FixedFormat(lines, 2)
	return output
}

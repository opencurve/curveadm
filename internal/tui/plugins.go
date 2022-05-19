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
 * Created Date: 2022-05-18
 * Author: Jingli Chen (Wine93)
 */

package tui

import (
	"sort"

	"github.com/opencurve/curveadm/internal/plugin"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func sortPlugins(plugins []*plugin.Plugin) {
	sort.Slice(plugins, func(i, j int) bool {
		p1, p2 := plugins[i], plugins[j]
		return p1.Name < p2.Name
	})
}

func FormatPlugins(plugins []*plugin.Plugin) string {
	lines := [][]interface{}{}

	// title
	title := []string{
		"Plugin",
		"Version",
		"Released Time",
		"Description",
	}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	// status
	sortPlugins(plugins)
	for _, plugin := range plugins {
		line := []interface{}{
			plugin.Name,
			plugin.Version,
			plugin.ReleasedTime,
			tuicommon.TrimPluginDescription(plugin.Description),
		}
		lines = append(lines, line)
	}

	output := tuicommon.FixedFormat(lines, 2)
	return output
}

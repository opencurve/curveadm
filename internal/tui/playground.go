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
	"strconv"

	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tui/common"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func FormatPlayground(playgrounds []storage.Playground) string {
	lines := [][]interface{}{}
	title := []string{"Id", "Nmae", "Create Time", "Status"}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	for _, playground := range playgrounds {
		line := []interface{}{}
		line = append(line, strconv.Itoa(playground.Id))
		line = append(line, playground.Name)
		line = append(line, playground.CreateTime.Format("2006-01-02 15:04:05"))
		line = append(line, playground.Status)

		lines = append(lines, line)
	}

	output := common.FixedFormat(lines, 2)
	return output
}

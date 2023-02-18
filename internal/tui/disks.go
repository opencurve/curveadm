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

package tui

import (
	"github.com/opencurve/curveadm/internal/storage"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func FormatDisks(diskRecords []storage.Disk) string {
	lines := [][]interface{}{}
	title := []string{
		"Host",
		"Device Path",
		"Device Size",
		"Device URI",
		"Disk Format Mount Point",
		"Chunkserver ID",
	}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	for _, dr := range diskRecords {
		lines = append(lines, []interface{}{
			dr.Host,
			dr.Device,
			dr.Size,
			dr.URI,
			dr.MountPoint,
			dr.ChunkServerID,
		})
	}

	return tuicommon.FixedFormat(lines, 2)
}

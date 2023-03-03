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
 * Created Date: 2023-02-24
 * Author: Lijin Xiong (lijin.xiong@zstack.io)
 */

package tui

import (
	"sort"

	"github.com/opencurve/curveadm/internal/storage"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func SortDiskRecords(diskRecords []storage.Disk) {
	sort.Slice(diskRecords, func(i, j int) bool {
		d1, d2 := diskRecords[i], diskRecords[j]
		if d1.Host == d2.Host {
			return d1.Device < d2.Device
		}
		return d1.Host < d2.Host
	})
}

func FormatDisks(diskRecords []storage.Disk) string {
	lines := [][]interface{}{}
	title := []string{
		"Host",
		"Device Path",
		"Device Size",
		"Device URI",
		"Disk Format Mount Point",
		"Service ID",
	}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	SortDiskRecords(diskRecords)
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

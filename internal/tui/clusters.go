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
	"strconv"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/tui/common"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

func currentDecorate(message string) string {
	return color.GreenString(message)
}

func FormatClusters(clusters []storage.Cluster, vebose bool) string {
	lines := [][]interface{}{}
	if vebose {
		title := []string{" ", "Cluster", "Id", "UUId", "Create Time", "Description"}
		first, second := tuicommon.FormatTitle(title)
		second[0] = ""
		lines = append(lines, first)
		lines = append(lines, second)
	}

	for i := 0; i < len(clusters); i++ {
		line := []interface{}{}
		cluster := clusters[i]
		if cluster.Current {
			line = append(line, common.DecorateMessage{Message: "*", Decorate: currentDecorate})
			line = append(line, common.DecorateMessage{Message: cluster.Name, Decorate: currentDecorate})
		} else {
			line = append(line, " ")
			line = append(line, cluster.Name)
		}

		if vebose {
			line = append(line, strconv.Itoa(cluster.Id))
			line = append(line, cluster.UUId)
			line = append(line, cluster.CreateTime.Format("2006-01-02 15:04:05"))
			line = append(line, cluster.Description)
		}

		lines = append(lines, line)
	}

	nspace := 1
	if vebose {
		nspace = 2
	}
	output := common.FixedFormat(lines, nspace)
	return output
}

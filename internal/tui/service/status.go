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

package service

import (
	"fmt"
	"sort"
	"strings"

	"github.com/fatih/color"
	longest "github.com/jpillora/longestcommon"
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	ROLE_ETCD          = topology.ROLE_ETCD
	ROLE_MDS           = topology.ROLE_MDS
	ROLE_CHUNKSERVER   = topology.ROLE_CHUNKSERVER
	ROLE_METASERVER    = topology.ROLE_METASERVER
	ROLE_SNAPSHOTCLONE = topology.ROLE_SNAPSHOTCLONE

	ITEM_ID = iota
	ITEM_CONTAINER_ID
	ITEM_STATUS
	ITEM_LISTEN_PORT
	ITEM_LOG_DIR
	ITEM_DATA_DIR

	STATUS_CLEANED = task.STATUS_CLEANED
	STATUS_LOSED   = task.STATUS_LOSED
	// for replica merged status
	STATUS_RUNNING  = "RUNNING"
	STATUS_STOPPED  = "STOPPED"
	STATUS_ABNORMAL = "ABNORMAL"

	CLEANED_CONTAINER_ID = task.CLEANED_CONTAINER_ID
)

var (
	ROLE_SCORE = map[string]int{
		ROLE_ETCD:          0,
		ROLE_MDS:           1,
		ROLE_CHUNKSERVER:   2,
		ROLE_METASERVER:    2,
		ROLE_SNAPSHOTCLONE: 3,
	}
)

func statusDecorate(status string) string {
	switch status {
	case STATUS_CLEANED:
		return color.BlueString(status)
	case STATUS_LOSED, STATUS_ABNORMAL:
		return color.RedString(status)
	}
	return status
}

func sortStatues(statuses []task.ServiceStatus) {
	sort.Slice(statuses, func(i, j int) bool {
		s1, s2 := statuses[i], statuses[j]
		if s1.Role == s2.Role {
			return s1.SortedKey < s2.SortedKey
		}
		return ROLE_SCORE[s1.Role] < ROLE_SCORE[s2.Role]
	})
}

func id(items []string) string {
	if len(items) == 1 {
		return items[0]
	}
	return "<replica>"
}

func status(items []string) string {
	if len(items) == 1 {
		return items[0]
	}

	count := map[string]int{}
	for _, item := range items {
		status := item
		if strings.HasPrefix(item, "Up") {
			status = STATUS_RUNNING
		} else if strings.HasPrefix(item, "Exited") {
			status = STATUS_STOPPED
		}
		count[status]++
	}

	for status, n := range count {
		if n == len(items) {
			return status
		}
	}
	return STATUS_ABNORMAL
}

func port(items []string) string {
	p := map[string]struct{}{}
	for _, item := range items {
		p[item] = struct{}{}
	}
	ports := make([]string, 0, len(p))
	for k := range p {
		if k != "" {
			ports = append(ports, k)
		}
	}
	if len(ports) > 0 {
		return strings.Join(ports, ",")
	}
	return ""
}

func dir(items []string) string {
	if len(items) == 1 {
		return items[0]
	}

	prefix := longest.Prefix(items)
	first := strings.TrimPrefix(items[0], prefix)
	last := strings.TrimPrefix(items[len(items)-1], prefix)
	limit := utils.Min(5, len(first), len(last))
	return fmt.Sprintf("%s{%s...%s}", prefix, first[:limit], last[:limit])
}

func merge(statuses []task.ServiceStatus, item int) string {
	items := []string{}
	for _, status := range statuses {
		switch item {
		case ITEM_ID:
			items = append(items, status.Id)
		case ITEM_CONTAINER_ID:
			items = append(items, status.ContainerId)
		case ITEM_STATUS:
			items = append(items, status.Status)
		case ITEM_LISTEN_PORT:
			ports := strings.Split(status.ListenPort, ",")
			items = append(items, ports...)
		case ITEM_LOG_DIR:
			items = append(items, status.LogDir)
		case ITEM_DATA_DIR:
			items = append(items, status.DataDir)
		}
	}

	sort.Strings(items)
	switch item {
	case ITEM_ID:
		return id(items)
	case ITEM_CONTAINER_ID:
		return id(items)
	case ITEM_STATUS:
		return status(items)
	case ITEM_LISTEN_PORT:
		return port(items)
	case ITEM_LOG_DIR:
		return dir(items)
	case ITEM_DATA_DIR:
		return dir(items)
	}
	return ""
}

func mergeStatues(statuses []task.ServiceStatus) []task.ServiceStatus {
	ss := []task.ServiceStatus{}
	i, j, n := 0, 0, len(statuses)
	for i = 0; i < n; i++ {
		for j = i + 1; j < n && statuses[i].ParentId == statuses[j].ParentId; j++ {
		}
		status := statuses[i]
		ss = append(ss, task.ServiceStatus{
			Id:          merge(statuses[i:j], ITEM_ID),
			Role:        status.Role,
			Host:        status.Host,
			Replica:     fmt.Sprintf("%d/%s", j-i, strings.Split(status.Replica, "/")[1]),
			ContainerId: merge(statuses[i:j], ITEM_CONTAINER_ID),
			Status:      merge(statuses[i:j], ITEM_STATUS),
			ListenPort:  merge(statuses[i:j], ITEM_LISTEN_PORT),
			LogDir:      merge(statuses[i:j], ITEM_LOG_DIR),
			DataDir:     merge(statuses[i:j], ITEM_DATA_DIR),
		})
		i = j - 1
	}
	return ss
}

func FormatStatus(statuses []task.ServiceStatus, vebose, expand bool) string {
	lines := [][]interface{}{}

	// title
	title := []string{
		"Id",
		"Role",
		"Host",
		"Replica",
		"Container Id",
		"Status",
		"Listen Port",
		"Log Dir",
		"Data Dir",
	}
	first, second := tui.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	// status
	sortStatues(statuses)
	if !expand {
		statuses = mergeStatues(statuses)
	}
	for _, status := range statuses {
		lines = append(lines, []interface{}{
			status.Id,
			status.Role,
			status.Host,
			status.Replica,
			status.ContainerId,
			tui.DecorateMessage{Message: status.Status, Decorate: statusDecorate},
			status.ListenPort,
			status.LogDir,
			status.DataDir,
		})
	}

	// cut column
	locate := utils.Locate(title)
	if !vebose {
		tui.CutColumn(lines, locate["Data Dir"]) // Data Dir
		tui.CutColumn(lines, locate["Log Dir"])  // Log Dir
	}

	output := tui.FixedFormat(lines, 2)
	return output
}

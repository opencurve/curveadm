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
 * Created Date: 2022-05-23
 * Author: Jingli Chen (Wine93)
 */

package tui

import (
	"strconv"

	"github.com/fatih/color"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/storage"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

var (
	code2str = map[int]string{
		comm.AUDIT_STATUS_ABORT:   "ABORT",
		comm.AUDIT_STATUS_SUCCESS: "SUCCESS",
		comm.AUDIT_STATUS_FAIL:    "FAIL",
		comm.AUDIT_STATUS_CANCEL:  "CANCEL",
	}
)

func statusDecorate(message string) string {
	if message == code2str[comm.AUDIT_STATUS_ABORT] {
		return color.HiWhiteString(message)
	} else if message == code2str[comm.AUDIT_STATUS_SUCCESS] {
		return color.GreenString(message)
	} else if message == code2str[comm.AUDIT_STATUS_FAIL] {
		return color.RedString(message)
	} else if message == code2str[comm.AUDIT_STATUS_CANCEL] {
		return color.YellowString(message)
	}

	return message
}

func FormatAuditLogs(auditLogs []storage.AuditLog, verbose bool) string {
	lines := [][]interface{}{}
	title := []string{"Id", "Status", "Execute Time", "Command"}
	if verbose {
		title = append(title, "Work Directory")
		title = append(title, "Error Code")
	}
	first, second := tuicommon.FormatTitle(title)
	lines = append(lines, first)
	lines = append(lines, second)

	for i := 0; i < len(auditLogs); i++ {
		line := []interface{}{}
		auditLog := auditLogs[i]

		// id
		line = append(line, strconv.Itoa(auditLog.Id))
		// status
		status := "UNKNOWN"
		if v, ok := code2str[auditLog.Status]; ok {
			status = v
		}
		line = append(line, tuicommon.DecorateMessage{Message: status, Decorate: statusDecorate})
		// execute time
		line = append(line, auditLog.ExecuteTime.Format("2006-01-02 15:04:05"))
		// command
		line = append(line, auditLog.Command)

		if verbose {
			// work directory
			line = append(line, auditLog.WorkDirectory)
			// error code
			line = append(line, utils.Atoa(auditLog.ErrorCode))
		}

		lines = append(lines, line)
	}

	output := tuicommon.FixedFormat(lines, 2)
	return output
}

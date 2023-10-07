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
 * Created Date: 2021-12-28
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2FormatStatus struct {
		config          *configure.FormatConfig
		deviceUsage     *string
		containerStatus *string
		containerName   string
		memStorage      *utils.SafeMap
	}

	FormatStatus struct {
		Host       string
		Device     string
		MountPoint string
		Formatted  string // 85/90
		Status     string // Done, Mounting, Pulling image, Formating
	}
)

func setFormatStatus(memStorage *utils.SafeMap, id string, status FormatStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]FormatStatus{}
		v := kv.Get(comm.KEY_ALL_FORMAT_STATUS)
		if v != nil {
			m = v.(map[string]FormatStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_FORMAT_STATUS, m)
		return nil
	})
}

/* deviceUsgae:
 *   Use%
 *     1%
 */
func (s *step2FormatStatus) Execute(ctx *context.Context) error {
	config := s.config
	host := config.GetHost()
	device := config.GetDevice()
	mountPoint := config.GetMountPoint()

	// formated
	deviceUsage := "-"
	if len(*s.deviceUsage) > 0 {
		deviceUsage = strings.Split(*s.deviceUsage, "\n")[1]
		deviceUsage = strings.TrimPrefix(deviceUsage, " ")
		deviceUsage = strings.TrimSuffix(deviceUsage, "%")
	}
	formated := fmt.Sprintf("%s/%d", deviceUsage, s.config.GetFormatPercent())

	// status
	status := "Done"
	usage, ok := utils.Str2Int(strings.TrimPrefix(deviceUsage, " "))
	if !ok {
		return errno.ERR_INVALID_DEVICE_USAGE.
			F("device usage: %s", deviceUsage)
	}
	if usage == 0 {
		status = "Mounting"
	} else if len(*s.containerStatus) > 1 && !strings.Contains(*s.containerStatus, "Exited") {
		status = "Formatting"
	} else if usage < s.config.GetFormatPercent() {
		status = "Pulling image"
	}

	id := fmt.Sprintf("%s:%s", host, device)
	setFormatStatus(s.memStorage, id, FormatStatus{
		Host:       host,
		Device:     device,
		MountPoint: mountPoint,
		Formatted:  formated,
		Status:     status,
	})
	return nil
}

func NewGetFormatStatusTask(curveadm *cli.CurveAdm, fc *configure.FormatConfig) (*task.Task, error) {
	host := fc.GetHost()
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	device := fc.GetDevice()
	subname := fmt.Sprintf("host=%s device=%s", fc.GetHost(), fc.GetDevice())
	t := task.NewTask("Get Format Status", subname, hc.GetSSHConfig())

	// add step to task
	var deviceUsage, containerStatus string
	containerName := device2ContainerName(device)
	t.AddStep(&step.ShowDiskFree{
		Files:       []string{device},
		Format:      "pcent",
		Out:         &deviceUsage,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &containerStatus,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2FormatStatus{
		config:          fc,
		deviceUsage:     &deviceUsage,
		containerStatus: &containerStatus,
		containerName:   containerName,
		memStorage:      curveadm.MemStorage(),
	})

	return t, nil
}

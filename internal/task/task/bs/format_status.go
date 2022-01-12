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
	"github.com/opencurve/curveadm/internal/configure/format"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2FormatStatus struct {
		config          *format.FormatConfig
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
		Status     string // Idle, Formating
	}
)

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
	formated := fmt.Sprintf("%s/%d", deviceUsage, s.config.GetUsagePercent())

	// status
	status := "Idle"
	if len(*s.containerStatus) > 1 {
		status = "Formatting"
	}

	id := fmt.Sprintf("%s:%s", host, device)
	s.memStorage.Set(id, FormatStatus{
		Host:       host,
		Device:     device,
		MountPoint: mountPoint,
		Formatted:  formated,
		Status:     status,
	})
	return nil
}

func NewGetFormatStatusTask(curveadm *cli.CurveAdm, fc *format.FormatConfig) (*task.Task, error) {
	device := fc.GetDevice()
	subname := fmt.Sprintf("host=%s device=%s", fc.GetHost(), fc.GetDevice())
	t := task.NewTask("Get Format Status", subname, fc.GetSSHConfig())

	// add step
	var deviceUsage, containerStatus string
	containerName := device2ContainerName(device)
	t.AddStep(&step.ShowDiskFree{
		Files:         []string{device},
		Format:        "pcent",
		Out:           &deviceUsage,
		ExecWithSudo:  false,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.Status}}'",
		Filter:        fmt.Sprintf("name=%s", containerName),
		Out:           &containerStatus,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
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

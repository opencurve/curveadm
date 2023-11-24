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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func checkTargetDaemonExist(containerId *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*containerId) == 0 {
			return task.ERR_TASK_DONE
		}
		return nil
	}
}

func NewStopTargetDaemonTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_TARGET_OPTIONS).(TargetOption)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s", options.Host)
	t := task.NewTask("Stop Target Daemon", subname, hc.GetSSHConfig(), hc.GetHttpConfig())

	// add step
	var containerId string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}}'",
		Filter:      fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkTargetDaemonExist(&containerId),
	})
	t.AddStep(&step.StopContainer{
		ContainerId: DEFAULT_TGTD_CONTAINER_NAME,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.RemoveContainer{
		ContainerId: DEFAULT_TGTD_CONTAINER_NAME,
		ExecOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

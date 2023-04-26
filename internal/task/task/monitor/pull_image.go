/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-04-19
* Author: wanghai (SeanHai)
 */

package monitor

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func NewPullImageTask(curveadm *cli.CurveAdm, cfg *configure.MonitorConfig) (*task.Task, error) {
	image := cfg.GetImage()
	host := cfg.GetHost()
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	subname := fmt.Sprintf("host=%s image=%s", host, image)
	t := task.NewTask("Pull Image", subname, hc.GetSSHConfig())
	// add step to task
	t.AddStep(&step.PullImage{
		Image:       image,
		ExecOptions: curveadm.ExecOptions(),
	})
	return t, nil
}

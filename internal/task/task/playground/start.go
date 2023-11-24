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
 * Created Date: 2022-11-07
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"fmt"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func wait(seconds int) step.LambdaType {
	return func(ctx *context.Context) error {
		time.Sleep(time.Duration(seconds) * time.Second)
		return nil
	}
}

func NewStartPlaygroundTask(curveadm *cli.CurveAdm, cfg *configure.PlaygroundConfig) (*task.Task, error) {
	// new task
	subname := fmt.Sprintf("kind=%s name=%s", cfg.GetKind(), cfg.GetName())
	t := task.NewTask("Start Playground", subname, nil, nil)

	// add step to task
	containerId := cfg.GetName()
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		ExecOptions: execOptions(curveadm),
	})
	t.AddStep(&step.Lambda{
		Lambda: wait(60),
	})

	return t, nil
}

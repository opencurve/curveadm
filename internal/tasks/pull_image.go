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

package tasks

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
)

type (
	step2PullImage struct{}
)

func (s *step2PullImage) Execute(ctx *task.Context) error {
	_, err := ctx.Module().SshShell("sudo docker pull %s", ctx.Config().GetContainerImage())
	return err
}

func (s *step2PullImage) Rollback(ctx *task.Context) {}

func NewPullImageTask(curvradm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	subname := fmt.Sprintf("host=%s image=%s", dc.GetHost(), dc.GetContainerImage())
	t := task.NewTask("Pull Image", subname, dc)
	t.AddStep(&step2PullImage{})
	return t, nil
}

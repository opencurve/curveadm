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
 * Created Date: 2022-02-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	KEY_DELETE_TID = "DELETE_TID"
)

type (
	step2CheckTgtdStatus struct{ output *string }
)

// check target daemon status
func (s *step2CheckTgtdStatus) Execute(ctx *context.Context) error {
	output := *s.output
	items := strings.Split(output, " ")
	if len(items) < 2 || !strings.HasPrefix(items[1], "Up") {
		return fmt.Errorf("Target daemon not running")
	}

	return nil
}

func NewDeleteTargetTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	tid := curveadm.MemStorage().Get(KEY_DELETE_TID).(string)
	subname := fmt.Sprintf("hostname=%s tid=%s", cc.GetHost(), tid)
	t := task.NewTask("Delete Target", subname, cc.GetSSHConfig())

	// add step
	var output string
	containerId := DEFAULT_TGTD_CONTAINER_NAME

	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}} {{.Status}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:           &output,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2CheckTgtdStatus{
		output: &output,
	})
	t.AddStep(&step.ContainerExec{
		ContainerId:   &containerId,
		Command:       fmt.Sprintf("tgtadm --lld iscsi --mode target --op delete --tid %s", tid),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})

	return t, nil
}

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
 * Created Date: 2022-08-16
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2FormatPlaygroundStatus struct {
		status     *string
		playground storage.Playground
		memStorage *utils.SafeMap
	}

	PlaygroundStatus struct {
		Id         string
		Name       string
		CreateTime string
		Status     string
	}
)

func setPlaygroundStatus(memStorage *utils.SafeMap, id string, status PlaygroundStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]PlaygroundStatus{}
		v := kv.Get(comm.KEY_ALL_PLAYGROUNDS_STATUS)
		if v != nil {
			m = v.(map[string]PlaygroundStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_PLAYGROUNDS_STATUS, m)
		return nil
	})
}

func (s *step2FormatPlaygroundStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if len(status) == 0 { // container losed
		status = comm.PLAYGROUDN_STATUS_LOSED
	}

	playground := s.playground
	id := utils.Atoa(playground.Id)
	setPlaygroundStatus(s.memStorage, id, PlaygroundStatus{
		Id:         id,
		Name:       playground.Name,
		CreateTime: playground.CreateTime.Format("2006-01-02 15:04:05"),
		Status:     status,
	})
	return nil
}

func NewGetPlaygroundStatusTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	// new task
	playground := v.(storage.Playground)
	subname := fmt.Sprintf("id=%d name=%s", playground.Id, playground.Name)
	t := task.NewTask("Get Playground Status", subname, nil, nil)

	// add step to task
	var status string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", playground.Name),
		Out:         &status,
		ExecOptions: execOptions(curveadm),
	})
	t.AddStep(&step2FormatPlaygroundStatus{
		status:     &status,
		playground: playground,
		memStorage: curveadm.MemStorage(),
	})

	return t, nil
}

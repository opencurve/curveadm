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
 * Created Date: 2022-07-31
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	MISSING_CLIENT_CONFIG = "-"
)

type (
	step2FormatClientStatus struct {
		client      storage.Client
		containerId string
		status      *string
		cfgPath     *string
		memStorage  *utils.SafeMap
	}

	ClientStatus struct {
		Id          string
		Host        string
		Kind        string
		ContainerId string
		Status      string
		AuxInfo     string
		CfgPath     string
	}
)

func dumpCfg(curveadm *cli.CurveAdm, id string, cfgPath *string) step.LambdaType {
	return func(ctx *context.Context) error {
		*cfgPath = MISSING_CLIENT_CONFIG
		cfgs, err := curveadm.Storage().GetClientConfig(id)
		if err != nil {
			return errno.ERR_SELECT_CLIENT_CONFIG_FAILED.E(err)
		} else if len(cfgs) == 0 {
			return nil
		}

		data := cfgs[0].Data
		path := utils.RandFilename("/tmp")
		err = utils.WriteFile(path, data, 0644)
		if err != nil {
			return errno.ERR_WRITE_FILE_FAILED.E(err)
		}

		*cfgPath = path
		return nil
	}
}

// TODO(P0): init client status
func setClientStatus(memStorage *utils.SafeMap, id string, status ClientStatus) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]ClientStatus{}
		v := kv.Get(comm.KEY_ALL_CLIENT_STATUS)
		if v != nil {
			m = v.(map[string]ClientStatus)
		}
		m[id] = status
		kv.Set(comm.KEY_ALL_CLIENT_STATUS, m)
		return nil
	})
}

func (s *step2FormatClientStatus) Execute(ctx *context.Context) error {
	status := *s.status
	if len(status) == 0 { // container losed
		status = comm.CLIENT_STATUS_LOSED
	}

	client := s.client
	id := client.Id
	setClientStatus(s.memStorage, id, ClientStatus{
		Id:          client.Id,
		Host:        client.Host,
		Kind:        client.Kind,
		ContainerId: s.containerId,
		Status:      status,
		AuxInfo:     client.AuxInfo,
		CfgPath:     *s.cfgPath,
	})
	return nil
}

func NewGetClientStatusTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	client := v.(storage.Client)
	hc, err := curveadm.GetHost(client.Host)
	if err != nil {
		return nil, err
	}

	containerId := client.ContainerId
	subname := fmt.Sprintf("host=%s kind=%s containerId=%s",
		hc.GetHost(), client.Kind, tui.TrimContainerId(containerId))
	t := task.NewTask("Get Client Status", subname, hc.GetSSHConfig())

	// add step
	var status, cfgPath string
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.Status}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &status,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: dumpCfg(curveadm, client.Id, &cfgPath),
	})
	t.AddStep(&step2FormatClientStatus{
		client:      client,
		containerId: containerId,
		status:      &status,
		cfgPath:     &cfgPath,
		memStorage:  curveadm.MemStorage(),
	})

	return t, nil
}

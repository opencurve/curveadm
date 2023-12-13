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
 * Created Date: 2022-08-06
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93_

package common

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

func updateTopology(curveadm *cli.CurveAdm) step.LambdaType {
	return func(ctx *context.Context) error {
		topology := curveadm.MemStorage().Get(comm.KEY_NEW_TOPOLOGY_DATA).(string)
		err := curveadm.Storage().SetClusterTopology(curveadm.ClusterId(), topology)
		if err != nil {
			return errno.ERR_UPDATE_CLUSTER_TOPOLOGY_FAILED.E(err)
		}
		return nil
	}
}

func NewUpdateTopologyTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	t := task.NewTask("Update Topology", "", nil)

	// add step to task
	t.AddStep(&step.Lambda{
		Lambda: updateTopology(curveadm),
	})

	return t, nil
}

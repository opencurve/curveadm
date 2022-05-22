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
 * Created Date: 2022-05-20
 * Author: Jingli Chen (Wine93)
 */

package common

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errors"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
)

type Step struct {
	Type          int
	DeployConfigs []*topology.DeployConfig
}

const (
	// task type
	PULL_IMAGE = iota
	CREATE_CONTAINER
	SYNC_CONFIG
	START_ETCD
	START_MDS
	START_CHUNKSERVER
	START_SNAPSHOTCLONE
	START_METASEREVR
	CREATE_PHYSICAL_POOL
	CREATE_LOGICAL_POOL
	BALANCE_LEADER
	BACKUP_ETCD_DATA
	STOP_SERVICE
	CLEAN_SERVICE_CONTAINER
)

var (
	ERR_EMPTY_TOPOLOGY = fmt.Errorf("empty topology")
	ERR_NO_SERVICE     = fmt.Errorf("no service")
)

func FilterDeployConfig(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, role string) []*topology.DeployConfig {
	options := topology.FilterOption{Id: "*", Role: role, Host: "*"}
	return curveadm.FilterDeployConfig(dcs, options)
}

func ExecDeploy(curveadm *cli.CurveAdm, steps []Step) error {
	for _, step := range steps {
		taskType := tasks.UNKNOWN
		dcs := step.DeployConfigs
		switch step.Type {
		case PULL_IMAGE:
			taskType = tasks.PULL_IMAGE
		case CREATE_CONTAINER:
			taskType = tasks.CREATE_CONTAINER
		case SYNC_CONFIG:
			taskType = tasks.SYNC_CONFIG
		case START_ETCD:
			taskType = tasks.START_SERVICE
		case START_MDS:
			taskType = tasks.START_SERVICE
		case START_CHUNKSERVER:
			taskType = tasks.START_SERVICE
		case START_SNAPSHOTCLONE:
			taskType = tasks.START_SERVICE
		case START_METASEREVR:
			taskType = tasks.START_SERVICE
		case CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_PHYSICAL_POOL)
			taskType = tasks.CREATE_POOL
		case CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_LOGICAL_POOL)
			taskType = tasks.CREATE_POOL
		case BALANCE_LEADER:
			taskType = tasks.BALANCE_LEADER
		case BACKUP_ETCD_DATA:
			taskType = tasks.BACKUP_ETCD_DATA
		case STOP_SERVICE:
			taskType = tasks.STOP_SERVICE
		case CLEAN_SERVICE_CONTAINER:
			curveadm.MemStorage().Set(task.KEY_CLEAN_ITEMS, []string{ task.ITEM_CONTAINER })
			taskType = tasks.CLEAN_SERVICE
		}

		if len(dcs) == 0 {
			return errors.ERR_CONFIGURE_NO_SERVICE
		}

		err := tasks.ExecTasks(taskType, curveadm, dcs)
		if err != nil {
			return err
		}

		// execute tasks success
		curveadm.WriteOut("\n")
	}
	return nil
}

func DiffTopology(oldData, newData string) ([]topology.TopologyDiff, error) {
	if len(oldData) == 0 {
		return nil, ERR_EMPTY_TOPOLOGY
	} else if dcs, err := topology.ParseTopology(oldData); err != nil {
		return nil, err
	} else if len(dcs) == 0 {
		return nil, ERR_NO_SERVICE
	}
	return topology.DiffTopology(oldData, newData)
}

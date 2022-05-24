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
	"sort"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errors"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
)

type DeployStep struct {
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
	STOP_ETCD
	STOP_MDS
	STOP_CHUNKSERVER
	STOP_SNAPSHOTCLONE
	STOP_METASEREVR
	CLEAN_SERVICE_CONTAINER
	BACKUP_ETCD_DATA
)

var (
	ERR_EMPTY_TOPOLOGY = fmt.Errorf("empty topology")
	ERR_NO_SERVICE     = fmt.Errorf("no service")
)

func ExecDeploy(curveadm *cli.CurveAdm, steps []DeployStep) error {
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
		case START_ETCD, START_MDS, START_CHUNKSERVER, START_SNAPSHOTCLONE, START_METASEREVR:
			taskType = tasks.START_SERVICE
		case CREATE_PHYSICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_PHYSICAL_POOL)
			taskType = tasks.CREATE_POOL
		case CREATE_LOGICAL_POOL:
			curveadm.MemStorage().Set(task.KEY_POOL_TYPE, task.TYPE_LOGICAL_POOL)
			taskType = tasks.CREATE_POOL
		case BALANCE_LEADER:
			taskType = tasks.BALANCE_LEADER
		case STOP_ETCD, STOP_MDS, STOP_CHUNKSERVER, STOP_SNAPSHOTCLONE, STOP_METASEREVR:
			taskType = tasks.STOP_SERVICE
		case CLEAN_SERVICE_CONTAINER:
			curveadm.MemStorage().Set(task.KEY_RECYCLE, false)
			curveadm.MemStorage().Set(task.KEY_CLEAN_ITEMS, []string{task.ITEM_CONTAINER})
			taskType = tasks.CLEAN_SERVICE
		case BACKUP_ETCD_DATA:
			taskType = tasks.BACKUP_ETCD_DATA
		}

		if len(dcs) == 0 {
			return errors.ERR_CONFIGURE_NO_SERVICE
		}

		err := tasks.ExecTasks(taskType, curveadm, dcs)
		if err != nil {
			return err
		}

		// execute tasks success
		curveadm.WriteOutln("")
	}
	return nil
}

func FilterDeployConfig(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, role string) []*topology.DeployConfig {
	options := topology.FilterOption{Id: "*", Role: role, Host: "*"}
	return curveadm.FilterDeployConfig(dcs, options)
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

func ParseDiff(diffs []topology.TopologyDiff) (dcs4add, dcs4del, dcs4change []*topology.DeployConfig) {
	for _, diff := range diffs {
		diffType := diff.DiffType
		if diffType == topology.DIFF_ADD {
			dcs4add = append(dcs4add, diff.DeployConfig)
		} else if diffType == topology.DIFF_DELETE {
			dcs4del = append(dcs4del, diff.DeployConfig)
		} else if diffType == topology.DIFF_CHANGE {
			dcs4change = append(dcs4change, diff.DeployConfig)
		}
	}
	return
}

func IsSameRole(dcs []*topology.DeployConfig) bool {
	role := dcs[0].GetRole()
	for _, dc := range dcs {
		if dc.GetRole() != role {
			return false
		}
	}
	return true
}

// we should sort the "dcs" for generate correct zone number
func SortDeployConfigs(dcs []*topology.DeployConfig) {
	sort.Slice(dcs, func(i, j int) bool {
		dc1, dc2 := dcs[i], dcs[j]
		if dc1.GetRole() == dc2.GetRole() {
			if dc1.GetHostSequence() == dc2.GetHostSequence() {
				return dc1.GetReplicaSequence() < dc2.GetReplicaSequence()
			}
			return dc1.GetHostSequence() < dc2.GetHostSequence()
		}
		return dc1.GetRole() < dc2.GetRole()
	})
}

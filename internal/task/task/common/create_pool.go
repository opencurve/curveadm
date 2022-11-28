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

package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/build"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

type step2SetClusterPool struct {
	curveadm    *cli.CurveAdm
	clusterPool string
	storage     *storage.Storage
}

func getClusterPool(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (configure.CurveClusterTopo, error) {
	poolset := curveadm.MemStorage().Get(comm.POOLSET).(string)
	diskType := curveadm.MemStorage().Get(comm.POOLSET_DISK_TYPE).(string)
	oldPool := configure.CurveClusterTopo{}
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return oldPool, err
	}

	// 1) generate a new default pool
	data := curveadm.ClusterPoolData()
	if len(data) == 0 {
		return configure.GenerateDefaultClusterPool(dcs, poolset, diskType)
	}

	// 2) OR change old pool and return it
	err = json.Unmarshal([]byte(data), &oldPool)
	if err != nil {
		return oldPool, err
	}
	pool, err := configure.GenerateDefaultClusterPool(dcs, poolset, diskType)
	if err != nil {
		return pool, err
	}

	// NOTE: curveadm gurantee oldPool and pool has same servers
	for i, server := range pool.Servers {
		oldPool.Servers[i].InternalIp = server.InternalIp
		oldPool.Servers[i].InternalPort = server.InternalPort
		oldPool.Servers[i].ExternalIp = server.ExternalIp
		oldPool.Servers[i].ExternalPort = server.ExternalPort
	}
	if dc.GetKind() == topology.KIND_CURVEBS {
		for i, pool := range pool.LogicalPools {
			oldPool.LogicalPools[i].Copysets = pool.Copysets
		}
	}

	return oldPool, err
}

func prepare(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (clusterPoolJson, clusterMDSAddrs string, err error) {
	// 1. get origin cluster pool
	var clusterPool configure.CurveClusterTopo
	clusterPool, err = getClusterPool(curveadm, dc)
	if err != nil {
		return
	}

	// 2. scale out cluster or migrate servers
	if curveadm.MemStorage().Get(comm.KEY_SCALE_OUT_CLUSTER) != nil { // scale out cluster
		dcs := curveadm.MemStorage().Get(comm.KEY_SCALE_OUT_CLUSTER).([]*topology.DeployConfig)
		poolset := curveadm.MemStorage().Get(comm.POOLSET).(string)
		diskType := curveadm.MemStorage().Get(comm.POOLSET_DISK_TYPE).(string)
		configure.ScaleOutClusterPool(&clusterPool, dcs, poolset, diskType)
	} else if curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS) != nil { // migrate servers
		migrates := curveadm.MemStorage().Get(comm.KEY_MIGRATE_SERVERS).([]*configure.MigrateServer)
		configure.MigrateClusterServer(&clusterPool, migrates)
	}

	// 3. encode cluster pool to json string
	var bytes []byte
	bytes, err = json.Marshal(clusterPool)
	if err != nil {
		return
	}
	clusterPoolJson = string(bytes)

	// cluster MDS address
	clusterMDSAddrs, err = dc.GetVariables().Get("cluster_mds_addr")
	clusterMDSAddrs = strings.Replace(clusterMDSAddrs, ",", " ", -1)
	return
}

func checkWaitMDSElectionSuccess(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_WAIT_MDS_ELECTION_SUCCESS_TIMEOUT
		}
		return nil
	}
}

func checkChunkserverOnline(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_WAIT_ALL_CHUNKSERVERS_ONLINE_TIMEOUT
		}
		return nil
	}
}

func genCreatePoolCommand(dc *topology.DeployConfig, pooltype, poolJSONPath string) string {
	layout := dc.GetProjectLayout()
	toolsBinaryPath := layout.ToolsBinaryPath
	if dc.GetKind() == topology.KIND_CURVEFS {
		// for curvefs, the default topology json path is current directory's topology.json
		return fmt.Sprintf("%s create-topology", toolsBinaryPath)
	}

	return fmt.Sprintf("%s -op=create_%s -cluster_map=%s",
		toolsBinaryPath, pooltype, poolJSONPath)
}

func checkCreatePoolStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if !*success {
			return errno.ERR_CREATE_LOGICAL_POOL_FAILED.S(*out)
		}
		return nil
	}
}

func (s *step2SetClusterPool) Execute(ctx *context.Context) error {
	curveadm := s.curveadm
	topology := curveadm.ClusterTopologyData()
	value := curveadm.MemStorage().Get(comm.KEY_NEW_TOPOLOGY_DATA)
	if value != nil {
		topology = value.(string)
	}

	err := s.storage.SetClusterPool(curveadm.ClusterId(), topology, s.clusterPool)
	if err != nil {
		return errno.ERR_UPDATE_CLUSTER_POOL_FAILED.E(err)
	}
	return nil
}

func NewCreateTopologyTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if curveadm.IsSkip(dc) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	hc, err := curveadm.GetHost(dc.GetHost())
	if err != nil {
		return nil, err
	}

	// new task
	pooltype := curveadm.MemStorage().Get(comm.KEY_CREATE_POOL_TYPE).(string)
	name := utils.Choose(pooltype == comm.POOL_TYPE_LOGICAL, "Create Logical Pool",
		"Create Physical Pool")
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask(name, subname, hc.GetSSHConfig())

	// add step to task
	var success bool
	var out string
	host, role := dc.GetHost(), dc.GetRole()
	layout := dc.GetProjectLayout()
	poolJSONPath := fmt.Sprintf("%s/topology.json", layout.ToolsConfDir)
	waitScript := scripts.SCRIPT_WAIT
	waitScriptPath := fmt.Sprintf("%s/wait.sh", layout.ToolsBinDir)
	clusterPoolJson, clusterMDSAddrs, err := prepare(curveadm, dc)
	if err != nil {
		return nil, err
	}
	build.DEBUG(build.DEBUG_CREATE_POOL,
		build.Field{"pool json", clusterPoolJson})

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      `"{{.ID}}"`,
		Filter:      fmt.Sprintf("id=%s", containerId),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkContainerExist(host, role, containerId, &out),
	})
	t.AddStep(&step.InstallFile{ // install curvebs/curvefs topology
		ContainerId:       &containerId,
		ContainerDestPath: poolJSONPath,
		Content:           &clusterPoolJson,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install wait script
		ContainerId:       &containerId,
		ContainerDestPath: waitScriptPath,
		Content:           &waitScript,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{ // wait mds leader election success
		ContainerId: &containerId,
		Command:     fmt.Sprintf("bash %s %s", waitScriptPath, clusterMDSAddrs),
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkWaitMDSElectionSuccess(&success, &out),
	})

	if dc.GetKind() == topology.KIND_CURVEBS && pooltype == comm.POOL_TYPE_LOGICAL {
		waitChunkserversScript := scripts.SCRIPT_WAIT_CHUNKSERVERS
		waitChunkserversScriptPath := fmt.Sprintf("%s/wait_chunkservers.sh", layout.ToolsBinDir)
		t.AddStep(&step.InstallFile{ // install wait_chunkservers script
			ContainerId:       &containerId,
			ContainerDestPath: waitChunkserversScriptPath,
			Content:           &waitChunkserversScript,
			ExecOptions:       curveadm.ExecOptions(),
		})
		t.AddStep(&step.ContainerExec{ // wait all chunkservers online before create logical pool
			ContainerId: &containerId,
			Command:     fmt.Sprintf("bash %s", waitChunkserversScriptPath),
			Success:     &success,
			Out:         &out,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step.Lambda{
			Lambda: checkChunkserverOnline(&success, &out),
		})
	}
	t.AddStep(&step.ContainerExec{ // create topology
		ContainerId: &containerId,
		Success:     &success,
		Out:         &out,
		Command:     genCreatePoolCommand(dc, pooltype, poolJSONPath),
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkCreatePoolStatus(&success, &out),
	})
	if pooltype == comm.POOL_TYPE_LOGICAL {
		t.AddStep(&step2SetClusterPool{
			curveadm:    curveadm,
			clusterPool: clusterPoolJson,
			storage:     curveadm.Storage(),
		})
	}

	return t, nil
}

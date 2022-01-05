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
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/pool"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	KEY_POOL_TYPE = "POOL_TYPE"

	TYPE_LOGICAL_POOL  = "logicalpool"
	TYPE_PHYSICAL_POOL = "physicalpool"
)

func prepare(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (clusterPool, clusterMDSAddrs string, err error) {
	clusterPool, err = pool.GenerateClusterPool(curveadm.ClusterTopologyData())
	if err != nil {
		return
	}

	clusterMDSAddrs, err = dc.GetVariables().Get("cluster_mds_addr")
	clusterMDSAddrs = strings.Replace(clusterMDSAddrs, ",", " ", -1)
	return
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

func NewCreateTopologyTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	pooltype := curveadm.MemStorage().Get(KEY_POOL_TYPE).(string)
	name := utils.Choose(pooltype == TYPE_LOGICAL_POOL, "Create Logical Pool",
		"Create Physical Pool")
	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask(name, subname, dc.GetSSHConfig())

	// add step
	layout := dc.GetProjectLayout()
	poolJSONPath := fmt.Sprintf("%s/topology.json", layout.ToolsConfDir)
	waitScript := scripts.SCRIPT_WAIT
	waitScriptPath := fmt.Sprintf("%s/wait.sh", layout.ToolsBinDir)
	clusterPool, clusterMDSAddrs, err := prepare(curveadm, dc)
	if err != nil {
		return nil, err
	}

	t.AddStep(&step.InstallFile{ // install curvebs/curvefs topology
		ContainerId:       &containerId,
		ContainerDestPath: poolJSONPath,
		Content:           &clusterPool,
		ExecWithSudo:      true,
		ExecInLocal:       false,
	})
	t.AddStep(&step.InstallFile{ // install wait script
		ContainerId:       &containerId,
		ContainerDestPath: waitScriptPath,
		Content:           &waitScript,
		ExecWithSudo:      true,
		ExecInLocal:       false,
	})
	t.AddStep(&step.ContainerExec{ // wait mds leader election success
		ContainerId:  &containerId,
		Command:      fmt.Sprintf("bash %s %s", waitScriptPath, clusterMDSAddrs),
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step.ContainerExec{ // create topology
		ContainerId:  &containerId,
		Command:      genCreatePoolCommand(dc, pooltype, poolJSONPath),
		ExecWithSudo: true,
		ExecInLocal:  false,
	})

	return t, nil
}

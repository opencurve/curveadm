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
	"os"
	"path/filepath"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

type (
	step2CreateCurveFSTopology struct {
		tempDir     string
		clusterName string
		data        string
	}
	step2BuildTopology struct{ containerId string }
)

func (s *step2CreateCurveFSTopology) Execute(ctx *task.Context) error {
	topo, err := configure.GenerateCurveFSTopology(s.data)
	if err != nil {
		return err
	}

	file, err := os.CreateTemp(s.tempDir, fmt.Sprintf("%s_CurveFSTopology_*", s.clusterName))
	if err != nil {
		return err
	}
	defer file.Close()

	config := ctx.Config()
	localPath := file.Name()
	remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(localPath))
	containerDstPath := fmt.Sprintf("%s/tools/conf/topology.json", config.GetCurveFSPrefix())
	ctx.Register().Set(KEY_LOCAL_PATH, localPath)
	ctx.Register().Set(KEY_REMOTE_PATH, remotePath)
	ctx.Register().Set(KEY_CONTAINER_DST_PATH, containerDstPath)

	_, err = file.WriteString(topo)
	return err
}

func (s *step2CreateCurveFSTopology) Rollback(ctx *task.Context) {
}

// TODO(@Wine93): refactor build topology
func (s *step2BuildTopology) Execute(ctx *task.Context) error {
	toolsBinray := fmt.Sprintf("%s/tools/sbin/curvefs_tool", ctx.Config().GetCurveFSPrefix())
	confPath := fmt.Sprintf("%s/tools/conf/tools.conf", ctx.Config().GetCurveFSPrefix())
	cmd := fmt.Sprintf("sudo docker exec %s %s build-topology --confPath=%s",
		s.containerId, toolsBinray, confPath)
	_, err := ctx.Module().SshShell(cmd)
	return err
}

func (s *step2BuildTopology) Rollback(ctx *task.Context) {
}

func NewCreateTopologyTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Create Topology", subname, dc)

	// (1): rendering tool config
	t.AddStep(&step2InitSyncConfig{
		tempDir:   curveadm.TempDir(),
		serviceId: fmt.Sprintf("%d_tools_%s", curveadm.ClusterId(), dc.GetId()),
		// /usr/local/curvefs/conf/tools.conf
		containerSrcPath: fmt.Sprintf("%s/conf/tools.conf", dc.GetCurveFSPrefix()),
		// /usr/local/curvefs/tools/conf/tools.conf
		containerDstPath: fmt.Sprintf("%s/tools/conf/tools.conf", dc.GetCurveFSPrefix()),
	})
	t.AddStep(&step2CopyFileFromRemote{containerId: containerId})
	t.AddStep(&step2RenderingConfig{})
	t.AddStep(&step2CopyFileToRemote{containerId: containerId})

	// (2): generate curvefs topology
	t.AddStep(&step2CreateCurveFSTopology{
		tempDir:     curveadm.TempDir(),
		clusterName: curveadm.ClusterName(),
		data:        curveadm.ClusterTopologyData(),
	})
	t.AddStep(&step2CopyFileToRemote{containerId: containerId})

	// (3): create topology
	t.AddStep(&step2BuildTopology{containerId: containerId})

	return t, nil
}

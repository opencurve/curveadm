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

// __SIGN_BY_WINE93__

package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR = topology.LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR
	LAYOUT_CURVEBS_COPYSETS_DIR       = topology.LAYOUT_CURVEBS_COPYSETS_DIR
	LAYOUT_CURVEBS_RECYCLER_DIR       = topology.LAYOUT_CURVEBS_RECYCLER_DIR
	METAFILE_CHUNKSERVER_ID           = topology.METAFILE_CHUNKSERVER_ID

	SIGNATURE_CONTAINER_REMOVED = "No such container"
)

type (
	step2RecycleChunk struct {
		dc                *topology.DeployConfig
		clean             map[string]bool
		recycleScriptPath string
		execOptions       module.ExecOptions
	}

	step2CleanContainer struct {
		config      *topology.DeployConfig
		serviceId   string
		containerId string
		storage     *storage.Storage
		execOptions module.ExecOptions
	}

	step2CleanDiskChunkServerId struct {
		serviceId   string
		storage     *storage.Storage
		execOptions module.ExecOptions
	}
)

func (s *step2RecycleChunk) Execute(ctx *context.Context) error {
	dc := s.dc
	if !s.clean[comm.CLEAN_ITEM_DATA] {
		return nil
	} else if dc.GetRole() != topology.ROLE_CHUNKSERVER {
		return nil
	} else if len(dc.GetDataDir()) == 0 {
		return nil
	}

	dataDir := dc.GetDataDir()
	copysetsDir := fmt.Sprintf("%s/%s", dataDir, LAYOUT_CURVEBS_COPYSETS_DIR)
	recyclerDir := fmt.Sprintf("%s/%s", dataDir, LAYOUT_CURVEBS_RECYCLER_DIR)
	source := fmt.Sprintf("'%s %s'", copysetsDir, recyclerDir)
	dest := fmt.Sprintf("%s/%s", dataDir, LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR)
	chunkSize := strconv.Itoa(bs.DEFAULT_CHUNKFILE_SIZE + bs.DEFAULT_CHUNKFILE_HEADER_SIZE)
	cmd := ctx.Module().Shell().BashScript(s.recycleScriptPath, source, dest, chunkSize)
	_, err := cmd.Execute(s.execOptions)
	if err != nil {
		errno.ERR_RUN_SCRIPT_FAILED.E(err)
	}
	return nil
}

func (s *step2CleanContainer) Execute(ctx *context.Context) error {
	containerId := s.containerId
	if len(containerId) == 0 { // container not created
		return nil
	} else if containerId == comm.CLEANED_CONTAINER_ID { // container has removed
		return nil
	}

	cli := ctx.Module().DockerCli().RemoveContainer(s.containerId)
	out, err := cli.Execute(s.execOptions)

	// container has removed
	if err != nil && !strings.Contains(out, SIGNATURE_CONTAINER_REMOVED) {
		return err
	}
	return s.storage.SetContainId(s.serviceId, comm.CLEANED_CONTAINER_ID)
}

func (s *step2CleanDiskChunkServerId) Execute(ctx *context.Context) error {
	return s.storage.CleanDiskChunkServerId(s.serviceId)
}

func getCleanFiles(clean map[string]bool, dc *topology.DeployConfig, recycle bool) []string {
	files := []string{}
	for item := range clean {
		switch item {
		case comm.CLEAN_ITEM_LOG:
			files = append(files, dc.GetLogDir())
		case comm.CLEAN_ITEM_DATA:
			if dc.GetRole() != topology.ROLE_CHUNKSERVER || !recycle {
				files = append(files, dc.GetDataDir())
			} else {
				dataDir := dc.GetDataDir()
				copysetsDir := fmt.Sprintf("%s/%s", dataDir, LAYOUT_CURVEBS_COPYSETS_DIR)
				chunkserverIdMetafile := fmt.Sprintf("%s/%s", dataDir, METAFILE_CHUNKSERVER_ID)
				files = append(files, copysetsDir, chunkserverIdMetafile)
			}
		}
	}
	return files
}

func NewCleanServiceTask(curveadm *cli.CurveAdm, dc *topology.DeployConfig) (*task.Task, error) {
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
	only := curveadm.MemStorage().Get(comm.KEY_CLEAN_ITEMS).([]string)
	recycle := curveadm.MemStorage().Get(comm.KEY_CLEAN_BY_RECYCLE).(bool)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Service", subname, hc.GetSSHConfig())

	// add step to task
	clean := utils.Slice2Map(only)
	files := getCleanFiles(clean, dc, recycle) // directorys which need cleaned
	recyleScript := scripts.SCRIPT_RECYCLE
	recyleScriptPath := utils.RandFilename(TEMP_DIR)

	if dc.GetKind() == topology.KIND_CURVEBS {
		t.AddStep(&step.Scp{
			Content:     &recyleScript,
			RemotePath:  recyleScriptPath,
			Mode:        0777,
			ExecOptions: curveadm.ExecOptions(),
		})
		t.AddStep(&step2RecycleChunk{
			dc:                dc,
			clean:             clean,
			recycleScriptPath: recyleScriptPath,
			execOptions:       curveadm.ExecOptions(),
		})
	}
	t.AddStep(&step.RemoveFile{
		Files:       files,
		ExecOptions: curveadm.ExecOptions(),
	})
	if clean[comm.CLEAN_ITEM_CONTAINER] == true {
		t.AddStep(&step2CleanContainer{
			config:      dc,
			serviceId:   serviceId,
			containerId: containerId,
			storage:     curveadm.Storage(),
			execOptions: curveadm.ExecOptions(),
		})
	}
	t.AddStep(&step2CleanDiskChunkServerId{
		serviceId:   serviceId,
		storage:     curveadm.Storage(),
		execOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

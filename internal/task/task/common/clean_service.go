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
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
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
	KEY_RECYCLE     = "RECYCLE"
	KEY_CLEAN_ITEMS = "CLEAN_ITEMS"
	ITEM_LOG        = "log"
	ITEM_DATA       = "data"
	ITEM_CONTAINER  = "container"

	LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR = topology.LAYOUT_CURVEBS_CHUNKFILE_POOL_DIR
	LAYOUT_CURVEBS_COPYSETS_DIR       = topology.LAYOUT_CURVEBS_COPYSETS_DIR
	LAYOUT_CURVEBS_RECYCLER_DIR       = topology.LAYOUT_CURVEBS_RECYCLER_DIR
	METAFILE_CHUNKSERVER_ID           = topology.METAFILE_CHUNKSERVER_ID
)

type (
	step2RecycleChunk struct {
		dc                *topology.DeployConfig
		clean             map[string]bool
		recycleScriptPath string
		execWithSudo      bool
		execInLocal       bool
		execSudoAlias     string
	}

	step2CleanContainer struct {
		config        *topology.DeployConfig
		serviceId     string
		containerId   string
		storage       *storage.Storage
		execWithSudo  bool
		execInLocal   bool
		execSudoAlias string
	}
)

func (s *step2RecycleChunk) Execute(ctx *context.Context) error {
	dc := s.dc
	if !s.clean[ITEM_DATA] {
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
	chunk_size := strconv.Itoa(bs.DEFAULT_CHUNKFILE_SIZE + bs.DEFAULT_CHUNKFILE_HEADER_SIZE)
	cmd := ctx.Module().Shell().ExecScript(s.recycleScriptPath, source, dest, chunk_size)
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.execWithSudo,
		ExecInLocal:   s.execInLocal,
		ExecSudoAlias: s.execSudoAlias,
	})
	return err
}

func (s *step2CleanContainer) Execute(ctx *context.Context) error {
	containerId := s.containerId
	if containerId == "" { // container not created
		return nil
	} else if containerId == "-" { // container has removed
		return nil
	}

	cli := ctx.Module().DockerCli().RemoveContainer(s.containerId)
	out, err := cli.Execute(module.ExecOption{
		ExecWithSudo:  s.execWithSudo,
		ExecInLocal:   s.execInLocal,
		ExecSudoAlias: s.execSudoAlias,
	})

	// container has removed
	if err != nil && !strings.Contains(out, "No such container") {
		return err
	}
	return s.storage.SetContainId(s.serviceId, "-")
}

func getCleanFiles(clean map[string]bool, dc *topology.DeployConfig, recycle bool) []string {
	files := []string{}
	for item := range clean {
		switch item {
		case ITEM_LOG:
			files = append(files, dc.GetLogDir())
		case ITEM_DATA:
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
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	only := curveadm.MemStorage().Get(KEY_CLEAN_ITEMS).([]string)
	recycle := curveadm.MemStorage().Get(KEY_RECYCLE).(bool)
	subname := fmt.Sprintf("host=%s role=%s containerId=%s clean=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId), strings.Join(only, ","))
	t := task.NewTask("Clean Service", subname, dc.GetSSHConfig())

	// add step
	clean := utils.Slice2Map(only)
	files := getCleanFiles(clean, dc, recycle) // directorys which need cleaned
	recyleScript := scripts.SCRIPT_RECYCLE
	recyleScriptPath := utils.RandFilename(TEMP_DIR)

	t.AddStep(&step.InstallFile{
		Content:       &recyleScript,
		HostDestPath:  recyleScriptPath,
		Mode:          "777",
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2RecycleChunk{
		dc:                dc,
		clean:             clean,
		recycleScriptPath: recyleScriptPath,
		execWithSudo:      true,
		execInLocal:       false,
		execSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.RemoveFile{
		Files:         files,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	if clean[ITEM_CONTAINER] == true {
		t.AddStep(&step2CleanContainer{
			config:        dc,
			serviceId:     serviceId,
			containerId:   containerId,
			storage:       curveadm.Storage(),
			execWithSudo:  true,
			execInLocal:   false,
			execSudoAlias: curveadm.SudoAlias(),
		})
	}

	return t, nil
}

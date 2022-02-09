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

	"github.com/opencurve/curveadm/cli/cli"
	bs_client "github.com/opencurve/curveadm/internal/configure/client/bs"
	fs_client "github.com/opencurve/curveadm/internal/configure/client/fs"
	"github.com/opencurve/curveadm/internal/configure/format"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	comm "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/task/fs"
)

const (
	TYPE_CONFIG_DEPLOY int = iota
	TYPE_CONFIG_BS_CLIENT
	TYPE_CONFIG_FS_CLIENT
	TYPE_CONFIG_FORMAT
	TYPE_CONFIG_NULL
)

type configs struct {
	ctype  int
	length int
	dcs    []*topology.DeployConfig
	bccs   []*bs_client.ClientConfig
	fccs   []*fs_client.ClientConfig
	fcs    []*format.FormatConfig
}

const (
	// common
	PULL_IMAGE int = iota
	CREATE_CONTAINER
	SYNC_CONFIG
	START_SERVICE
	STOP_SERVICE
	RESTART_SERVICE
	CREATE_POOL
	GET_SERVICE_STATUS
	CLEAN_SERVICE
	SYNC_BINARY
	COLLECT_SERVICE
	// bs
	START_NEBD_SERVICE
	START_TARGET_DAEMON
	MAP_IMAGE
	UNMAP_IMAGE
	ADD_TARGET
	DELETE_TARGET
	LIST_TARGETS
	// fs
	MOUNT_FILESYSTEM
	UMOUNT_FILESYSTEM
	CHECK_MOUNT_STATUS
	FORMAT_CHUNKFILE_POOL
	GET_FORMAT_STATUS
	UNKNOWN // unknown
)

func newConfigs(configSlice interface{}) (*configs, error) {
	configs := &configs{
		dcs:  []*topology.DeployConfig{},
		bccs: []*bs_client.ClientConfig{},
		fccs: []*fs_client.ClientConfig{},
		fcs:  []*format.FormatConfig{},
	}
	switch configSlice.(type) {
	case []*topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = configSlice.([]*topology.DeployConfig)
		configs.length = len(configs.dcs)
	case []*bs_client.ClientConfig:
		configs.ctype = TYPE_CONFIG_BS_CLIENT
		configs.bccs = configSlice.([]*bs_client.ClientConfig)
		configs.length = len(configs.bccs)
	case []*fs_client.ClientConfig:
		configs.ctype = TYPE_CONFIG_FS_CLIENT
		configs.fccs = configSlice.([]*fs_client.ClientConfig)
		configs.length = len(configs.fccs)
	case []*format.FormatConfig:
		configs.ctype = TYPE_CONFIG_FORMAT
		configs.fcs = configSlice.([]*format.FormatConfig)
		configs.length = len(configs.fcs)
	case *topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = append(configs.dcs, configSlice.(*topology.DeployConfig))
		configs.length = 1
	case *bs_client.ClientConfig:
		configs.ctype = TYPE_CONFIG_BS_CLIENT
		configs.bccs = append(configs.bccs, configSlice.(*bs_client.ClientConfig))
		configs.length = 1
	case *fs_client.ClientConfig:
		configs.ctype = TYPE_CONFIG_FS_CLIENT
		configs.fccs = append(configs.fccs, configSlice.(*fs_client.ClientConfig))
		configs.length = 1
	case *format.FormatConfig:
		configs.ctype = TYPE_CONFIG_FORMAT
		configs.fcs = append(configs.fcs, configSlice.(*format.FormatConfig))
		configs.length = 1
	case nil:
		configs.ctype = TYPE_CONFIG_NULL
		configs.length = 1
	default:
		return nil, fmt.Errorf("unknown config type")
	}
	return configs, nil
}

func ExecTasks(taskType int, curveadm *cli.CurveAdm, configSlice interface{}) error {
	var t *task.Task
	var dc *topology.DeployConfig
	var bcc *bs_client.ClientConfig
	var fcc *fs_client.ClientConfig
	var fc *format.FormatConfig

	configs, err := newConfigs(configSlice)
	if err != nil {
		return err
	}

	tasks := NewTasks()
	option := ExecOption{
		Concurrency:   10,
		SilentSubBar:  false,
		SilentMainBar: false,
		SkipError:     false,
	}

	// add task into tasks
	pullImage := map[string]bool{}
	ctype := configs.ctype
	for i := 0; i < configs.length; i++ {
		// config type
		switch ctype {
		case TYPE_CONFIG_DEPLOY:
			dc = configs.dcs[i]
		case TYPE_CONFIG_BS_CLIENT:
			bcc = configs.bccs[i]
		case TYPE_CONFIG_FS_CLIENT:
			fcc = configs.fccs[i]
		case TYPE_CONFIG_FORMAT:
			fc = configs.fcs[i]
		case TYPE_CONFIG_NULL: // do nothing
		}

		// task type
		switch taskType {
		case PULL_IMAGE:
			// prevent reacheing docker hub pull rate limit
			if pullImage[dc.GetParentId()] == true {
				continue
			}
			t, err = comm.NewPullImageTask(curveadm, dc)
			pullImage[dc.GetParentId()] = true
		case CREATE_CONTAINER:
			t, err = comm.NewCreateContainerTask(curveadm, dc)
		case SYNC_CONFIG:
			t, err = comm.NewSyncConfigTask(curveadm, dc)
		case START_SERVICE:
			t, err = comm.NewStartServiceTask(curveadm, dc)
		case STOP_SERVICE:
			t, err = comm.NewStopServiceTask(curveadm, dc)
		case RESTART_SERVICE:
			t, err = comm.NewRestartServiceTask(curveadm, dc)
		case CREATE_POOL:
			t, err = comm.NewCreateTopologyTask(curveadm, dc)
		case GET_SERVICE_STATUS:
			option.SilentSubBar = true
			option.SkipError = true
			t, err = comm.NewGetServiceStatusTask(curveadm, dc)
		case CLEAN_SERVICE:
			t, err = comm.NewCleanServiceTask(curveadm, dc)
		case COLLECT_SERVICE:
			t, err = comm.NewCollectServiceTask(curveadm, dc)
		case SYNC_BINARY:
			//t, err = comm.NewSyncBinaryTask(curveadm, dc)
		case MOUNT_FILESYSTEM:
			option.SilentSubBar = true
			t, err = fs.NewMountFSTask(curveadm, fcc)
		case UMOUNT_FILESYSTEM:
			option.SilentSubBar = true
			t, err = fs.NewUmountFSTask(curveadm, fcc)
		case CHECK_MOUNT_STATUS:
			option.SilentMainBar = true
			option.SilentSubBar = true
			t, err = fs.NewGetMountStatusTask(curveadm, fcc)
		case FORMAT_CHUNKFILE_POOL:
			t, err = bs.NewFormatChunkfilePoolTask(curveadm, fc)
		case GET_FORMAT_STATUS:
			option.SilentSubBar = true
			option.SkipError = true
			t, err = bs.NewGetFormatStatusTask(curveadm, fc)
		case START_NEBD_SERVICE:
			option.SilentSubBar = true
			t, err = bs.NewStartNEBDServiceTask(curveadm, bcc)
		case START_TARGET_DAEMON:
			option.SilentSubBar = true
			t, err = bs.NewStartTargetDaemonTask(curveadm, bcc)
		case MAP_IMAGE:
			option.SilentSubBar = true
			t, err = bs.NewMapTask(curveadm, bcc)
		case UNMAP_IMAGE:
			option.SilentSubBar = true
			t, err = bs.NewUnmapTask(curveadm, bcc)
		case ADD_TARGET:
			option.SilentSubBar = true
			t, err = bs.NewAddTargetTask(curveadm, bcc)
		case DELETE_TARGET:
			option.SilentSubBar = true
			t, err = bs.NewDeleteTargetTask(curveadm, bcc)
		case LIST_TARGETS:
			option.SilentSubBar = true
			t, err = bs.NewListTargetsTask(curveadm, bcc)
		default:
			return fmt.Errorf("unknown task type %d", taskType)
		}

		if err != nil {
			return err
		}
		if ctype == TYPE_CONFIG_DEPLOY { // merge task status into one
			t.SetTid(dc.GetId())
			t.SetPtid(dc.GetParentId())
		}
		tasks.AddTask(t)
	}

	return tasks.Execute(option)
}

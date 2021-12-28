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
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/client"
	"github.com/opencurve/curveadm/internal/configure/format"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	comm "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/task/fs"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	TYPE_CONFIG_DEPLOY int = iota
	TYPE_CONFIG_CLIENT
	TYPE_CONFIG_FORMAT
	TYPE_CONFIG_NULL
)

type configs struct {
	ctype  int
	length int
	dcs    []*topology.DeployConfig
	ccs    []*client.ClientConfig
	fcs    []*format.FormatConfig
}

const (
	PULL_IMAGE int = iota // common
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
	MOUNT_FILESYSTEM // fs
	UMOUNT_FILESYSTEM
	CHECK_MOUNT_STATUS
	FORMAT_CHUNKFILE_POOL
	GET_FORMAT_STATUS
	UNKNOWN // unknown
)

/*
 * before:
 *   host=10.0.0.1 role=mds containerId=1863158e02a6
 *   host=10.0.0.2 role=metaserver containerId=0e6dcd718b85
 *
 * after:
 *   host=10.0.0.1  role=mds         containerId=1863158e02a6
 *   host=10.0.0.2  role=metaserver  containerId=0e6dcd718b85
 */
func prettyTasksSubName(ts []*task.Task) {
	lines := [][]interface{}{}
	for _, t := range ts {
		line := []interface{}{}
		for _, v := range strings.Split(t.Subname(), " ") {
			line = append(line, v)
		}
		lines = append(lines, line)
	}

	output := tui.FixedFormat(lines, 2)
	subnames := strings.Split(output, "\n")
	for i, t := range ts {
		t.SetSubName(subnames[i])
	}
}

func newConfigs(configSlice interface{}) (*configs, error) {
	configs := &configs{
		dcs: []*topology.DeployConfig{},
		ccs: []*client.ClientConfig{},
	}
	switch configSlice.(type) {
	case []*topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = configSlice.([]*topology.DeployConfig)
		configs.length = len(configs.dcs)
	case []*client.ClientConfig:
		configs.ctype = TYPE_CONFIG_CLIENT
		configs.ccs = configSlice.([]*client.ClientConfig)
		configs.length = len(configs.ccs)
	case []*format.FormatConfig:
		configs.ctype = TYPE_CONFIG_FORMAT
		configs.fcs = configSlice.([]*format.FormatConfig)
		configs.length = len(configs.fcs)
	case *topology.DeployConfig:
		configs.ctype = TYPE_CONFIG_DEPLOY
		configs.dcs = append(configs.dcs, configSlice.(*topology.DeployConfig))
		configs.length = 1
	case *client.ClientConfig:
		configs.ctype = TYPE_CONFIG_CLIENT
		configs.ccs = append(configs.ccs, configSlice.(*client.ClientConfig))
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
	var cc *client.ClientConfig
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

	for i := 0; i < configs.length; i++ {
		// config type
		switch configs.ctype {
		case TYPE_CONFIG_DEPLOY:
			dc = configs.dcs[i]
		case TYPE_CONFIG_CLIENT:
			cc = configs.ccs[i]
		case TYPE_CONFIG_FORMAT:
			fc = configs.fcs[i]
		case TYPE_CONFIG_NULL:
			// do nothing
		}

		// task type
		switch taskType {
		case PULL_IMAGE:
			t, err = comm.NewPullImageTask(curveadm, dc)
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
			t, err = fs.NewMountFSTask(curveadm, cc)
		case UMOUNT_FILESYSTEM:
			option.SilentSubBar = true
			t, err = fs.NewUmountFSTask(curveadm, cc)
		case CHECK_MOUNT_STATUS:
			option.SilentMainBar = true
			option.SilentSubBar = true
			t, err = fs.NewGetMountStatusTask(curveadm, cc)
		case FORMAT_CHUNKFILE_POOL:
			t, err = bs.NewFormatChunkfilePoolTask(curveadm, fc)
		case GET_FORMAT_STATUS:
			t, err = bs.NewGetFormatStatusTask(curveadm, fc)
		default:
			return fmt.Errorf("unknown task type %d", taskType)
		}

		if err != nil {
			return err
		}
		tasks.AddTask(t)
	}

	prettyTasksSubName(tasks.GetTask())
	return tasks.Execute(option)
}

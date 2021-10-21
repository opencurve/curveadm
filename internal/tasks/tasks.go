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
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tuicommon "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	PULL_IMAGE int = iota
	CREATE_CONTAINER
	SYNC_CONFIG
	START_SERVICE
	STOP_SERVICE
	RESTART_SERVICE
	CREATE_TOPOLOGY
	GET_SERVICE_STATUS
	CLEAN_SERVICE
	UNKNOWN
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
		for _, v := range strings.Split(t.SubName(), " ") {
			line = append(line, v)
		}
		lines = append(lines, line)
	}

	output := tuicommon.FixedFormat(lines, 2)
	subnames := strings.Split(output, "\n")
	for i, t := range ts {
		t.SetSubName(subnames[i])
	}
}

func ExecParallelTasks(taskType int, curveadm *cli.CurveAdm, dcs []*configure.DeployConfig) error {
	ts := []*task.Task{}
	options := task.Options{false, false, false}
	for _, dc := range dcs {
		var t *task.Task
		var err error
		switch taskType {
		case PULL_IMAGE:
			t, err = NewPullImageTask(curveadm, dc)
		case CREATE_CONTAINER:
			t, err = NewCreateContainerTask(curveadm, dc)
		case SYNC_CONFIG:
			t, err = NewSyncConfigTask(curveadm, dc)
		case START_SERVICE:
			t, err = NewStartServiceTask(curveadm, dc)
		case STOP_SERVICE:
			t, err = NewStopServiceTask(curveadm, dc)
		case RESTART_SERVICE:
			t, err = NewRestartServiceTask(curveadm, dc)
		case CREATE_TOPOLOGY:
			t, err = NewCreateTopologyTask(curveadm, dc)
		case GET_SERVICE_STATUS:
			options.SilentSubBar = true
			options.SkipError = true
			t, err = NewGetServiceStatusTask(curveadm, dc)
		case CLEAN_SERVICE:
			t, err = NewCleanServiceTask(curveadm, dc)
		default:
			return fmt.Errorf("unknown task type %d", taskType)
		}

		if err != nil {
			return err
		}
		ts = append(ts, t)
	}
	prettyTasksSubName(ts)
	return task.ParallelExecute(10, ts, options)
}

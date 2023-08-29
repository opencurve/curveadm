/*
 *  Copyright (c) 2023 NetEase Inc.
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
 * Created Date: 2023-03-16
 * Author: Cyber-SiKu
 */

package step

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	AFTER_TASK_DIR = "/curve/init.d/"
)

var ALLOC_ID int = 0

type afterRunTask struct {
	ID         int      `json:"ID"`
	Path       string   `json:"Path"`
	Args       []string `json:"Args"`
	Env        []string `json:"Env"`
	Dir        string   `json:"Dir"`
	OutputPath string   `json:"OutputPath"`
	InputPath  string   `json:"InputPath"`
}

func (task afterRunTask) ToString() string {
	b, err := json.Marshal(task)
	if err != nil {
		return ""
	}
	return string(b)
}

func newDaemonTask(path string, args ...string) *afterRunTask {
	task := afterRunTask{
		ID:   ALLOC_ID,
		Path: path,
		Args: args,
	}
	return &task
}

type AddDaemonTask struct {
	ContainerId *string
	Cmd         string
	Args        []string
	TaskName    string
	module.ExecOptions
}

type DelDaemonTask struct {
	ContainerId *string
	Cmd         string
	Args        []string
	TaskName    string
	Tid         string
	MemStorage  *utils.SafeMap
	module.ExecOptions
}

func (s *AddDaemonTask) getAllocId(ctx *context.Context) {
	if ALLOC_ID != 0 {
		ALLOC_ID++
		return
	} else {
		// first add daemon task
		// create dir AFTER_TASK_DIR
		step := ContainerExec{
			ContainerId: s.ContainerId,
			Command:     fmt.Sprintf("mkdir -p %s", AFTER_TASK_DIR),
			ExecOptions: s.ExecOptions,
		}
		err := step.Execute(ctx)
		if err != nil {
			return
		}
	}
	var count string
	// get max id
	// The contents of the file are as follows:
	// {"ID":1,"Path":"tgtd","Args":null,"Env":null,"Dir":"","OutputPath":"","InputPath":""}
	step := ContainerExec{
		ContainerId: s.ContainerId,
		Command:     fmt.Sprintf("grep -r '\"ID\":[0-9]*,' %s | awk -F \":\" '{print $3}' | awk -F \",\" '{print $1}' | sort -n | tail -1", AFTER_TASK_DIR),
		Out:         &count,
		ExecOptions: s.ExecOptions,
	}
	err := step.Execute(ctx)
	if err != nil {
		ALLOC_ID = 1
	}
	id, err := strconv.Atoi(count)
	if err != nil {
		ALLOC_ID = 1
	}
	ALLOC_ID = id + 1
}

func (s *AddDaemonTask) Execute(ctx *context.Context) error {
	s.getAllocId(ctx)
	content := newDaemonTask(s.Cmd, s.Args...).ToString()
	step := InstallFile{
		Content:           &content,
		ContainerId:       s.ContainerId,
		ContainerDestPath: AFTER_TASK_DIR + s.TaskName + ".task",
		ExecOptions:       s.ExecOptions,
	}
	return step.Execute(ctx)
}

type Target struct {
	Host   string
	Tid    string
	Name   string
	Store  string
	Portal string
}

func (s *DelDaemonTask) Execute(ctx *context.Context) error {
	v := s.MemStorage.Get(comm.KEY_ALL_TARGETS)
	target := v.(map[string]*Target)[s.Tid]
	if target == nil {
		return nil
	}
	stores := strings.Split(target.Store, "//")
	if len(stores) < 2 {
		// unable to recognize cbd:pool
		return nil
	}
	path := AFTER_TASK_DIR + "addTarget_" + stores[1] + ".task"
	step := ContainerExec{
		ContainerId: s.ContainerId,
		Command:     "rm -f " + path,
		ExecOptions: s.ExecOptions,
	}
	return step.Execute(ctx)
}

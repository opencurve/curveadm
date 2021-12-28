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

package task

import (
	"errors"

	"github.com/google/uuid"
	ssh "github.com/melbahja/goph"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/pkg/module"
)

var (
	ERR_SKIP_TASK = errors.New("skip task")
)

type (
	Step interface {
		Execute(ctx *context.Context) error
	}

	Task struct {
		tid       string // task id
		ptid      string // parent task id
		name      string
		subname   string
		steps     []Step
		sshConfig *module.SSHConfig
		context   context.Context
	}
)

func NewTask(name, subname string, sshConfig *module.SSHConfig) *Task {
	tid := uuid.NewString()[:12]
	return &Task{
		tid:       tid,
		ptid:      tid,
		name:      name,
		subname:   subname,
		sshConfig: sshConfig,
	}
}

func (t *Task) Tid() string {
	return t.tid
}

func (t *Task) Ptid() string {
	return t.ptid
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) Subname() string {
	return t.subname
}

func (t *Task) SetTid(tid string) {
	t.tid = tid
}

func (t *Task) SetPtid(ptid string) {
	t.ptid = ptid
}

func (t *Task) SetSubname(name string) {
	t.subname = name
}

func (t *Task) AddStep(step Step) {
	t.steps = append(t.steps, step)
}

func (t *Task) Execute() error {
	var sshClient *ssh.Client
	if t.sshConfig != nil {
		client, err := module.NewSshClient(*t.sshConfig)
		if err != nil {
			return err
		}
		sshClient = client
	}

	ctx, err := context.NewContext(sshClient)
	if err != nil {
		return err
	}
	defer ctx.Close()

	for _, step := range t.steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

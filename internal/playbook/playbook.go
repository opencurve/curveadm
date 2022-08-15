/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-07-27
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package playbook

import (
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/tasks"
)

/*
 * playbook
 * ├── tasks1 (e.g.: pull image)
 * ├── tasks2 (e.g.: create container)
 * ├── ...
 * └── tasksn (e.g.: start container)
 *     ├── task1 (e.g.: start container in host1)
 *     ├── task2 (e.g.: start container in host2)
 *     ├── ...
 *     └── taskn (e.g.: start container in host3)
 *         ├── step1 (e.g: start container)
 *         ├── step2 (e.g: check container status)
 *         ├── ...
 *         └── stepn (e.g: start crotab iff status is ok)
 *
 * tasks are made up of many same type tasks which only executed in different hosts or roles
 */
type (
	PlaybookStep struct {
		Name    string
		Type    int
		Configs interface{}
		Options map[string]interface{}
		tasks.ExecOptions
	}

	Playbook struct {
		curveadm  *cli.CurveAdm
		steps     []*PlaybookStep
		postSteps []*PlaybookStep
	}

	ExecOptions = tasks.ExecOptions
)

func NewPlaybook(curveadm *cli.CurveAdm) *Playbook {
	return &Playbook{
		curveadm: curveadm,
		steps:    []*PlaybookStep{},
	}
}

func (p *Playbook) AddStep(s *PlaybookStep) {
	p.steps = append(p.steps, s)
}

func (p *Playbook) AddPostStep(s *PlaybookStep) {
	p.postSteps = append(p.postSteps, s)
}

func (p *Playbook) run(steps []*PlaybookStep) error {
	for i, step := range steps {
		tasks, err := p.createTasks(step)
		if err != nil {
			return err
		}

		err = tasks.Execute(step.ExecOptions)
		if err != nil {
			return err
		}

		isLast := (i == len(steps)-1)
		if !step.ExecOptions.SilentMainBar && !isLast {
			p.curveadm.WriteOutln("")
		}
	}
	return nil
}

func (p *Playbook) Run() error {
	defer func() {
		if len(p.postSteps) == 0 {
			return
		}
		p.curveadm.WriteOutln("")
		p.run(p.postSteps)
	}()

	return p.run(p.steps)
}

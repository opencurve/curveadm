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
	"sync"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/vbauerster/mpb/v7"
	"github.com/vbauerster/mpb/v7/decor"
)

type Step interface {
	Execute(ctx *Context) error
	Rollback(ctx *Context)
}

type Task struct {
	name    string
	subname string
	config  *configure.DeployConfig
	steps   []Step
	context Context
}

type Options struct {
	SilentMainBar bool
	SilentSubBar  bool
	SkipError     bool
}

type Monitor struct {
	err    error
	result map[int]error
	mutex  sync.Mutex
}

func NewTask(name, subname string, config *configure.DeployConfig) *Task {
	return &Task{
		name:    name,
		subname: subname,
		config:  config,
	}
}

func (t *Task) Name() string {
	return t.name
}

func (t *Task) SubName() string {
	return t.subname
}

func (t *Task) SetSubName(name string) {
	t.subname = name
}

func (t *Task) AddStep(step Step) {
	t.steps = append(t.steps, step)
}

func (t *Task) Execute() error {
	ctx, err := NewContext(t.config)
	if err != nil {
		return err
	}

	defer ctx.Clean()
	for _, step := range t.steps {
		err := step.Execute(ctx)
		if err != nil {
			step.Rollback(ctx)
			return err
		}
	}

	return nil
}

func newMonitor() *Monitor {
	return &Monitor{
		err:    nil,
		result: map[int]error{},
		mutex:  sync.Mutex{},
	}
}

func (m *Monitor) error() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.err
}

func (m *Monitor) set(id int, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.result[id] = err
	if m.err == nil {
		m.err = err
	}
}
func (m *Monitor) get(id int) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.result[id]
}

func genShowStatusCb(monitor *Monitor) func(static decor.Statistics) string {
	return func(static decor.Statistics) string {
		if static.Completed {
			if monitor.get(static.ID) == nil {
				return color.GreenString("[OK]")
			} else {
				return color.RedString("[ERROR]")
			}
		}
		return ""
	}
}

// TODO(@Wine93): control step timeout
/*
 * Cluster Name: my-cluster
 *
 * Pull image: [SUCCESS]
 *   + 10.0.0.1: opencurve/curvefs-etcd [DONE]
 *   + 10.0.0.2: opencurve/curvefs-etcd [DONE]
 *   + 10.0.0.3: opencurve/curvefs-etcd [ERROR]
 *   + 10.0.0.1: opencurve/curvefs-mds [DONE]
 *   + 10.0.0.2: opencurve/curvefs-mds [DONE]
 *   + 10.0.0.3: opencurve/curvefs-mds [DONE]
 */
func ParallelExecute(concurrency int, tasks []*Task, options Options) error {
	if len(tasks) == 0 {
		return nil
	}

	var mainbar, bar *mpb.Bar
	wg := sync.WaitGroup{}
	workers := make(chan struct{}, concurrency)
	p := mpb.New(mpb.WithWaitGroup(&wg))
	monitor := newMonitor()
	showStatus := genShowStatusCb(monitor)

	if !options.SilentMainBar {
		mainbar = p.Add(1, nil,
			mpb.PrependDecorators(
				decor.Name(tasks[0].Name()+": "),
				decor.OnComplete(decor.Spinner([]string{}), ""),
				decor.Any(showStatus),
			),
		)
	}

	for _, t := range tasks {
		if monitor.error() != nil && options.SkipError == false {
			break
		}

		wg.Add(1)
		workers <- struct{}{}
		if !options.SilentSubBar {
			bar = p.Add(1, nil,
				mpb.PrependDecorators(
					decor.Name("  + "),
					decor.Name(t.subname+" "),
					decor.OnComplete(decor.Spinner([]string{}), ""),
					decor.Any(showStatus),
				),
			)
		}

		// worker
		go func(t *Task, bar *mpb.Bar) {
			defer func() {
				if bar != nil {
					bar.IncrBy(1)
				}
				<-workers
				wg.Done()
			}()

			err := t.Execute()

			id := 0
			if bar != nil {
				id = bar.ID()
			}

			monitor.set(id, err)
		}(t, bar)
	}

	wg.Wait()

	if mainbar != nil {
		mainbar.IncrBy(1)
		monitor.set(mainbar.ID(), monitor.error())
	}

	p.Wait()

	return monitor.error()
}

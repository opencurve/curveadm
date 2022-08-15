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
 * Created Date: 2022-05-19
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/plugin"
	analyzer2 "github.com/opencurve/curveadm/internal/plugin/analyzer"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type (
	step2StoreVariables struct {
		host       string
		variables  map[string]*analyzer2.Variable
		memStorage *utils.SafeMap
	}
)

func (s *step2StoreVariables) Execute(ctx *context.Context) error {
	// TODO(@Wine93): implement post_task which can output by variable
	if variable, ok := s.variables["output"]; ok {
		s.memStorage.Set(s.host, variable.Value)
	}
	return nil
}

func newStep(s *plugin.PluginStep, variables map[string]*analyzer2.Variable) (task.Step, error) {
	var a analyzer2.Analyzer
	switch s.ModuleName {
	case "shell":
		a = analyzer2.NewShellAnalyzer()
	default:
		return nil, fmt.Errorf("unknown module '%s'", s.ModuleName)
	}

	for key, value := range s.Options {
		err := a.ParseOption(key, value)
		if err != nil {
			return nil, err
		}
	}

	for param, name := range s.BindVariables {
		if _, ok := variables[name]; !ok {
			variables[name] = &analyzer2.Variable{}
		}
		err := a.BindVariable(param, variables[name])
		if err != nil {
			return nil, err
		}
	}

	return a.Build()
}

func NewRunPluginTask(curveadm *cli.CurveAdm, pc *plugin.PluginConfig) (*task.Task, error) {
	subname := fmt.Sprintf("host=%s plugin=%s", pc.GetHost(), pc.GetName())
	t := task.NewTask("Run Plugin", subname, pc.GetSSHConfig())

	variables := map[string]*analyzer2.Variable{}
	for _, s := range pc.GetSteps() {
		step, err := newStep(s, variables)
		if err != nil {
			return nil, err
		}
		t.AddStep(step)
	}

	t.AddStep(&step2StoreVariables{
		host:       pc.GetHost(),
		variables:  variables,
		memStorage: curveadm.MemStorage(),
	})
	return t, nil
}

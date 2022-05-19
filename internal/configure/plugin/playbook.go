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
 * Created Date: 2022-05-18
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/variable"
	"github.com/spf13/viper"
)

type (
	Option struct {
		Key          string `mapstructure:"key"`
		BindVariable string `mapstructure:"bind_variable"`
		Default      string `mapstructure:"default"`
	}

	Step struct {
		Name         string            `mapstructure:"name"`
		Module       string            `mapstructure:"module"`
		Options      map[string]string `mapstructure:"options"`
		BindVariable map[string]string `mapstructure:"bind_variable"`
	}

	Playbook struct {
		Options  []Option `mapstructure:"options"`
		Task     []Step   `mapstructure:"task"`
		PostTask []Step   `mapstructure:"post_task"`
	}
)

func ParsePlaybook(filename string) (*Playbook, error) {
	if !utils.PathExist(filename) {
		return nil, fmt.Errorf("'%s': not exist", filename)
	}

	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	playbook := &Playbook{}
	err = parser.Unmarshal(playbook)
	if err != nil {
		return nil, err
	}
	return playbook, nil
}

/*
 * agrs: []{ "cmd=ls", "sudo=true" }
 * options: [
 *   {
 *     key: cmd
 *     bind_value: command
 *     default:
 *   },
 *   {
 *     key: sudo
 *     bind_value: exec_in_local
 *     default: false
 *   }
 * ]
 */
func (playbook *Playbook) newVariables(args []string) *variable.Variables {
	m := map[string]string{}
	for _, arg := range args {
		items := strings.Split(arg, "=")
		if len(items) >= 2 {
			m[items[0]] = strings.Join(items[1:], "")
		}
	}

	vars := variable.NewVariables()
	for _, option := range playbook.Options {
		if len(option.BindVariable) == 0 {
			continue
		}

		key := option.Key
		value := option.Default
		if v, ok := m[key]; ok { // defined in arguments
			value = v
		}

		vars.Register(variable.Variable{
			Name:     option.BindVariable,
			Value:    value,
			Resolved: true,
		})
	}
	return vars
}

func (playbook *Playbook) Build(name string, args []string, targets []Target) ([]*PluginConfig, error) {
	vars := playbook.newVariables(args)
	if err := vars.Build(); err != nil {
		return nil, err
	}

	pc := &PluginConfig{
		name:      name,
		steps:     []*PluginStep{},
		postSteps: []*PluginStep{},
		vars:      vars,
	}

	for _, step := range playbook.Task {
		pc.steps = append(pc.steps, &PluginStep{
			Name:          step.Name,
			ModuleName:    step.Module,
			Options:       step.Options,
			BindVariables: step.BindVariable,
		})
	}
	for _, step := range playbook.PostTask {
		pc.postSteps = append(pc.postSteps, &PluginStep{
			Name:          step.Name,
			ModuleName:    step.Module,
			Options:       step.Options,
			BindVariables: step.BindVariable,
		})
	}

	err := pc.Build()
	if err != nil {
		return nil, err
	}
	pcs := []*PluginConfig{}
	for _, target := range targets {
		pcs = append(pcs, &PluginConfig{
			name:      pc.name,
			target:    target,
			steps:     pc.steps,
			postSteps: pc.postSteps,
			vars:      pc.vars,
		})
	}
	return pcs, nil
}

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

	"github.com/opencurve/curveadm/pkg/module"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	DEFAULT_SSH_TIMEOUT_SECONDS = 10
)

type PluginStep struct {
	Name          string
	ModuleName    string
	Options       map[string]string
	BindVariables map[string]string
}

type PluginConfig struct {
	name      string
	target    Target
	steps     []*PluginStep
	postSteps []*PluginStep
	vars      *variable.Variables
}

func (pc *PluginConfig) GetName() string         { return pc.name }
func (pc *PluginConfig) GetHost() string         { return pc.target.Host }
func (pc *PluginConfig) GetSteps() []*PluginStep { return pc.steps }
func (pc *PluginConfig) GetSSHConfig() *module.SSHConfig {
	target := pc.target
	return &module.SSHConfig{
		User:           target.User,
		Host:           target.Host,
		Port:           target.Port,
		PrivateKeyPath: target.PrivateKeyFile,
		//Timeout:        DEFAULT_SSH_TIMEOUT_SECONDS,
	}
}

func (pc *PluginConfig) Build() error {
	for _, step := range pc.steps {
		options := step.Options
		for key, value := range options {
			v, err := pc.vars.Rendering(value)
			if err != nil {
				return fmt.Errorf("Rendering variable '%s' failed", value)
			} else {
				value = v
			}
			options[key] = value
		}
	}
	return nil
}

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
	"github.com/melbahja/goph"
	"github.com/opencurve/curveadm/internal/configure"
)

type Register struct {
	m map[string]interface{}
}

type Context struct {
	config   *configure.DeployConfig
	module   *Module
	register *Register
}

func NewRegister() *Register {
	return &Register{map[string]interface{}{}}
}

func (r *Register) Set(key string, value interface{}) {
	r.m[key] = value
}

func (r *Register) Get(key string) interface{} {
	return r.m[key]
}

func NewContext(dc *configure.DeployConfig) (*Context, error) {
	var sshClient *goph.Client
	if dc != nil {
		if client, err := NewSshClient(dc.GetSshConfig()); err != nil {
			return nil, err
		} else {
			sshClient = client
		}
	}

	ctx := &Context{
		config:   dc,
		module:   NewModule(sshClient),
		register: NewRegister(),
	}
	return ctx, nil
}

func (ctx *Context) Clean() {
	ctx.module.Clean()
}

func (ctx *Context) Config() *configure.DeployConfig {
	return ctx.config
}

func (ctx *Context) Module() *Module {
	return ctx.module
}

func (ctx *Context) Register() *Register {
	return ctx.register
}

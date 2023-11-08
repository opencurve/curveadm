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

package playbook

import (
	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
)

const (
	TYPE_CONFIG_HOST int = iota
	TYPE_CONFIG_FORMAT
	TYPE_CONFIG_DEPLOY
	TYPE_CONFIG_CLIENT
	TYPE_CONFIG_PLAYGROUND
	TYPE_CONFIG_MONITOR
	TYPE_CONFIG_WEBSITE
	TYPE_CONFIG_ANY
	TYPE_CONFIG_NULL
)

type SmartConfig struct {
	ctype int
	len   int
	hcs   []*hosts.HostConfig
	fcs   []*configure.FormatConfig
	dcs   []*topology.DeployConfig
	ccs   []*configure.ClientConfig
	pgcs  []*configure.PlaygroundConfig
	mcs   []*configure.MonitorConfig
	wcs   []*configure.WebsiteConfig
	anys  []interface{}
}

func (c *SmartConfig) GetType() int {
	return c.ctype
}

func (c *SmartConfig) Len() int {
	return c.len
}

func (c *SmartConfig) GetHC(index int) *hosts.HostConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_HOST {
		return nil
	}
	return c.hcs[index]
}

func (c *SmartConfig) GetFC(index int) *configure.FormatConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_FORMAT {
		return nil
	}
	return c.fcs[index]
}

func (c *SmartConfig) GetDC(index int) *topology.DeployConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_DEPLOY {
		return nil
	}
	return c.dcs[index]
}

func (c *SmartConfig) GetCC(index int) *configure.ClientConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_CLIENT {
		return nil
	}
	return c.ccs[index]
}

func (c *SmartConfig) GetPGC(index int) *configure.PlaygroundConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_PLAYGROUND {
		return nil
	}
	return c.pgcs[index]
}

func (c *SmartConfig) GetMC(index int) *configure.MonitorConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_MONITOR {
		return nil
	}
	return c.mcs[index]
}

func (c *SmartConfig) GetWC(index int) *configure.WebsiteConfig {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_WEBSITE {
		return nil
	}
	return c.wcs[index]
}

func (c *SmartConfig) GetAny(index int) interface{} {
	if index < 0 || index >= c.len || c.ctype != TYPE_CONFIG_ANY {
		return nil
	}
	return c.anys[index]
}

func NewSmartConfig(configs interface{}) (*SmartConfig, error) {
	c := &SmartConfig{
		ctype: TYPE_CONFIG_NULL,
		len:   0,
		hcs:   []*hosts.HostConfig{},
		fcs:   []*configure.FormatConfig{},
		dcs:   []*topology.DeployConfig{},
		ccs:   []*configure.ClientConfig{},
		pgcs:  []*configure.PlaygroundConfig{},
		mcs:   []*configure.MonitorConfig{},
		wcs:   []*configure.WebsiteConfig{},
		anys:  []interface{}{},
	}
	build.DEBUG(build.DEBUG_SMART_CONFIGS,
		build.Field{Key: "len", Value: c.len},
		build.Field{Key: "type", Value: c.ctype})

	switch configs := configs.(type) {
	// multi-configs
	case []*hosts.HostConfig:
		c.ctype = TYPE_CONFIG_HOST
		c.hcs = configs
		c.len = len(c.hcs)
	case []*configure.FormatConfig:
		c.ctype = TYPE_CONFIG_FORMAT
		c.fcs = configs
		c.len = len(c.fcs)
	case []*topology.DeployConfig:
		c.ctype = TYPE_CONFIG_DEPLOY
		c.dcs = configs
		c.len = len(c.dcs)
	case []*configure.ClientConfig:
		c.ctype = TYPE_CONFIG_CLIENT
		c.ccs = configs
		c.len = len(c.ccs)
	case []*configure.PlaygroundConfig:
		c.ctype = TYPE_CONFIG_PLAYGROUND
		c.pgcs = configs
		c.len = len(c.pgcs)
	case []*configure.MonitorConfig:
		c.ctype = TYPE_CONFIG_MONITOR
		c.mcs = configs
		c.len = len(c.mcs)
	case []*configure.WebsiteConfig:
		c.ctype = TYPE_CONFIG_WEBSITE
		c.wcs = configs
		c.len = len(c.wcs)
	case []interface{}:
		c.ctype = TYPE_CONFIG_ANY
		c.anys = configs
		c.len = len(c.anys)

	// single-config
	case *hosts.HostConfig:
		c.ctype = TYPE_CONFIG_HOST
		c.hcs = append(c.hcs, configs)
		c.len = 1
	case *configure.FormatConfig:
		c.ctype = TYPE_CONFIG_FORMAT
		c.fcs = append(c.fcs, configs)
		c.len = 1
	case *topology.DeployConfig:
		c.ctype = TYPE_CONFIG_DEPLOY
		c.dcs = append(c.dcs, configs)
		c.len = 1
	case *configure.ClientConfig:
		c.ctype = TYPE_CONFIG_CLIENT
		c.ccs = append(c.ccs, configs)
		c.len = 1
	case *configure.PlaygroundConfig:
		c.ctype = TYPE_CONFIG_PLAYGROUND
		c.pgcs = append(c.pgcs, configs)
		c.len = 1
	case *configure.MonitorConfig:
		c.ctype = TYPE_CONFIG_MONITOR
		c.mcs = append(c.mcs, configs)
		c.len = 1
	case *configure.WebsiteConfig:
		c.ctype = TYPE_CONFIG_WEBSITE
		c.wcs = append(c.wcs, configs)
		c.len = 1
	case nil:
		c.ctype = TYPE_CONFIG_NULL
		c.len = 1
	default:
		return nil, errno.ERR_UNSUPPORT_CONFIG_TYPE
	}

	return c, nil
}

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

// __SIGN_BY_WINE93__

package topology

import (
	"fmt"
	"strconv"

	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	KIND_CURVEBS = "curvebs"
	KIND_CURVEFS = "curvefs"

	ROLE_ETCD          = "etcd"
	ROLE_MDS           = "mds"
	ROLE_CHUNKSERVER   = "chunkserver"
	ROLE_SNAPSHOTCLONE = "snapshotclone"
	ROLE_METASERVER    = "metaserver"
)

type (
	DeployConfig struct {
		kind             string // KIND_CURVEFS / KIND_CUVREBS
		id               string // role_host_[name/hostSequence]_replicasSequence
		parentId         string // role_host_[name/hostSequence]_0
		role             string // etcd/mds/metaserevr/chunkserver
		host             string
		hostname         string
		name             string
		instance         int
		hostSequence     int // start with 0
		replicasSequence int // start with 0

		config        map[string]interface{}
		serviceConfig map[string]string
		variables     *variable.Variables
		ctx           *Context
	}

	FilterOption struct {
		Id   string
		Role string
		Host string
	}
)

// etcd_hostname_0_0
func formatId(role, host, name string, replicasSequence int) string {
	return fmt.Sprintf("%s_%s_%s_%d", role, host, name, replicasSequence)
}

func formatName(name string, hostSequence int) string {
	if len(name) == 0 {
		return strconv.Itoa(hostSequence)
	}
	return name
}

func newVariables(m map[string]interface{}) (*variable.Variables, error) {
	vars := variable.NewVariables()
	if m == nil || len(m) == 0 {
		return vars, nil
	}

	for k, v := range m {
		value, ok := utils.All2Str(v)
		if !ok {
			return nil, errno.ERR_UNSUPPORT_VARIABLE_VALUE_TYPE.
				F("%s: %v", k, v)
		} else if len(value) == 0 {
			return nil, errno.ERR_INVALID_VARIABLE_VALUE.
				F("%s: %v", k, v)
		}
		vars.Register(variable.Variable{Name: k, Value: value})
	}
	return vars, nil
}

func NewDeployConfig(ctx *Context, kind, role, host, name string, instance int,
	hostSequence, replicasSequence int, config map[string]interface{}) (*DeployConfig, error) {
	// variable section
	v := config[CONFIG_VARIABLE.key]
	if !utils.IsStringAnyMap(v) && v != nil {
		return nil, errno.ERR_INVALID_VARIABLE_SECTION.
			F("%s: %v", CONFIG_VARIABLE.key, v)
	}

	m := map[string]interface{}{}
	if v != nil {
		m = v.(map[string]interface{})
	}
	vars, err := newVariables(m)
	if err != nil {
		return nil, err
	}
	delete(config, CONFIG_VARIABLE.key)

	// We should convert all value to string for rendering variable,
	// after that we will convert the value to specified type according to
	// the its require type
	for k, v := range config {
		if strv, ok := utils.All2Str(v); ok {
			config[k] = strv
		} else {
			return nil, errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE.
				F("%s: %v", k, v)
		}
	}

	name = formatName(name, hostSequence)
	return &DeployConfig{
		kind:             kind,
		id:               formatId(role, host, name, replicasSequence),
		parentId:         formatId(role, host, name, 0),
		role:             role,
		host:             host,
		name:             name,
		instance:         instance,
		hostSequence:     hostSequence,
		replicasSequence: replicasSequence,
		config:           config,
		serviceConfig:    map[string]string{},
		variables:        vars,
		ctx:              ctx,
	}, nil
}

func (dc *DeployConfig) renderVariables() error {
	vars := dc.GetVariables()
	if err := vars.Build(); err != nil {
		log.Error("Build variables failed",
			log.Field("error", err))
		return errno.ERR_RESOLVE_VARIABLE_FAILED.E(err)
	}

	err := func(values ...*string) error {
		for _, value := range values {
			realValue, err := vars.Rendering(*value)
			if err != nil {
				return err
			}
			*value = realValue
		}
		return nil
	}(&dc.name, &dc.id, &dc.parentId)
	if err != nil {
		return errno.ERR_RENDERING_VARIABLE_FAILED.E(err)
	}

	for k, v := range dc.config {
		realv, err := vars.Rendering(v.(string))
		if err != nil {
			return errno.ERR_RENDERING_VARIABLE_FAILED.E(err)
		}
		dc.config[k] = realv
		build.DEBUG(build.DEBUG_TOPOLOGY,
			build.Field{k, v},
			build.Field{k, realv})
	}
	return nil
}

func (dc *DeployConfig) convert() error {
	// init service config
	for k, v := range dc.config {
		item := itemset.get(k)
		if item == nil || item.exclude == false {
			dc.serviceConfig[k] = v.(string)
		}
	}

	// convret config item to its require type,
	// return error if convert failed
	for _, item := range itemset.getAll() {
		k := item.key
		value := dc.get(item) // return config value or default value
		if value == nil {
			continue
		}
		v, ok := utils.All2Str(value)
		if !ok {
			return errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE.
				F("%s: %v", k, value)
		}

		switch item.require {
		case REQUIRE_ANY:
			// do nothing
		case REQUIRE_INT:
			if intv, ok := utils.Str2Int(v); !ok {
				return errno.ERR_CONFIGURE_VALUE_REQUIRES_INTEGER.
					F("%s: %v", k, value)
			} else {
				dc.config[k] = intv
			}
		case REQUIRE_STRING:
			if len(v) == 0 {
				return errno.ERR_CONFIGURE_VALUE_REQUIRES_NON_EMPTY_STRING.
					F("%s: %v", k, value)
			}
		case REQUIRE_BOOL:
			if boolv, ok := utils.Str2Bool(v); !ok {
				return errno.ERR_CONFIGURE_VALUE_REQUIRES_BOOL.
					F("%s: %v", k, value)
			} else {
				dc.config[k] = boolv
			}
		case REQUIRE_POSITIVE_INTEGER:
			if intv, ok := utils.Str2Int(v); !ok {
				return errno.ERR_CONFIGURE_VALUE_REQUIRES_INTEGER.
					F("%s: %v", k, value)
			} else if intv <= 0 {
				return errno.ERR_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER.
					F("%s: %v", k, value)
			} else {
				dc.config[k] = intv
			}
		}
	}

	return nil
}

func (dc *DeployConfig) ResolveHost() error {
	if dc.ctx == nil {
		dc.hostname = dc.host
		return nil
	}

	vars := dc.GetVariables()
	if err := vars.Build(); err != nil {
		log.Error("Build variables failed",
			log.Field("error", err))
		return errno.ERR_RESOLVE_VARIABLE_FAILED.E(err)
	}

	var err error
	dc.host, err = vars.Rendering(dc.GetHost())
	if err != nil {
		return errno.ERR_RENDERING_VARIABLE_FAILED.E(err)
	}
	dc.hostname = dc.ctx.Lookup(dc.GetHost())
	if len(dc.hostname) == 0 {
		return errno.ERR_HOST_NOT_FOUND.
			F("host: %s", dc.GetHost())
	}
	return nil
}

func (dc *DeployConfig) Build() error {
	err := dc.renderVariables()
	if err != nil {
		return err
	}
	return dc.convert()
}

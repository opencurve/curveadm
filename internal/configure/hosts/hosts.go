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
 * Created Date: 2022-07-20
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package hosts

import (
	"bytes"
	"strings"

	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/configure/os"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	KEY_LABELS = "labels"
	KEY_ENVS   = "envs"

	PERMISSIONS_600 = 384 // -rw------- (256 + 128 = 384)
)

type (
	Hosts struct {
		Global map[string]interface{}   `mapstructure:"global"`
		Host   []map[string]interface{} `mapstructure:"hosts"`
	}

	HostConfig struct {
		sequence int
		config   map[string]interface{}
		labels   []string
		envs     []string
	}
)

func NewIfNil(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return map[string]interface{}{}
	}
	return config
}

func Merge(parent, child map[string]interface{}) {
	for k, v := range parent {
		if child[k] == nil {
			child[k] = v
		}
	}
}

func (hc *HostConfig) convertLables() error {
	value := hc.config[KEY_LABELS]
	slice, ok := (value).([]interface{})
	if !ok {
		return errno.ERR_CONFIGURE_VALUE_REQUIRES_STRING_SLICE.
			F("hosts[%d].%s = %v", hc.sequence, KEY_LABELS, value)
	}

	for _, value := range slice {
		if v, ok := utils.All2Str(value); !ok {
			return errno.ERR_CONFIGURE_VALUE_REQUIRES_STRING_SLICE.
				F("hosts[%d].%s = %v", hc.sequence, KEY_LABELS, value)
		} else {
			hc.labels = append(hc.labels, v)
		}
	}

	return nil
}

func (hc *HostConfig) convertEnvs() error {
	value := hc.config[KEY_ENVS]
	slice, ok := (value).([]interface{})
	if !ok {
		return errno.ERR_CONFIGURE_VALUE_REQUIRES_STRING_SLICE.
			F("hosts[%d].%s = %v", hc.sequence, KEY_ENVS, value)
	}

	for _, value := range slice {
		if v, ok := utils.All2Str(value); !ok {
			return errno.ERR_CONFIGURE_VALUE_REQUIRES_STRING_SLICE.
				F("hosts[%d].%s = %v", hc.sequence, KEY_ENVS, value)
		} else {
			hc.envs = append(hc.envs, v)
		}
	}

	return nil
}

func (hc *HostConfig) Build() error {
	for key, value := range hc.config {
		if key == KEY_LABELS { // convert labels
			if err := hc.convertLables(); err != nil {
				return err
			}
			hc.config[key] = nil // delete labels section
			continue
		} else if key == KEY_ENVS { // convert envs
			if err := hc.convertEnvs(); err != nil {
				return err
			}
			hc.config[key] = nil // delete labels section
			continue
		}

		if itemset.Get(key) == nil {
			return errno.ERR_UNSUPPORT_HOSTS_CONFIGURE_ITEM.
				F("hosts[%d].%s = %v", hc.sequence, key, value)
		}

		v, err := itemset.Build(key, value)
		if err != nil {
			return err
		} else {
			hc.config[key] = v
		}
	}

	privateKeyFile := hc.GetPrivateKeyFile()
	if len(hc.GetHost()) == 0 {
		return errno.ERR_HOST_FIELD_MISSING.
			F("hosts[%d].host = nil", hc.sequence)
	} else if len(hc.GetHostname()) == 0 {
		return errno.ERR_HOSTNAME_FIELD_MISSING.
			F("hosts[%d].hostname = nil", hc.sequence)
	} else if !utils.IsValidAddress(hc.GetHostname()) {
		return errno.ERR_HOSTNAME_REQUIRES_VALID_IP_ADDRESS.
			F("hosts[%d].hostname = %s", hc.sequence, hc.GetHostname())
	} else if hc.GetSSHPort() > os.GetMaxPortNum() {
		return errno.ERR_HOSTS_SSH_PORT_EXCEED_MAX_PORT_NUMBER.
			F("hosts[%d].ssh_port = %d", hc.sequence, hc.GetSSHPort())
	} else if !strings.HasPrefix(privateKeyFile, "/") {
		return errno.ERR_PRIVATE_KEY_FILE_REQUIRE_ABSOLUTE_PATH.
			F("hosts[%d].private_key_file = %s", hc.sequence, privateKeyFile)
	}

	if hc.GetForwardAgent() == false {
		if !utils.PathExist(privateKeyFile) {
			return errno.ERR_PRIVATE_KEY_FILE_NOT_EXIST.
				F("%s: no such file", privateKeyFile)
		} else if utils.GetFilePermissions(privateKeyFile) != PERMISSIONS_600 {
			return errno.ERR_PRIVATE_KEY_FILE_REQUIRE_600_PERMISSIONS.
				F("%s: mode (%d)", privateKeyFile, utils.GetFilePermissions(privateKeyFile))
		}
	}
	return nil
}

func NewHostConfig(sequence int, config map[string]interface{}) *HostConfig {
	return &HostConfig{
		sequence: sequence,
		config:   config,
		labels:   []string{},
	}
}

func ParseHosts(data string) ([]*HostConfig, error) {
	if len(data) == 0 {
		return nil, errno.ERR_EMPTY_HOSTS
	}

	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigType("yaml")
	err := parser.ReadConfig(bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, errno.ERR_PARSE_HOSTS_FAILED.E(err)
	}

	hosts := &Hosts{}
	if err := parser.Unmarshal(hosts); err != nil {
		return nil, errno.ERR_PARSE_HOSTS_FAILED.E(err)
	}

	hcs := []*HostConfig{}
	exist := map[string]bool{}
	for i, host := range hosts.Host {
		host = NewIfNil(host)
		Merge(hosts.Global, host)
		hc := NewHostConfig(i, host)
		err = hc.Build()
		if err != nil {
			return nil, err
		}

		if _, ok := exist[hc.GetHost()]; ok {
			return nil, errno.ERR_DUPLICATE_HOST.
				F("duplicate host: %s", hc.GetHost())
		}
		hcs = append(hcs, hc)
		exist[hc.GetHost()] = true
	}
	build.DEBUG(build.DEBUG_HOSTS, hosts)
	return hcs, nil
}

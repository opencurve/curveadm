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
 * Created Date: 2022-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
	"github.com/opencurve/curveadm/pkg/variable"
	"github.com/spf13/viper"
)

const (
	KEY_USER             = "user"
	KEY_HOST             = "host"
	KEY_SSH_PORT         = "ssh_port"
	KEY_PRIVATE_KEY_FILE = "private_key_file"
	KEY_CONTAINER_IMAGE  = "container_image"
	KEY_LOG_DIR          = "log_dir"
	KEY_LISTEN_MDS_ADDR  = "mds.listen.addr"

	DEFAULT_NEBD_DATA_DIR       = "/var/run/nebd/data"
	DEFAULT_SSH_TIMEOUT_SECONDS = 10
)

var (
	excludeClientConfig = map[string]bool{
		KEY_USER:             true,
		KEY_HOST:             true,
		KEY_SSH_PORT:         true,
		KEY_PRIVATE_KEY_FILE: true,
		KEY_CONTAINER_IMAGE:  true,
		KEY_LOG_DIR:          true,
	}
)

type (
	ClientConfig struct {
		config        map[string]string
		serviceConfig map[string]string
		variables     *variable.Variables
	}
)

func ParseClientConfig(filename string) (*ClientConfig, error) {
	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	if err := parser.ReadInConfig(); err != nil {
		return nil, err
	}

	m := map[string]interface{}{}
	if err := parser.Unmarshal(&m); err != nil {
		return nil, err
	}

	// add variables (only one)
	vars := variable.NewVariables()
	vars.Register(variable.Variable{Name: "prefix", Value: "/curvebs/nebd"})
	err := vars.Build()
	if err != nil {
		return nil, err
	}

	//
	config := map[string]string{}
	serviceConfig := map[string]string{}
	for k, v := range m {
		value, ok := utils.All2Str(v)
		if !ok {
			return nil, fmt.Errorf("unsupport value type fot config key '%s'", k)
		}
		config[k] = value
		if !excludeClientConfig[k] {
			serviceConfig[k] = value
		}
	}

	return &ClientConfig{
		config:        config,
		serviceConfig: serviceConfig,
		variables:     vars,
	}, nil
}

func (c *ClientConfig) GetConfig(key string) string {
	key = strings.ToLower(key)
	return c.config[key]
}

func (c *ClientConfig) GetServiceConfig() map[string]string {
	return c.serviceConfig
}

func (cc *ClientConfig) GetUser() string { return cc.GetConfig(KEY_USER) }
func (cc *ClientConfig) GetHost() string { return cc.GetConfig(KEY_HOST) }
func (cc *ClientConfig) GetSSHPort() int {
	port, _ := strconv.Atoi(cc.GetConfig(KEY_SSH_PORT))
	return port
}
func (cc *ClientConfig) GetPrivateKeyFile() string { return cc.GetConfig(KEY_PRIVATE_KEY_FILE) }
func (cc *ClientConfig) GetContainerImage() string { return cc.GetConfig(KEY_CONTAINER_IMAGE) }
func (cc *ClientConfig) GetDataDir() string        { return DEFAULT_NEBD_DATA_DIR }
func (cc *ClientConfig) GetLogDir() string         { return cc.GetConfig(KEY_LOG_DIR) }
func (cc *ClientConfig) GetClusterMDSAddr() string { return cc.GetConfig(KEY_LISTEN_MDS_ADDR) }

func (cc *ClientConfig) GetVariables() *variable.Variables {
	return cc.variables
}

func (cc *ClientConfig) GetSSHConfig() *module.SSHConfig {
	return &module.SSHConfig{
		User:           cc.GetUser(),
		Host:           cc.GetHost(),
		Port:           (uint)(cc.GetSSHPort()),
		PrivateKeyPath: cc.GetPrivateKeyFile(),
		Timeout:        DEFAULT_SSH_TIMEOUT_SECONDS,
	}
}

func (cc *ClientConfig) GetProjectLayout() topology.Layout {
	return topology.GetCurveBSProjectLayout()
}

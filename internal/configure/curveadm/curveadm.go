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

package curveadm

import (
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

/*
 * [defaults]
 * log_level = error
 * sudo_alias = ""
 *
 * [ssh_connection]
 * retries = 3
 * timeout = 10
 */
const (
	KEY_LOG_LEVEL   = "log_level"
	KEY_SUDO_ALIAS  = "sudo_method"
	KEY_SSH_RETRIES = "retries"
	KEY_SSH_TIMEOUT = "timeout"
)

type (
	CurveAdmConfig struct {
		LogLevel   string
		SudoAlias  string
		SSHRetries int
		SSHTimeout int
	}

	global struct {
		Defaults       map[string]interface{} `mapstructure:"defaults"`
		SSHConnections map[string]interface{} `mapstructure:"ssh_connections"`
	}
)

var (
	defaultCurveAdmConfig = &CurveAdmConfig{
		LogLevel:   "error",
		SudoAlias:  "sudo",
		SSHRetries: 3,
		SSHTimeout: 10,
	}
)

func ParseCurveAdmConfig(filename string) (*CurveAdmConfig, error) {
	admConfig := defaultCurveAdmConfig
	if !utils.PathExist(filename) {
		return admConfig, nil
	}

	// parse curveadm config
	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("ini")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	global := &global{}
	err = parser.Unmarshal(global)
	if err != nil {
		return nil, err
	}

	// default section
	defaults := global.Defaults
	if defaults != nil {
		if v, ok := defaults[KEY_LOG_LEVEL]; ok {
			admConfig.LogLevel = v.(string)
		}
		if v, ok := defaults[KEY_SUDO_ALIAS]; ok {
			admConfig.SudoAlias = v.(string)
		}
	}

	// ssh_connection
	sshConnection := global.SSHConnections
	if sshConnection != nil {
		if v, ok := sshConnection[KEY_SSH_RETRIES]; ok {
			admConfig.SSHRetries = v.(int)
		}
		if v, ok := sshConnection[KEY_SSH_TIMEOUT]; ok {
			admConfig.SSHTimeout = v.(int)
		}
	}

	return admConfig, nil
}

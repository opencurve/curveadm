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

package configure

import (
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	LOG_LEVEL = "log_level"
)

type CurveAdmConfig struct {
	LogLevel   string
	SshTimeout int
	SshRetries int
}

type (
	curveadm struct {
		Defaults       map[string]interface{} `mapstructure:"defaults"`
		SshConnections map[string]interface{} `mapstructure:"ssh_connections"`
	}
)

func ParseCurveAdmConfig(filename string) (*CurveAdmConfig, error) {
	curveAdmConfig := &CurveAdmConfig{"error", 10, 3}
	if !utils.PathExist(filename) {
		return curveAdmConfig, nil
	}

	// parse curveadm config
	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("ini")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	curveadm := &curveadm{}
	err = parser.Unmarshal(curveadm)
	if err != nil {
		return nil, err
	}

	defaults := curveadm.Defaults
	if defaults != nil {
		if v, ok := defaults[LOG_LEVEL]; ok {
			curveAdmConfig.LogLevel = v.(string)
		}
	}

	return curveAdmConfig, nil
}

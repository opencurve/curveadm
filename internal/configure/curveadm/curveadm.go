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

package curveadm

import (
	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

/*
 * [defaults]
 * log_level = error
 * sudo_alias = "sudo"
 * timeout = 180
 *
 * [ssh_connections]
 * retries = 3
 * timeout = 10
 */
const (
	KEY_LOG_LEVEL   = "log_level"
	KEY_SUDO_ALIAS  = "sudo_alias"
	KEY_TIMEOUT     = "timeout"
	KEY_SSH_RETRIES = "retries"
	KEY_SSH_TIMEOUT = "timeout"

	WITHOUT_SUDO = " "
)

type (
	CurveAdmConfig struct {
		LogLevel   string
		SudoAlias  string
		Timeout    int
		SSHRetries int
		SSHTimeout int
	}

	CurveAdm struct {
		Defaults       map[string]interface{} `mapstructure:"defaults"`
		SSHConnections map[string]interface{} `mapstructure:"ssh_connections"`
	}
)

var (
	GlobalCurveAdmConfig *CurveAdmConfig

	defaultCurveAdmConfig = &CurveAdmConfig{
		LogLevel:   "error",
		SudoAlias:  "sudo",
		Timeout:    180,
		SSHRetries: 3,
		SSHTimeout: 10,
	}

	SUPPORT_LOG_LEVEL = map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
)

func ReplaceGlobals(cfg *CurveAdmConfig) {
	GlobalCurveAdmConfig = cfg
}

// TODO(P2): using ItemSet to check value type
func requirePositiveInt(k string, v interface{}) (int, error) {
	num, ok := utils.Str2Int(v.(string))
	if !ok {
		return 0, errno.ERR_CONFIGURE_VALUE_REQUIRES_INTEGER.
			F("%s: %v", k, v)
	} else if num <= 0 {
		return 0, errno.ERR_CONFIGURE_VALUE_REQUIRES_POSITIVE_INTEGER.
			F("%s: %v", k, v)
	}
	return num, nil
}

func parseDefaultsSection(cfg *CurveAdmConfig, defaults map[string]interface{}) error {
	if defaults == nil {
		return nil
	}

	for k, v := range defaults {
		switch k {
		// log_level
		case KEY_LOG_LEVEL:
			if !SUPPORT_LOG_LEVEL[v.(string)] {
				return errno.ERR_UNSUPPORT_CURVEADM_LOG_LEVEL.
					F("%s: %s", KEY_LOG_LEVEL, v.(string))
			}
			cfg.LogLevel = v.(string)

		// sudo_alias
		case KEY_SUDO_ALIAS:
			cfg.SudoAlias = v.(string)

		// timeout
		case KEY_TIMEOUT:
			num, err := requirePositiveInt(KEY_TIMEOUT, v)
			if err != nil {
				return err
			}
			cfg.Timeout = num

		default:
			return errno.ERR_UNSUPPORT_CURVEADM_CONFIGURE_ITEM.
				F("%s: %s", k, v)
		}
	}

	return nil
}

func parseConnectionSection(cfg *CurveAdmConfig, connection map[string]interface{}) error {
	if connection == nil {
		return nil
	}

	for k, v := range connection {
		switch k {
		// ssh_retries
		case KEY_SSH_RETRIES:
			num, err := requirePositiveInt(KEY_SSH_RETRIES, v)
			if err != nil {
				return err
			}
			cfg.SSHRetries = num

		// ssh_timeout
		case KEY_SSH_TIMEOUT:
			num, err := requirePositiveInt(KEY_SSH_TIMEOUT, v)
			if err != nil {
				return err
			}
			cfg.SSHTimeout = num

		default:
			return errno.ERR_UNSUPPORT_CURVEADM_CONFIGURE_ITEM.
				F("%s: %s", k, v)
		}
	}

	return nil
}

func ParseCurveAdmConfig(filename string) (*CurveAdmConfig, error) {
	cfg := defaultCurveAdmConfig
	if !utils.PathExist(filename) {
		build.DEBUG(build.DEBUG_CURVEADM_CONFIGURE, cfg)
		return cfg, nil
	}

	// parse curveadm config
	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("ini")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, errno.ERR_PARSE_CURVRADM_CONFIGURE_FAILED.E(err)
	}

	global := &CurveAdm{}
	err = parser.Unmarshal(global)
	if err != nil {
		return nil, errno.ERR_PARSE_CURVRADM_CONFIGURE_FAILED.E(err)
	}

	err = parseDefaultsSection(cfg, global.Defaults)
	if err != nil {
		return nil, err
	}
	err = parseConnectionSection(cfg, global.SSHConnections)
	if err != nil {
		return nil, err
	}

	build.DEBUG(build.DEBUG_CURVEADM_CONFIGURE, cfg)
	return cfg, nil
}

func (cfg *CurveAdmConfig) GetLogLevel() string { return cfg.LogLevel }
func (cfg *CurveAdmConfig) GetTimeout() int     { return cfg.Timeout }
func (cfg *CurveAdmConfig) GetSSHRetries() int  { return cfg.SSHRetries }
func (cfg *CurveAdmConfig) GetSSHTimeout() int  { return cfg.SSHTimeout }
func (cfg *CurveAdmConfig) GetSudoAlias() string {
	if len(cfg.SudoAlias) == 0 {
		return WITHOUT_SUDO
	}
	return cfg.SudoAlias
}

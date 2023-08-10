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
	"fmt"
	"os"
	"regexp"

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
 *
 * [database]
 * url = "sqlite:///home/curve/.curveadm/data/curveadm.db"
 */
const (
	KEY_LOG_LEVEL    = "log_level"
	KEY_SUDO_ALIAS   = "sudo_alias"
	KEY_ENGINE       = "engine"
	KEY_TIMEOUT      = "timeout"
	KEY_AUTO_UPGRADE = "auto_upgrade"
	KEY_SSH_RETRIES  = "retries"
	KEY_SSH_TIMEOUT  = "timeout"
	KEY_DB_URL       = "url"

	// rqlite://127.0.0.1:4000
	// sqlite:///home/curve/.curveadm/data/curveadm.db
	REGEX_DB_URL = "^(sqlite|rqlite)://(.+)$"
	DB_SQLITE    = "sqlite"
	DB_RQLITE    = "rqlite"

	WITHOUT_SUDO = " "
)

type (
	CurveAdmConfig struct {
		LogLevel    string
		SudoAlias   string
		Engine      string
		Timeout     int
		AutoUpgrade bool
		SSHRetries  int
		SSHTimeout  int
		DBUrl       string
	}

	CurveAdm struct {
		Defaults       map[string]interface{} `mapstructure:"defaults"`
		SSHConnections map[string]interface{} `mapstructure:"ssh_connections"`
		DataBase       map[string]interface{} `mapstructure:"database"`
	}
)

var (
	GlobalCurveAdmConfig *CurveAdmConfig

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

func newDefault() *CurveAdmConfig {
	home, _ := os.UserHomeDir()
	cfg := &CurveAdmConfig{
		LogLevel:    "error",
		SudoAlias:   "sudo",
		Engine:      "docker",
		Timeout:     180,
		AutoUpgrade: true,
		SSHRetries:  3,
		SSHTimeout:  10,
		DBUrl:       fmt.Sprintf("sqlite://%s/.curveadm/data/curveadm.db", home),
	}
	return cfg
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

func requirePositiveBool(k string, v interface{}) (bool, error) {
	yes, ok := utils.Str2Bool(v.(string))
	if !ok {
		return false, errno.ERR_CONFIGURE_VALUE_REQUIRES_BOOL.
			F("%s: %v", k, v)
	}
	return yes, nil
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

		// container engine
		case KEY_ENGINE:
			cfg.Engine = v.(string)

		// timeout
		case KEY_TIMEOUT:
			num, err := requirePositiveInt(KEY_TIMEOUT, v)
			if err != nil {
				return err
			}
			cfg.Timeout = num

		// auto upgrade
		case KEY_AUTO_UPGRADE:
			yes, err := requirePositiveBool(KEY_AUTO_UPGRADE, v)
			if err != nil {
				return err
			}
			cfg.AutoUpgrade = yes

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

func parseDatabaseSection(cfg *CurveAdmConfig, database map[string]interface{}) error {
	if database == nil {
		return nil
	}

	for k, v := range database {
		switch k {
		// database url
		case KEY_DB_URL:
			dbUrl := v.(string)
			pattern := regexp.MustCompile(REGEX_DB_URL)
			mu := pattern.FindStringSubmatch(dbUrl)
			if len(mu) == 0 {
				return errno.ERR_UNSUPPORT_CURVEADM_DATABASE_URL.F("url: %s", dbUrl)
			}
			cfg.DBUrl = dbUrl

		default:
			return errno.ERR_UNSUPPORT_CURVEADM_CONFIGURE_ITEM.
				F("%s: %s", k, v)
		}
	}

	return nil
}

type sectionParser struct {
	parser  func(*CurveAdmConfig, map[string]interface{}) error
	section map[string]interface{}
}

func ParseCurveAdmConfig(filename string) (*CurveAdmConfig, error) {
	cfg := newDefault()
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

	items := []sectionParser{
		{parseDefaultsSection, global.Defaults},
		{parseConnectionSection, global.SSHConnections},
		{parseDatabaseSection, global.DataBase},
	}
	for _, item := range items {
		err := item.parser(cfg, item.section)
		if err != nil {
			return nil, err
		}
	}

	build.DEBUG(build.DEBUG_CURVEADM_CONFIGURE, cfg)
	return cfg, nil
}

func (cfg *CurveAdmConfig) GetLogLevel() string  { return cfg.LogLevel }
func (cfg *CurveAdmConfig) GetTimeout() int      { return cfg.Timeout }
func (cfg *CurveAdmConfig) GetAutoUpgrade() bool { return cfg.AutoUpgrade }
func (cfg *CurveAdmConfig) GetSSHRetries() int   { return cfg.SSHRetries }
func (cfg *CurveAdmConfig) GetSSHTimeout() int   { return cfg.SSHTimeout }
func (cfg *CurveAdmConfig) GetEngine() string    { return cfg.Engine }
func (cfg *CurveAdmConfig) GetSudoAlias() string {
	if len(cfg.SudoAlias) == 0 {
		return WITHOUT_SUDO
	}
	return cfg.SudoAlias
}

func (cfg *CurveAdmConfig) GetDBUrl() string {
	return cfg.DBUrl
}

func (cfg *CurveAdmConfig) GetDBPath() string {
	pattern := regexp.MustCompile(REGEX_DB_URL)
	mu := pattern.FindStringSubmatch(cfg.DBUrl)
	if len(mu) == 0 || mu[1] != DB_SQLITE {
		return ""
	}
	return mu[2]
}

/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-05-06
* Author: wanghai (SeanHai)
 */

package configure

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	ROLE_WEBSITE = "website"
)

type website struct {
	Kind   string                 `mapstructure:"kind"`
	Host   string                 `mapstructure:"host"`
	Config map[string]interface{} `mapstructure:"config"`
}

type WebsiteConfig struct {
	kind   string
	id     string // role_host
	role   string
	host   string
	config map[string]interface{}
}

func (w *WebsiteConfig) getString(key string) string {
	v := w.config[strings.ToLower(key)]
	if v == nil {
		return ""
	}
	return v.(string)
}

func (m *WebsiteConfig) getInt(key string) int {
	v := m.config[strings.ToLower(key)]
	if v == nil {
		return -1
	}
	return v.(int)
}

func (w *WebsiteConfig) GetImage() string {
	return w.getString(KEY_CONTAINER_IMAGE)
}

func (w *WebsiteConfig) GetHost() string {
	return w.host
}

func (w *WebsiteConfig) GetRole() string {
	return w.role
}

func (w *WebsiteConfig) GetKind() string {
	return w.kind
}

func (w *WebsiteConfig) GetId() string {
	return w.id
}

func (w *WebsiteConfig) GetDataDir() string {
	return w.getString(KEY_DATA_DIR)
}

func (w *WebsiteConfig) GetLogDir() string {
	return w.getString(KEY_LOG_DIR)
}

func (w *WebsiteConfig) GetListenPort() int {
	return w.getInt(KEY_LISTEN_PORT)
}

func (w *WebsiteConfig) GetServiceConfig() map[string]interface{} {
	return w.config
}

func ParseWebsiteConfig(filename string) ([]*WebsiteConfig, error) {
	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigType("yaml")
	if !utils.PathExist(filename) {
		return nil, errno.ERR_WEBSITE_CONF_FILE_NOT_FOUND.F("filepath: %s", filename)
	}
	parser.SetConfigFile(filename)
	if err := parser.ReadInConfig(); err != nil {
		return nil, errno.ERR_PARSE_WEBSITE_CONFIGURE_FAILED.E(err)
	}
	config := website{}
	if err := parser.Unmarshal(&config); err != nil {
		return nil, errno.ERR_PARSE_WEBSITE_CONFIGURE_FAILED.E(err)
	}

	ret := []*WebsiteConfig{}
	k := config.Kind
	h := config.Host
	ret = append(ret, &WebsiteConfig{
		kind:   k,
		host:   h,
		id:     fmt.Sprintf("%s_%s", ROLE_WEBSITE, h),
		role:   ROLE_WEBSITE,
		config: config.Config,
	})
	return ret, nil
}

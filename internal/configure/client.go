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
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

var (
	KEY_MDS_ADDR = "mdsOpt.rpcRetryOpt.addrs"

	defaultValue = map[string]string{
		KEY_CONTAINER_IMAGE: "opencurvedocker/curvefs:latest",
	}

	excludeClientConfig = map[string]bool{
		KEY_LOG_DIR:         true,
		KEY_CONTAINER_IMAGE: true,
	}
)

type (
	ClientConfig struct {
		config        map[string]string
		serviceConfig map[string]string
	}
)

func all2str(v interface{}) string {
	if utils.IsString(v) {
		return v.(string)
	} else if utils.IsInt(v) {
		return strconv.Itoa(v.(int))
	} else if utils.IsBool(v) {
		return strconv.FormatBool(v.(bool))
	}
	return ""
}

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

	c := &ClientConfig{
		config:        map[string]string{},
		serviceConfig: map[string]string{},
	}
	for k, v := range m {
		value := all2str(v)
		c.config[k] = value
		if !excludeClientConfig[k] {
			c.serviceConfig[k] = value
		}
	}

	if err := c.check(filename); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *ClientConfig) check(filename string) error {
	if c.GetConfig(KEY_MDS_ADDR) == "" {
		return fmt.Errorf("please specify '%s' in %s", KEY_MDS_ADDR, filename)
	}
	return nil
}

func (c *ClientConfig) GetConfig(key string) string {
	key = strings.ToLower(key)
	if v, ok := c.config[key]; ok {
		return v
	}
	return defaultValue[key]
}

func (c *ClientConfig) GetServiceConfig() map[string]string {
	return c.serviceConfig
}

// wrapper interface
func (c *ClientConfig) GetContainerImage() string {
	return c.GetConfig(KEY_CONTAINER_IMAGE)
}

func (c *ClientConfig) GetLogDir() string {
	return c.GetConfig(KEY_LOG_DIR)
}

func (c *ClientConfig) GetCurveFSConfPath() string {
	return fmt.Sprintf("%s/conf/client.conf", DEFAULT_CURVEFS_DIR)
}

func (c *ClientConfig) GetProjectConfPath() string {
	return fmt.Sprintf("%s/client/conf/client.conf", DEFAULT_CURVEFS_DIR)
}

func (c *ClientConfig) GetProjectMountPath() string {
	return fmt.Sprintf("%s/client/mnt", DEFAULT_CURVEFS_DIR)
}

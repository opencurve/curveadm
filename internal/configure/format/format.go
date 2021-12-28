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
 * Created Date: 2021-12-27
 * Author: Jingli Chen (Wine93)
 */

package format

import (
	"fmt"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
	"github.com/spf13/viper"
)

type (
	FormatConfig struct {
		Host         string
		SSHPort      int
		Device       string
		MountPoint   string
		UsagePercent int
	}

	Format struct {
		User           string   `mapstructure:"user"`
		SSHPort        string   `mapstructure:"ssh_port"`
		PrivateKeyFile string   `mapstructure:"private_key_file"`
		Host           []string `mapstructure:"host"`
		Disk           []string `mapstructure:"disk"`
	}
)

func (fc *FormatConfig) GetHost() string {

}

func (fc *FormatConfig) GetDevice() string {

}

func (fc *FormatConfig) GetUsagePercent() int {
}

func (fc *FormatConfig) GetMountPoint() string {

}

func (fc *FormatConfig) GetContainerIamge() string {

}

func (fc *FormatConfig) GetSSHConfig() *module.SSHConfig {
	return ssh
}

func ParseFormat(filename string) ([]FormatConfig, error) {
	if !utils.PathExist(filename) {
		return nil, fmt.Errorf("'%s': not exist", filename)
	}

	// parse disk list
	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	disks := &Disks{}
	err = parser.Unmarshal(disks)
	if err != nil {
		return nil, err
	} else if disks.Disks == nil {
		return nil, fmt.Errorf("")
	}

	return curveAdmConfig, nil
}

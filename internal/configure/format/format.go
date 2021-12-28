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
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	LATEST_CURVBS_VERSION  = "v1.2"
	FORMAT_CONTAINER_IMAGE = "opencurvedocker/curvebs:%s"
)

type (
	Format struct {
		User           string   `mapstructure:"user"`
		SSHPort        int      `mapstructure:"ssh_port"`
		PrivateKeyFile string   `mapstructure:"private_key_file"`
		Version        string   `mapstructure:"version"`
		Hosts          []string `mapstructure:"host"`
		Disks          []string `mapstructure:"disk"`
	}
)

func ParseFormat(filename string) ([]*FormatConfig, error) {
	if !utils.PathExist(filename) {
		return nil, fmt.Errorf("'%s': not exist", filename)
	}

	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, err
	}

	format := &Format{}
	err = parser.Unmarshal(format)
	if err != nil {
		return nil, err
	}

	// version
	version := LATEST_CURVBS_VERSION
	if len(format.Version) > 0 {
		version = format.Version
	}
	containerImage := fmt.Sprintf(FORMAT_CONTAINER_IMAGE, version)

	fcs := []*FormatConfig{}
	for _, host := range format.Hosts {
		for _, disk := range format.Disks {
			// /dev/sda:/data/chunkserver0:90
			items := strings.Split(disk, ":")
			if len(items) != 3 {
				return nil, fmt.Errorf("'%s': invalid disk format", disk)
			}
			usagePercent, ok := utils.Str2Int(items[2])
			if !ok {
				return nil, fmt.Errorf("'%s': invalid disk format", disk)
			}

			fc := &FormatConfig{
				Host:           host,
				User:           format.User,
				SSHPort:        format.SSHPort,
				PrivateKeyFile: format.PrivateKeyFile,
				ContainerIamge: containerImage,
				Device:         items[0],
				MountPoint:     items[1],
				UsagePercent:   usagePercent,
			}
			fcs = append(fcs, fc)
		}
	}
	return fcs, nil
}

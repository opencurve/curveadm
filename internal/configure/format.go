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

package configure

import (
	"strings"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	DEFAULT_CONTAINER_IMAGE = "opencurvedocker/curvebs:v1.2"
)

/*
 * host:
 *   - machine1
 *   - machine2
 *   - machine3
 * disk:
 *   - /dev/sda:/data/chunkserver0:10  # device:mount_path:format_percent
 *   - /dev/sdb:/data/chunkserver1:10
 *   - /dev/sdc:/data/chunkserver2:10
 */
type (
	FormatConfig struct {
		ContainerIamge string
		Host           string
		Device         string
		MountPoint     string
		FormtPercent   int
		UseDiskUri     bool
	}

	Format struct {
		ContainerImage string   `mapstructure:"container_image"`
		Hosts          []string `mapstructure:"host"`
		Disks          []string `mapstructure:"disk"`
	}
)

func NewFormatConfig(containerImage, host, disk string) (*FormatConfig, error) {
	items := strings.Split(disk, ":")
	if len(items) != 3 {
		return nil, errno.ERR_INVALID_DISK_FORMAT.S(disk)
	}

	device, mountPoint, percent := items[0], items[1], items[2]
	if !strings.HasPrefix(device, "/") {
		return nil, errno.ERR_INVALID_DEVICE.
			F("device: %s", device)
	} else if !strings.HasPrefix(mountPoint, "/") {
		return nil, errno.ERR_MOUNT_POINT_REQUIRE_ABSOLUTE_PATH.
			F("mountPoint: %s", mountPoint)
	}

	formatPercent, ok := utils.Str2Int(percent)
	if !ok {
		return nil, errno.ERR_FORMAT_PERCENT_REQUIRES_INTERGET.
			F("percent: %s", percent)
	} else if formatPercent <= 0 || formatPercent > 100 {
		return nil, errno.ERR_FORMAT_PERCENT_MUST_BE_BETWEEN_1_AND_100.
			F("percent: %s", percent)
	}

	return &FormatConfig{
		ContainerIamge: containerImage,
		Host:           host,
		Device:         device,
		MountPoint:     mountPoint,
		FormtPercent:   formatPercent,
	}, nil
}

func ParseFormat(filename string) ([]*FormatConfig, error) {
	if !utils.PathExist(filename) {
		return nil, errno.ERR_FORMAT_CONFIGURE_FILE_NOT_EXIST.
			F("filepath: %s", filename)
	}

	parser := viper.New()
	parser.SetConfigFile(filename)
	parser.SetConfigType("yaml")
	err := parser.ReadInConfig()
	if err != nil {
		return nil, errno.ERR_PARSE_FORMAT_CONFIGURE_FAILED.E(err)
	}

	format := &Format{}
	err = parser.Unmarshal(format)
	if err != nil {
		return nil, errno.ERR_PARSE_FORMAT_CONFIGURE_FAILED.E(err)
	}

	containerImage := DEFAULT_CONTAINER_IMAGE
	if len(format.ContainerImage) > 0 {
		containerImage = format.ContainerImage
	}

	fcs := []*FormatConfig{}
	for _, host := range format.Hosts {
		for _, disk := range format.Disks {
			fc, err := NewFormatConfig(containerImage, host, disk)
			if err != nil {
				return nil, err
			}
			fcs = append(fcs, fc)
		}
	}

	return fcs, nil
}

func (fc *FormatConfig) GetContainerImage() string { return fc.ContainerIamge }
func (fc *FormatConfig) GetHost() string           { return fc.Host }
func (fc *FormatConfig) GetDevice() string         { return fc.Device }
func (fc *FormatConfig) GetMountPoint() string     { return fc.MountPoint }
func (fc *FormatConfig) GetFormatPercent() int     { return fc.FormtPercent }

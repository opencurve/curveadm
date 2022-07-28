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
 * Created Date: 2021-12-29
 * Author: Jingli Chen (Wine93)
 */

package format

import (
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	DEFAULT_SSH_TIMEOUT = 10 // seconds
)

type (
	FormatConfig struct {
		Host           string
		User           string
		SSHPort        int
        ChunkSize      int
		PrivateKeyFile string
		ContainerIamge string
		Device         string
		MountPoint     string
		UsagePercent   int
	}
)

func (fc *FormatConfig) GetHost() string           { return fc.Host }
func (fc *FormatConfig) GetContainerIamge() string { return fc.ContainerIamge }
func (fc *FormatConfig) GetDevice() string         { return fc.Device }
func (fc *FormatConfig) GetMountPoint() string     { return fc.MountPoint }
func (fc *FormatConfig) GetUsagePercent() int      { return fc.UsagePercent }
func (fc *FormatConfig) GetChunkSize() int         { return fc.ChunkSize }

func (fc *FormatConfig) GetSSHConfig() *module.SSHConfig {
	return &module.SSHConfig{
		User:           fc.User,
		Host:           fc.Host,
		Port:           uint(fc.SSHPort),
		PrivateKeyPath: fc.PrivateKeyFile,
		Timeout:        DEFAULT_SSH_TIMEOUT,
	}
}

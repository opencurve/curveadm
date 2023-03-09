/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2023-03-09
 * Author: Lijin Xiong (lijin.xiong@zstack.io)
 */

package disks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	disksDataGlobal = `
global:
  format_percent: 95
  container_image: docker-registry:5000/curvebs:v1.2
  service_mount_device: true
  host:
    - curve-1
    - curve-2
    - curve-3
`
	disksDataNormal = disksDataGlobal + `
disk:
  - device: /dev/sda1
    mount: /data/chunkserver0
    format_percent: 90
  - device: /dev/sdb1
    mount: /data/chunkserver1
    service_mount_device: false
  - device: /dev/sdc1
    mount: /data/chunkserver2
  - device: /dev/sdd1
    mount: /data/chunkserver3
    host:
    - curve-1
    - curve-2
  - device: /dev/sde1
    mount: /data/chunkserver4
    exclude:
    - curve-3
`
	disksDataInvalidFormatPercent = disksDataGlobal + `
disk:
  - device: /dev/sda1
    mount: /data/chunkserver0
    format_percent: test
`

	disksDataInvalidServiceMount = disksDataGlobal + `
disk:

  - device: /dev/sdd1
    mount: /data/chunkserver3
	service_mount_device: yes
`
	disksDataInvalidHostValue = disksDataGlobal + `
disk:
  - device: /dev/sdd1
    mount: /data/chunkserver3
    host: curve-1
`
	disksDataInvalidHostValue2 = disksDataGlobal + `
disk:
  - device: /dev/sdd1
    mount: /data/chunkserver3
    host: 1
`

	disksDataInvalidHostValue3 = disksDataGlobal + `
disk:
  - device: /dev/sdd1
    mount: /data/chunkserver3
    host:
	- 2023
`

	disksDataInvalidHostExclude = disksDataGlobal + `
disk:
  - device: /dev/sde1
    mount: /data/chunkserver4
    exclude: curve-3
`
	disksDataInvalidHostExclude2 = disksDataGlobal + `
disk:
  - device: /dev/sde1
    mount: /data/chunkserver4
    exclude:
`
	disksDataInvalidHostExclude3 = disksDataGlobal + `
disk:
  - device: /dev/sde1
    mount: /data/chunkserver4
    exclude:
    - no-such-host
`
)

func TestDiskConfigNormal(t *testing.T) {
	assert := assert.New(t)

	disks, err := ParseDisks(disksDataNormal)
	assert.Nil(err)
	assert.Equal(len(disks), 5)
	assert.Equal(disks[0].GetFormatPercent(), 90)
	assert.Equal(disks[1].GetFormatPercent(), 95)
	assert.True(disks[0].GetServiceMount())
	assert.False(disks[1].GetServiceMount())
	assert.Equal(len(disks[3].GetHost()), 2)
	assert.Equal(len(disks[4].GetHostExclude()), 1)
	assert.IsType([]string{}, disks[3].GetHost())
	assert.IsType([]string{}, disks[4].GetHostExclude())
}

func TestDiskConfigInvalidFormatPercent(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidFormatPercent)
	assert.Error(err)

}

// tservice_mount_device is a boolean,
// It accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False
func TestDiskConfigInvalidServiceMount(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidServiceMount)
	assert.Error(err)
}

func TestDiskConfigInvalidHostValue(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostValue)
	assert.Error(err)
}

func TestDiskConfigInvalidHostValue2(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostValue2)
	assert.Error(err)
}

func TestDiskConfigInvalidHostValue3(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostValue3)
	assert.Error(err)
}

func TestDiskConfigInvalidHostExclude(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostExclude)
	assert.Error(err)
}

func TestDiskConfigInvalidHostExclude2(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostExclude2)
	assert.Error(err)
}

func TestDiskConfigInvalidHostExclude3(t *testing.T) {
	assert := assert.New(t)
	_, err := ParseDisks(disksDataInvalidHostExclude3)
	assert.Error(err)
}

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
 * Project: CurveAdm
 * Created Date: 2023-02-24
 * Author: Lijin Xiong (lijin.xiong@zstack.io)
 */

package disks

import (
	"bytes"
	"strings"

	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

const (
	DISK_URI_SEP           = "//"
	DISK_URI_PROTO_FS_UUID = "fs:uuid"
)

type (
	Disks struct {
		Global map[string]interface{}   `mapstructure:"global"`
		Disk   []map[string]interface{} `mapstructure:"disk"`
	}

	DiskConfig struct {
		sequence int
		config   map[string]interface{}
	}
)

func merge(parent, child map[string]interface{}) {
	for k, v := range parent {
		if child[k] == nil {
			child[k] = v
		}
	}

}

func newIfNil(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return map[string]interface{}{}
	}
	return config
}

func mergeFinalHost(dc *DiskConfig) error {
	hostExclude := dc.GetHostExclude()
	if len(hostExclude) > 0 {
		diskHost := []string{}
		hosts := dc.GetHost()
		hostMap := utils.Slice2Map(hosts)
		hostExcludeMap := utils.Slice2Map(hostExclude)
		for _, h := range hosts {
			if _, ok := hostExcludeMap[h]; !ok {
				diskHost = append(diskHost, h)
			}
		}
		// check if the host to be excluded is the member of global host list
		for _, h := range hostExclude {
			if _, ok := hostMap[h]; !ok {
				return errno.ERR_HOST_NOT_FOUND.
					F("no such host: %s", h)
			}
		}
		dc.config[common.DISK_FILTER_HOST] = diskHost
	}
	return nil
}

func checkDupHost(dc *DiskConfig) error {
	existHost := map[string]bool{}
	for _, h := range dc.GetHost() {
		if _, ok := existHost[h]; ok {
			return errno.ERR_DUPLICATE_HOST.
				F("duplicated host: %s", h)
		}
		existHost[h] = true
	}
	return nil
}

func checkDiskRootConfig(disks *Disks) error {
	if disks.Global == nil {
		return errno.ERR_GLOBAL_FIELD_MISSING.
			F("disks yaml has not 'global' field")
	}
	if disks.Disk == nil {
		return errno.ERR_DISK_FIELD_MISSING.
			F("disks yaml has not 'disk' field")
	}
	return nil
}

func GenDiskURI(proto, uri string) string {
	return strings.Join([]string{proto, uri}, DISK_URI_SEP)
}

func returnInvalidDiskUriError(disk storage.Disk) error {
	return errno.ERR_INVALID_DISK_URI.
		F("The URI[%s] of disk[%s:%s] is invalid", disk.Host, disk.Device, disk.URI)
}

func GetDiskId(disk storage.Disk) (diskId, diskUriProto string, err error) {
	// valide disk uri:
	// 1. fs:uuid//8035a617-72ec-4c06-8719-8aca79234ef9
	// 2. (not implemented) maybe "nvme:pci//00:00.1"
	diskUriComponants := strings.Split(disk.URI, DISK_URI_SEP)
	if len(diskUriComponants) < 2 {
		return "", diskUriProto, returnInvalidDiskUriError(disk)
	}

	diskUriProto = diskUriComponants[0]
	switch diskUriProto {
	case DISK_URI_PROTO_FS_UUID:
		return diskUriComponants[1], diskUriProto, nil
	default:
		return "", diskUriProto, returnInvalidDiskUriError(disk)
	}
}

func (dc *DiskConfig) Build() error {
	for key, value := range dc.config {
		if itemset.Get(key) == nil {
			return errno.ERR_UNSUPPORT_DISKS_CONFIGURE_ITEM.
				F("disks[%d].%s = %v", dc.sequence, key, value)
		}
		v, err := itemset.Build(key, value)
		if err != nil {
			return err
		} else {
			dc.config[key] = v
		}
	}

	if err := mergeFinalHost(dc); err != nil {
		return err
	}

	if len(dc.GetHost()) == 0 {
		return errno.ERR_HOST_FIELD_MISSING.
			F("disks[%d].host = []", dc.sequence)
	} else if dc.GetDevice() == "" {
		return errno.ERR_DEVICE_FIELD_MISSING.
			F("disks[%d].device = nil", dc.sequence)
	} else if dc.GetMountPoint() == "" {
		return errno.ERR_MOUNT_POINT_FIELD_MISSING.
			F("disks[%d].mount = nil", dc.sequence)
	} else if dc.GetFormatPercent() == 0 {
		return errno.ERR_FORMAT_PERCENT_FIELD_MISSING.
			F("disks[%d].format_percent = nil", dc.sequence)
	} else if dc.GetFormatPercent() > 100 {
		return errno.ERR_DISK_FORMAT_PERCENT_EXCEED_100.
			F("disks[%d].format_percent = %d", dc.sequence, dc.GetFormatPercent())
	}

	if err := checkDupHost(dc); err != nil {
		return err
	}

	return nil
}

func NewDiskConfig(sequence int, config map[string]interface{}) *DiskConfig {
	return &DiskConfig{
		sequence: sequence,
		config:   config,
	}
}

func ParseDisks(data string) ([]*DiskConfig, error) {
	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigType("yaml")
	err := parser.ReadConfig(bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, errno.ERR_PARSE_DISKS_FAILED.E(err)
	}

	disks := &Disks{}
	if err := parser.Unmarshal(disks); err != nil {
		return nil, errno.ERR_PARSE_DISKS_FAILED.E(err)
	}

	var dcs []*DiskConfig
	exist := map[string]bool{}
	if err := checkDiskRootConfig(disks); err != nil {
		return dcs, err
	}
	for i, disk := range disks.Disk {
		disk = newIfNil(disk)
		merge(disks.Global, disk)
		dc := NewDiskConfig(i, disk)
		err = dc.Build()
		if err != nil {
			return nil, err
		}

		if _, ok := exist[dc.GetDevice()]; ok {
			return nil, errno.ERR_DUPLICATE_DISK.
				F("duplicate disk: %s", dc.GetDevice())
		}
		if _, ok := exist[dc.GetMountPoint()]; ok {
			return nil, errno.ERR_DUPLICATE_DISK.
				F("duplicate disk mount point: %s", dc.GetMountPoint())
		}

		dcs = append(dcs, dc)
		exist[dc.GetDevice()] = true
		exist[dc.GetMountPoint()] = true
	}

	build.DEBUG(build.DEBUG_DISKS, disks)
	return dcs, nil
}

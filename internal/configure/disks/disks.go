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

package disks

import (
	"bytes"

	"github.com/opencurve/curveadm/internal/build"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/spf13/viper"
)

type (
	Disks struct {
		Global map[string]interface{}   `mapstructure:"global"`
		Disk   []map[string]interface{} `mapstructure:"disks"`
	}

	DiskConfig struct {
		sequence int
		config   map[string]interface{}
	}
)

func (dc *DiskConfig) Build() error {
	for key, value := range dc.config {
		if itemset.Get(key) == nil {
			return errno.ERR_UNSUPPORT_DISKS_CONFIGURE_ITEM.
				F("disks[%d].%s = %v", dc.sequence, key, value)
		}
		if _, ok := value.([]interface{}); !ok {
			v, err := itemset.Build(key, value)
			if err != nil {
				return err
			} else {
				dc.config[key] = v
			}
		}
	}

	if len(dc.GetDevice()) == 0 {
		return errno.ERR_DEVICE_FIELD_MISSING.
			F("disks[%d].device = nil", dc.sequence)
	} else if len(dc.GetMountPoint()) == 0 {
		return errno.ERR_MOUNT_POINT_FIELD_MISSING.
			F("disks[%d].mount = nil", dc.sequence)
	} else if dc.GetFormatPercent() > 100 {
		return errno.ERR_DISK_FORMAT_PERCENT_EXCEED_100.
			F("disks[%d].format_percent = %d", dc.sequence, dc.GetFormatPercent())
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
	if len(data) == 0 {
		return nil, errno.ERR_EMPTY_DISKS
	}

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

	dcs := []*DiskConfig{}
	exist := map[string]bool{}
	for i, disk := range disks.Disk {
		disk = hosts.NewIfNil(disk)

		hosts.Merge(disks.Global, disk)
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
		hostsExclude := dc.GetHostsExclude()
		hostsOnly := dc.GetHostsOnly()
		if len(hostsExclude) > 0 && len(hostsOnly) > 0 {
			return nil, errno.ERR_ONLY_EXCLUDE_HOSTS.
				F("conflict fields hosts_exclude: %s and hosts_only: %s", hostsExclude, hostsOnly)
		}
		dcs = append(dcs, dc)
		exist[dc.GetDevice()] = true
		exist[dc.GetMountPoint()] = true
	}
	build.DEBUG(build.DEBUG_DISKS, disks)
	return dcs, nil
}

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
		ContainerImage string   `mapstructure:"container_image"`
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
	containerImage := format.ContainerImage

	fcs := []*FormatConfig{}
	for _, host := range format.Hosts {
		for _, disk := range format.Disks {
			items := strings.Split(disk, ":")
			if len(items) != 3 {
				return nil, fmt.Errorf("'%s': invalid disk format", disk)
			}
			usagePercent, ok := utils.Str2Int(items[2])
			if !ok {
				return nil, fmt.Errorf("'%s': invalid disk format", disk)
			}

			// /dev/sd[h-k]:/data/chunkserver[5-8]:70
			// /dev/sd[a-d,^bc]:/data/chunkserver[1-2]:90
			// only support one disk letter in range,
			// unsupport [aa-ad], use /dev/sda[a-d] instead.
			// only support digist in mount point range, eg. /data/chunkserver[10-20]
			if strings.Contains(disk, "[") {
				diskList := items[0]                        // disks in range: /dev/sd[a-d,^bc] or /dev/sd[h-k]
				start := strings.Index(diskList, "[") + 1   // index of start disk letter('a' or 'h') in range
				end := strings.Index(diskList, "]")         // index of end disk letter('d' or 'k') in range
				diskSeq := diskList[start:end]              // get a-d,^bc or h-k
				disks := strings.Split(diskSeq, ",")        // split include disks and exclude disks
				includeDisks := strings.TrimSpace(disks[0]) // get a-d or h-k
				excludeDisks := ""
				var excludes []rune
				// get exclude disks: bc
				if len(disks) == 2 {
					excludeDisks = strings.TrimSpace(disks[1])[1:] // remove ^ and space
					excludes = []rune(excludeDisks)
				}
				disks = strings.Split(includeDisks, "-")         // split include disk range: [a d] or [h k]
				s := []rune(disks[0])[0]                         // get a or h
				e := []rune(disks[1])[0]                         // get d or k
				diskCount := int(e) - int(s) + 1 - len(excludes) // calc disk count in device range

				// get /data/chunkserver[5-8] or /data/chunkserver[1-2]
				mountPoints := items[1]
				startm := strings.Index(mountPoints, "[") + 1                          // get index of 5 or 1
				endm := strings.Index(mountPoints, "]")                                // get index of 8 or 2
				mps := strings.Split(strings.TrimSpace(mountPoints[startm:endm]), "-") // get [5 8] or [1 2]
				sm, _ := utils.Str2Int(strings.TrimSpace(mps[0]))                      // transfer string to int
				em, _ := utils.Str2Int(strings.TrimSpace(mps[1]))                      // transfer string to int

				// disk count is not equal to mountpoint count
				// eg. /dev/sd[a-b]:/data/chunkserver[1-3]:70
				if diskCount != (em - sm + 1) {
					return nil, fmt.Errorf("'%s': invalid disk format, disk count does not match mount point count", disk)
				}
				// add all disks in range(pop exclude disks)
				for ; s <= e; s++ {
					find := false
					for _, ex := range excludes {
						// find the disk letter in exclude disks
						if s == ex {
							find = true
							break
						}
					}
					// disk not in excludes, add it
					if !find {
						device := diskList[:start-1] + string(s) + diskList[end+1:]          // /dev/sd + a + suffix
						mp := mountPoints[:startm-1] + utils.Atoa(sm) + mountPoints[endm+1:] // /data/chunkserver + 1 + suffix
						sm++                                                                 // next mount point number
						fc := &FormatConfig{
							Host:           host,
							User:           format.User,
							SSHPort:        format.SSHPort,
							PrivateKeyFile: format.PrivateKeyFile,
							ContainerIamge: containerImage,
							Device:         device,
							MountPoint:     mp,
							UsagePercent:   usagePercent,
						}
						fcs = append(fcs, fc)
					}
				}
			} else {
				// /dev/sda:/data/chunkserver0:90
				// single disk, add it
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
	}
	return fcs, nil
}

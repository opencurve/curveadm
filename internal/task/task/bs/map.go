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
 * Created Date: 2022-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	TOOLSV2_CONFIG_DELIMITER = ":"
	TOOLSV2_CONFIG_SRC_PATH  = "/curvebs/conf/curve.yaml"
	TOOLSV2_CONFIG_DEST_PATH = "/etc/curve/curve.yaml"
)

type (
	MapOptions struct {
		Host        string
		User        string
		Volume      string
		Create      bool
		Size        int
		NoExclusive bool
	}
)

func checkMapStatus(success *bool, out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *success {
			return nil
		}
		return errno.ERR_MAP_VOLUME_FAILED.S(*out)
	}
}

func getMapOptions(options MapOptions) string {
	mapOptions := []string{}
	if options.NoExclusive {
		mapOptions = append(mapOptions, "--no-exclusive")
	}
	return strings.Join(mapOptions, " ")
}

func checkMapDiskStatus(out *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if *out == "SUCCESS" {
			return nil
		} else if *out == "EXIST" {
			return task.ERR_SKIP_TASK
		}
		return errno.ERR_SIZE_EXIST
	}
}

func NewMapTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_MAP_OPTIONS).(MapOptions)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("hostname=%s volume=%s:%s", hc.GetHostname(), options.User, options.Volume)
	t := task.NewTask("Map Volume", subname, hc.GetSSHConfig())

	// add step
	var out string
	var success bool
	containerName := volume2ContainerName(options.User, options.Volume)
	containerId := containerName
	script := scripts.MAP
	scriptPath := "/curvebs/nebd/sbin/map.sh"
	mapOptions := getMapOptions(options)
	command := fmt.Sprintf("/bin/bash %s %s %s %s", scriptPath, options.User, options.Volume, mapOptions)

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.Status}}'",
		Quiet:       true,
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkVolumeStatus(&out),
	})
	t.AddStep(&step.ModProbe{
		Name:        comm.KERNERL_MODULE_NBD,
		Args:        []string{"nbds_max=64"},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.AddDaemonTask{ // install modprobe.task
		ContainerId: &containerId,
		Cmd:         "modprobe",
		Args:        []string{comm.KERNERL_MODULE_NBD, "nbds_max=64"},
		TaskName:    "modProbe",
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.SyncFile{ // sync nebd-client config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  "/curvebs/conf/nebd-client.conf",
		ContainerDestId:   &containerId,
		ContainerDestPath: "/etc/nebd/nebd-client.conf",
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.InstallFile{ // install map.sh
		Content:           &script,
		ContainerId:       &containerId,
		ContainerDestPath: scriptPath,
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.TrySyncFile{ // sync toolsv2 config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  TOOLSV2_CONFIG_SRC_PATH,
		ContainerDestId:   &containerId,
		ContainerDestPath: TOOLSV2_CONFIG_DEST_PATH,
		KVFieldSplit:      TOOLSV2_CONFIG_DELIMITER,
		Mutate:            newToolsV2Mutate(cc, TOOLSV2_CONFIG_DELIMITER),
		ExecOptions:       curveadm.ExecOptions(),
	})
	t.AddStep(&step.AddDaemonTask{ // install map.task
		ContainerId: &containerId,
		Cmd:         "/bin/bash",
		Args:        []string{scriptPath, options.User, options.Volume, mapOptions},
		TaskName:    "map",
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     command,
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkMapDiskStatus(&out),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkMapStatus(&success, &out),
	})

	return t, nil
}

func newToolsV2Mutate(cc *configure.ClientConfig, delimiter string) step.Mutate {
	clientConfig := cc.GetServiceConfig()
	tools2client := map[string]string{
		"mdsAddr": "mdsOpt.rpcRetryOpt.addrs",
	}
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}
		replaceKey := strings.TrimSpace(key)
		if tools2client[strings.TrimSpace(key)] != "" {
			replaceKey = tools2client[strings.TrimSpace(key)]
		}
		v, ok := clientConfig[strings.ToLower(replaceKey)]
		if ok {
			value = v
		}
		out = fmt.Sprintf("%s%s %s", key, delimiter, value)
		return
	}
}

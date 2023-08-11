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
 * Created Date: 2022-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/task/task/checker"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	CLIENT_CONFIG_DELIMITER = "="
)

type (
	step2InsertClient struct {
		curveadm    *cli.CurveAdm
		options     MapOptions
		config      *configure.ClientConfig
		containerId *string
	}

	AuxInfo struct {
		User   string `json:"user"`
		Volume string `json:"volume,"`
		Config string `json:"config,omitempty"` // TODO(P1)
	}
)

func formatImage(user, volume string) string {
	return fmt.Sprintf("cbd:pool/%s_%s_", volume, user)
}

func volume2ContainerName(user, volume string) string {
	return fmt.Sprintf("curvebs-volume-%s", utils.MD5Sum(formatImage(user, volume)))
}

func checkVolumeExist(volume string, containerId *string) step.LambdaType {
	return func(ctx *context.Context) error {
		if len(*containerId) > 0 {
			return errno.ERR_VOLUME_ALREADY_MAPPED.
				F("volume: %s", volume)
		}
		return nil
	}
}

func newMutate(cc *configure.ClientConfig, delimiter string) step.Mutate {
	serviceConfig := cc.GetServiceConfig()
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}

		// replace config
		v, ok := serviceConfig[strings.ToLower(key)]
		if ok {
			value = v
		}

		// replace variable
		value, err = cc.GetVariables().Rendering(value)
		if err != nil {
			return
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func getVolumes(cc *configure.ClientConfig) []step.Volume {
	volumes := []step.Volume{
		{HostPath: "/dev", ContainerPath: "/dev"},
		{HostPath: "/lib/modules", ContainerPath: "/lib/modules"},
	}
	// see also: https://github.com/opencurve/curve/tree/master/nebd/etc/nebd
	if len(cc.GetLogDir()) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      cc.GetLogDir(),
			ContainerPath: "/curvebs/nebd/logs",
		})
	}
	return volumes
}

func (s *step2InsertClient) Execute(ctx *context.Context) error {
	config := s.config
	curveadm := s.curveadm
	options := s.options
	volumeId := curveadm.GetVolumeId(options.Host, options.User, options.Volume)

	auxInfo := &AuxInfo{
		User:   options.User,
		Volume: options.Volume,
	}
	bytes, err := json.Marshal(auxInfo)
	if err != nil {
		return errno.ERR_ENCODE_VOLUME_INFO_TO_JSON_FAILED.E(err)
	}

	err = curveadm.Storage().InsertClient(volumeId, config.GetKind(),
		options.Host, *s.containerId, string(bytes))
	if err != nil {
		return errno.ERR_INSERT_CLIENT_FAILED.E(err)
	}
	return nil
}

func NewStartNEBDServiceTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_MAP_OPTIONS).(MapOptions)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("hostname=%s image=%s", hc.GetHostname(), cc.GetContainerImage())
	t := task.NewTask("Start NEBD Service", subname, hc.GetSSHConfig())

	// add step
	var containerId, out string
	var success bool
	volume := fmt.Sprintf("%s:%s", options.User, options.Volume)
	containerName := volume2ContainerName(options.User, options.Volume)
	hostname := containerName
	host2addr := fmt.Sprintf("%s:%s", hostname, hc.GetHostname())

	t.AddStep(&step.EngineInfo{
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checker.CheckEngineInfo(options.Host, curveadm.ExecOptions().ExecWithEngine, &success, &out),
	})
	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}}'",
		Filter:      fmt.Sprintf("name=%s", containerName),
		Out:         &containerId,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Lambda{
		Lambda: checkVolumeExist(volume, &containerId),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:       []string{cc.GetLogDir(), cc.GetDataDir()},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.PullImage{
		Image:       cc.GetContainerImage(),
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.CreateContainer{
		Image:       cc.GetContainerImage(),
		AddHost:     []string{host2addr},
		Envs:        []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Hostname:    hostname,
		Command:     fmt.Sprintf("--role nebd"),
		Name:        containerName,
		Pid:         "host",
		Privileged:  true,
		Volumes:     getVolumes(cc),
		Out:         &containerId,
		Restart:     comm.POLICY_UNLESS_STOPPED,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2InsertClient{
		curveadm:    curveadm,
		options:     options,
		config:      cc,
		containerId: &containerId,
	})
	for _, filename := range []string{"client.conf", "nebd-server.conf"} {
		t.AddStep(&step.SyncFile{
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  "/curvebs/conf/" + filename,
			ContainerDestId:   &containerId,
			ContainerDestPath: "/curvebs/nebd/conf/" + filename,
			KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
			Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
			ExecOptions:       curveadm.ExecOptions(),
		})
	}
	t.AddStep(&step.StartContainer{
		ContainerId: &containerId,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})

	return t, nil
}

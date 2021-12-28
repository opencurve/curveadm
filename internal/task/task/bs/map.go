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
 * Created Date: 2021-01-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	KEY_MAP_OPTION = "MAP_OPTION"
)

type MapOption struct {
	Volume string
	User   string
	Create bool
	Size   int
}

type (
	step2CheckNBDDaemon struct {
		containerId *string
		user        string
		volume      string
	}
)

func (s *step2CheckNBDDaemon) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return fmt.Errorf("image mapped, please run 'curveadm unmap %s:%s' first", s.user, s.volume)
	}
	return nil
}

func image2ContainerName(user, volume string) string {
	return utils.MD5Sum(fmt.Sprintf("curvebs-iamge-%s_%s", user, volume))
}

func NewMapTask(curvradm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	option := curvradm.MemStorage().Get(KEY_MAP_OPTION).(MapOption)
	user, volume := option.User, option.Volume
	subname := fmt.Sprintf("hostname=%s volume=%s", cc.GetHost(), volume)
	t := task.NewTask("Map Image", subname, cc.GetSSHConfig())

	// add step
	var containerId string
	containerName := image2ContainerName(user, volume)
	mapScriptPath := "/curvebs/nebd/sbin/map.sh"
	toolsConf := fmt.Sprintf("mdsAddr=%s", cc.GetClusterMDSAddr())
	mapScript := scripts.MAP
	t.AddStep(&step.ListContainers{
		ShowAll:      true,
		Format:       "'{{.ID}}'",
		Quiet:        true,
		Filter:       fmt.Sprintf("name=%s", containerName),
		Out:          &containerId,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step2CheckNBDDaemon{
		containerId: &containerId,
		user:        user,
		volume:      volume,
	})
	t.AddStep(&step.PullImage{
		Image:        cc.GetContainerImage(),
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step.CreateContainer{
		Image:        cc.GetContainerImage(),
		Command:      fmt.Sprintf("bash %s %s %s %v %d", mapScriptPath, user, volume, option.Create, option.Size),
		Entrypoint:   "/exec.sh",
		Envs:         []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Name:         containerName,
		Pid:          "host",
		Privileged:   true,
		Volumes:      getVolumes(cc),
		Out:          &containerId,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})
	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerId,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecWithSudo:      true,
		ExecInLocal:       false,
	})
	t.AddStep(&step.InstallFile{ // install map.sh
		Content:           &mapScript,
		ContainerId:       &containerId,
		ContainerDestPath: mapScriptPath,
		ExecWithSudo:      true,
		ExecInLocal:       false,
	})
	t.AddStep(&step.SyncFile{ // sync nebd-client config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  "/curvebs/conf/nebd-client.conf",
		ContainerDestId:   &containerId,
		ContainerDestPath: "/etc/nebd/nebd-client.conf",
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecWithSudo:      true,
		ExecInLocal:       false,
	})
	t.AddStep(&step.StartContainer{
		ContainerId:  &containerId,
		ExecWithSudo: true,
		ExecInLocal:  false,
	})

	return t, nil
}

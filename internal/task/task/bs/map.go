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
	"strconv"
	"strings"
	"time"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	KEY_MAP_OPTION = "MAP_OPTION"
	FORMAT_IMAGE   = "cbd:pool/%s_%s_"

	FORMAT_TOOLS_CONF = `mdsAddr=%s
rootUserName=root
rootUserPassword=root_password
`
)

type MapOption struct {
	User   string
	Volume string
	Create bool
	Size   int
}

type (
	step2CheckNEBDClient struct {
		containerId *string
		user        string
		volume      string
	}

	step2CreateNBDDevice struct {
		execWithSudo  bool
		execInLocal   bool
		execSudoAlias string
	}
)

type (
	waitMapDone struct {
		ContainerId    *string
		ExecWithSudo   bool
		ExecInLocal    bool
		ExecTimeoutSec int
		ExecSudoAlias  string
	}
)

var (
	CONTAINER_ABNORMAL_STATUS = map[string]bool{
		"dead":   true,
		"exited": true,
	}
)

func (s *step2CheckNEBDClient) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return fmt.Errorf("volume mapped, please run 'curveadm unmap %s:%s' first",
			s.user, s.volume)
	}
	return nil
}

func (s *step2CreateNBDDevice) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().ModProbe("nbd", "nbds_max=64")
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo:  s.execWithSudo,
		ExecInLocal:   s.execInLocal,
		ExecSudoAlias: s.execSudoAlias,
	})
	return err
}

func (s *waitMapDone) Execute(ctx *context.Context) error {
	var output string
	getStatusStep := &step.InspectContainer{
		ContainerId:   *s.ContainerId,
		Format:        "'{{.State.ExitCode}} {{.State.Status}}'",
		Out:           &output,
		ExecWithSudo:  s.ExecWithSudo,
		ExecInLocal:   s.ExecInLocal,
		ExecSudoAlias: s.ExecSudoAlias,
	}
	// check status
	var err error
	var status string
	var exitCode int
	for i := 0; i < s.ExecTimeoutSec; i++ {
		time.Sleep(time.Second)
		err = getStatusStep.Execute(ctx)
		stringSlice := strings.Split(strings.TrimRight(output, "\n"), " ")
		exitCode, _ = strconv.Atoi(stringSlice[0])
		status = stringSlice[1]
		if err != nil || CONTAINER_ABNORMAL_STATUS[status] {
			break
		}
	}
	if err == nil && exitCode != 0 {
		return fmt.Errorf("please use `docker logs %s` for details", *s.ContainerId)
	}
	return err
}

func formatImage(user, volume string) string {
	return fmt.Sprintf(FORMAT_IMAGE, volume, user)
}

func volume2ContainerName(user, volume string) string {
	return fmt.Sprintf("curvebs-volume-%s", utils.MD5Sum(formatImage(user, volume)))
}

func NewMapTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	option := curveadm.MemStorage().Get(KEY_MAP_OPTION).(MapOption)
	user, volume := option.User, option.Volume
	subname := fmt.Sprintf("hostname=%s volume=%s", cc.GetHost(), volume)
	t := task.NewTask("Map Volume", subname, cc.GetSSHConfig())

	// add step
	var containerId string
	containerName := volume2ContainerName(user, volume)
	mapScriptPath := "/curvebs/nebd/sbin/map.sh"
	toolsConf := fmt.Sprintf(FORMAT_TOOLS_CONF, cc.GetClusterMDSAddr())
	mapScript := scripts.MAP
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", containerName),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2CheckNEBDClient{
		containerId: &containerId,
		user:        user,
		volume:      volume,
	})
	t.AddStep(&step2CreateNBDDevice{
		execWithSudo:  true,
		execInLocal:   false,
		execSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.PullImage{
		Image:         cc.GetContainerImage(),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.CreateContainer{
		Image:         cc.GetContainerImage(),
		Command:       fmt.Sprintf("%s %s %s %v %d", mapScriptPath, user, volume, option.Create, option.Size),
		Entrypoint:    "/bin/bash",
		Envs:          []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Name:          containerName,
		Pid:           "host",
		Privileged:    true,
		Volumes:       getVolumes(cc),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{ // install tools.conf
		Content:           &toolsConf,
		ContainerId:       &containerId,
		ContainerDestPath: "/etc/curve/tools.conf",
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{ // install map.sh
		Content:           &mapScript,
		ContainerId:       &containerId,
		ContainerDestPath: mapScriptPath,
		ExecWithSudo:      true,
		ExecInLocal:       false,
		ExecSudoAlias:     curveadm.SudoAlias(),
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
		ExecSudoAlias:     curveadm.SudoAlias(),
	})
	t.AddStep(&step.StartContainer{
		ContainerId:   &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&waitMapDone{
		ContainerId:    &containerId,
		ExecWithSudo:   false,
		ExecInLocal:    true,
		ExecTimeoutSec: 10,
		ExecSudoAlias:  curveadm.SudoAlias(),
	})

	return t, nil
}

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
	"github.com/opencurve/curveadm/internal/configure/client/bs"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	DEFAULT_NEBD_CONTAINER_NAME = "curvebs-nebd-server"

	CLIENT_CONFIG_DELIMITER = "="
)

type (
	step2CheckNEBDServer struct{ containerId *string }
)

func (s *step2CheckNEBDServer) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return task.ERR_SKIP_TASK
	}
	return nil
}

func newMutate(cc *bs.ClientConfig, delimiter string) step.Mutate {
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

func getVolumes(cc *client.ClientConfig) []step.Volume {
	volumes := []step.Volume{
		{HostPath: "/dev", ContainerPath: "/dev"},
		{HostPath: "/lib/modules", ContainerPath: "/lib/modules"},
		{HostPath: cc.GetDataDir(), ContainerPath: "/curvebs/nebd/data"},
	}
	if len(cc.GetLogDir()) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      cc.GetLogDir(),
			ContainerPath: "/curvebs/nebd/logs",
		})
	}
	return volumes
}

func NewStartNEBDServiceTask(curvradm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	subname := fmt.Sprintf("hostname=%s image=%s", cc.GetHost(), cc.GetContainerImage())
	t := task.NewTask("Start NEBD Service", subname, cc.GetSSHConfig())

	// add step
	var containerId string
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", DEFAULT_NEBD_CONTAINER_NAME),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curvradm.SudoAlias(),
	})
	t.AddStep(&step2CheckNEBDServer{ // skip if nebd-server exist
		containerId: &containerId,
	})
	t.AddStep(&step.CreateDirectory{
		Paths:         []string{cc.GetLogDir(), cc.GetDataDir()},
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curvradm.SudoAlias(),
	})
	t.AddStep(&step.PullImage{
		Image:         cc.GetContainerImage(),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curvradm.SudoAlias(),
	})
	t.AddStep(&step.CreateContainer{
		Image:         cc.GetContainerImage(),
		Envs:          []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Command:       fmt.Sprintf("--role nebd"),
		Name:          DEFAULT_NEBD_CONTAINER_NAME,
		Privileged:    true,
		Volumes:       getVolumes(cc),
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curvradm.SudoAlias(),
	})
	for _, filename := range []string{"client.conf", "nebd-server.conf"} {
		t.AddStep(&step.SyncFile{
			ContainerSrcId:    &containerId,
			ContainerSrcPath:  "/curvebs/conf/" + filename,
			ContainerDestId:   &containerId,
			ContainerDestPath: "/curvebs/nebd/conf/" + filename,
			KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
			Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
			ExecWithSudo:      true,
			ExecInLocal:       false,
			ExecSudoAlias:     curvradm.SudoAlias(),
		})
	}
	t.AddStep(&step.StartContainer{
		ContainerId:   &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curvradm.SudoAlias(),
	})

	return t, nil
}

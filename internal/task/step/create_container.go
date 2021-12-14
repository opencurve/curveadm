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
 * Created Date: 2021-12-12
 * Author: Jingli Chen (Wine93)
 */

package step

import (
	"strings"

	"github.com/opencurve/curveadm/internal/task/context"
)

type (
	Volume struct { // bind mount a volume
		HostPath      string
		ContainerPath string
	}

	CreateContainer struct {
		Image        string
		Command      string
		Name         string
		Volumes      []Volume
		Restart      string
		ExecWithSudo bool
		Out          *string
	}
)

func (s *CreateContainer) Execute(ctx *context.Context) error {
	cli := ctx.DockerCli().CreateContainer(s.Image, s.Command)
	if len(s.Name) > 0 {
		cli.AddOption("--name %s", s.Name)
	}
	for _, volume := range s.Volumes {
		cli.AddOption("--volume %s:%s", volume.HostPath, volume.ContainerPath)
	}
	if len(s.Restart) > 0 {
		cli.AddOption("--restart %s", s.Restart)
	}
	cli.AddOption("--network host")

	out, err := cli.Execute(context.ExecOption{Sudo: s.ExecWithSudo})
	*s.Out = strings.TrimSuffix(out, "\n")
	return err
}

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
 * Created Date: 2021-12-13
 * Author: Jingli Chen (Wine93)
 */

package step

import (
	"fmt"

	"github.com/opencurve/curveadm/internal/task/context"
)

type ReadFile struct {
	HostSrcPath      string
	ContainerId      string
	ContainerSrcPath string
	Content          *string
}

func (s *ReadFile) Execute(ctx *context.Context) error {
	if len(s.HostDestPath) > 0 {
		return ctx.Sshell().InstallFile(s.Content, s.HostDestPath)
	}

	if len(s.ContainerId) == 0 || len(s.ContainerDestPath) == 0 {
		return "", fmt.Errorf()
	}

	defer func() {
		ctx.Sshell().RemoveFile()
	}()
	return ctx.DockerCli().CopyFromContainer(s.ContainerId, s.DestPath)
}

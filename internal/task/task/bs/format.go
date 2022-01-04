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

package bs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/format"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	DEFAULT_CHUNKFILE_SIZE = 16 * 1024 * 1024 // 16MB
)

type (
	step2SkipFormat struct {
		device      string
		containerId *string
	}
)

func (s *step2SkipFormat) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return fmt.Errorf("'%s': busy formatting", s.device)
	}
	return nil
}

func genFormatCommand(layout topology.Layout, percent int) string {
	args := []string{
		fmt.Sprintf("-allocatePercent=%d", percent),
		fmt.Sprintf("-fileSize=%d", DEFAULT_CHUNKFILE_SIZE),
		fmt.Sprintf("-filePoolDir=%s", layout.ChunkfilePoolDir),
		fmt.Sprintf("-filePoolMetaPath=%s", layout.ChunkfilePoolMetaPath),
		fmt.Sprintf("-fileSystemPath=%s", layout.ChunkfilePoolDir),
	}
	//return fmt.Sprintf("/bin/bash -c 'mkdir -p %s'", layout.ChunkfilePoolDir)
	return fmt.Sprintf("'mkdir -p %s && %s %s'", layout.ChunkfilePoolDir, layout.FormatBinaryPath, strings.Join(args, " "))
}

func device2ContainerName(device string) string {
	return utils.MD5Sum(device)
}

func NewFormatChunkfilePoolTask(curveadm *cli.CurveAdm, fc *format.FormatConfig) (*task.Task, error) {
	device := fc.GetDevice()
	mountPoint := fc.GetMountPoint()
	usagePercent := fc.GetUsagePercent()
	subname := fmt.Sprintf("host=%s device=%s mountPoint=%s usage=%d%%",
		fc.GetHost(), device, mountPoint, usagePercent)
	t := task.NewTask("Start Format Chunkfile Pool", subname, fc.GetSSHConfig())

	// add step
	var oldContainerId, containerId string
	containerName := device2ContainerName(device)
	layout := topology.GetCurveBSProjectLayout()
	toolsDataDir := layout.ToolsDataDir
	// 1: skip if formating container exist
	t.AddStep(&step.ListContainers{
		ShowAll:      true,
		Format:       "'{{.ID}}'",
		Quiet:        true,
		Filter:       fmt.Sprintf("name=%s", containerName),
		Out:          &oldContainerId,
		ExecInLocal:  false,
		ExecWithSudo: true,
	})
	t.AddStep(&step2SkipFormat{
		device:      device,
		containerId: &oldContainerId,
	})
	// 2: mkfs and mount device
	t.AddStep(&step.UmountFilesystem{
		Directory:      device,
		IgnoreUmounted: true,
		IgnoreNotFound: true,
		ExecInLocal:    false,
		ExecWithSudo:   true,
	})
	t.AddStep(&step.CreateDirectory{
		Paths:        []string{mountPoint},
		ExecInLocal:  false,
		ExecWithSudo: true,
	})
	/*
		t.AddStep(&step.CreateFilesystem{ // mkfs.ext4 MOUNT_POINT
			Device:       device,
			ExecInLocal:  false,
			ExecWithSudo: true,
		})
	*/
	t.AddStep(&step.MountFilesystem{
		Source:       device,
		Directory:    mountPoint,
		ExecInLocal:  false,
		ExecWithSudo: true,
	})
	// 3: run container to format chunkfile pool
	t.AddStep(&step.PullImage{
		Image:        fc.GetContainerIamge(),
		ExecInLocal:  false,
		ExecWithSudo: true,
	})

	t.AddStep(&step.CreateContainer{
		Image:        fc.GetContainerIamge(),
		Command:      genFormatCommand(layout, usagePercent),
		Entrypoint:   "/bin/bash",
		Volumes:      []step.Volume{{HostPath: mountPoint, ContainerPath: toolsDataDir}},
		Out:          &containerId,
		ExecInLocal:  false,
		ExecWithSudo: true,
	})
	t.AddStep(&step.InstallFile{})
	t.AddStep(&step.StartContainer{
		ContainerId:  &containerId,
		ExecInLocal:  false,
		ExecWithSudo: true,
	})
	return t, nil
}

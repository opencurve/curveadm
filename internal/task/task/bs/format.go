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
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	DEFAULT_CHUNKFILE_SIZE        = 16 * 1024 * 1024 // 16MB
	DEFAULT_CHUNKFILE_HEADER_SIZE = 4 * 1024         // 4KB
)

type (
	step2SkipFormat struct {
		device      string
		containerId *string
	}
)

func (s *step2SkipFormat) Execute(ctx *context.Context) error {
	if len(*s.containerId) > 0 {
		return task.ERR_SKIP_TASK
	}
	return nil
}

func device2ContainerName(device string) string {
	return utils.MD5Sum(device)
}

func NewPersistMountPointTask(curveadm *cli.CurveAdm, fcs []*format.FormatConfig) (*task.Task, error) {
	// create task
	subname := fmt.Sprintf("host=%s", fcs[0].GetHost())
	t := task.NewTask("Start Persist Mount Point", subname, fcs[0].GetSSHConfig())

	// persist mount points with fstab file.
	var devices string
	var mountPoints string
	for _, fc := range fcs {
		// add device
		device_arr := strings.Split(fc.GetDevice(), "/")
		device_lastname := device_arr[len(device_arr)-1]
		devices = devices + " " + device_lastname

		// add mount point
		mountPoints = mountPoints + " " + fc.GetMountPoint()
	}
	t.AddStep(&step.ExecScript{
		Content:       &scripts.PERSIST_MOUNTPOINTS,
		Device:        devices,
		MountPoint:    mountPoints,
		Mode:          "777",
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	return t, nil
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
	chunkfilePoolRootDir := layout.ChunkfilePoolRootDir
	formatScript := scripts.SCRIPT_FORMAT
	formatScriptPath := fmt.Sprintf("%s/format.sh", layout.ToolsBinDir)
	formatCommand := fmt.Sprintf("%s %s %d %d %s %s", formatScriptPath, layout.FormatBinaryPath,
		usagePercent, DEFAULT_CHUNKFILE_SIZE, layout.ChunkfilePoolDir, layout.ChunkfilePoolMetaPath)
	// 1: skip if formating container exist
	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", containerName),
		Out:           &oldContainerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2SkipFormat{
		device:      device,
		containerId: &oldContainerId,
	})
	// 2: mkfs and mount device
	t.AddStep(&step.UmountFilesystem{
		Directorys:     []string{device},
		IgnoreUmounted: true,
		IgnoreNotFound: true,
		ExecWithSudo:   true,
		ExecInLocal:    false,
		ExecSudoAlias:  curveadm.SudoAlias(),
	})
	t.AddStep(&step.CreateDirectory{
		Paths:         []string{mountPoint},
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.CreateFilesystem{ // mkfs.ext4 MOUNT_POINT
		Device:        device,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.MountFilesystem{
		Source:        device,
		Directory:     mountPoint,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	// 3: run container to format chunkfile pool
	t.AddStep(&step.PullImage{
		Image:         fc.GetContainerIamge(),
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.CreateContainer{
		Image:         fc.GetContainerIamge(),
		Command:       formatCommand,
		Entrypoint:    "/bin/bash",
		Name:          containerName,
		Remove:        true,
		Volumes:       []step.Volume{{HostPath: mountPoint, ContainerPath: chunkfilePoolRootDir}},
		Out:           &containerId,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step.InstallFile{
		ContainerId:       &containerId,
		ContainerDestPath: formatScriptPath,
		Content:           &formatScript,
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
	return t, nil
}

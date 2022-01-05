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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

package fs

import (
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/client"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/scripts"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	KEY_MOUNT_FSNAME = "MOUNT_FSNAME"
	KEY_MOUNT_POINT  = "MOUNT_POINT"

	FORMAT_MOUNT_OPTION = "type=bind,source=%s,target=%s,bind-propagation=rshared"

	CLIENT_CONFIG_DELIMITER = "="

	RPC_TIMEOUT_MS = 10000
)

var (
	FORMAT_FUSE_ARGS = []string{
		"-f",
		"-o default_permissions",
		"-o allow_other",
		"-o fsname=%s", // fsname
		"-o fstype=s3",
		"-o user=curvefs",
		"-o conf=%s", // config path
		"%s",         // mount path
	}
)

func getMountCommand(cc *client.ClientConfig, mountFSName string) string {
	format := strings.Join(FORMAT_FUSE_ARGS, " ")
	fuseArgs := fmt.Sprintf(format, mountFSName, cc.GetClientConfPath(), cc.GetClientMountPath())
	return fmt.Sprintf("/client.sh %s --role=client --args='%s'", mountFSName, fuseArgs)
}

func getMountVolumes(cc *client.ClientConfig) []step.Volume {
	volumes := []step.Volume{}
	prefix := cc.GetClientPrefix()
	logDir := cc.GetLogDir()
	dataDir := cc.GetDataDir()
	coreDir := cc.GetCoreDir()

	if len(logDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      logDir,
			ContainerPath: fmt.Sprintf("%s/logs", prefix),
		})
	}

	if len(dataDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      dataDir,
			ContainerPath: fmt.Sprintf("%s/data", prefix),
		})
	}

	if len(coreDir) > 0 {
		volumes = append(volumes, step.Volume{
			HostPath:      coreDir,
			ContainerPath: cc.GetCoreLocateDir(),
		})
	}

	return volumes
}

func newMutate(cc *client.ClientConfig, delimiter string) step.Mutate {
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

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func newToolsMutate(cc *client.ClientConfig, delimiter string) step.Mutate {
	clientConfig := cc.GetServiceConfig()
	tools2client := map[string]string{
		"mdsAddr" : "mdsOpt.rpcRetryOpt.addrs",
	}
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}
		replaceKey := key
		if tools2client[key] != "" {
			replaceKey = tools2client[key]
		}
		v, ok := clientConfig[strings.ToLower(replaceKey)]
		if ok {
			value = v
		}
		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func NewMountFSTask(curvradm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	mountPoint := curvradm.MemStorage().Get(KEY_MOUNT_POINT).(string)
	mountFSName := curvradm.MemStorage().Get(KEY_MOUNT_FSNAME).(string)
	subname := fmt.Sprintf("mountFSName=%s mountPoint=%s", mountFSName, mountPoint)
	t := task.NewTask("Mount FileSystem", subname, nil)

	// add step
	var containerId string
	root := cc.GetCurveFSPrefix()
	prefix := cc.GetClientPrefix()
	containerMountPath := cc.GetClientMountPath()
	createfsScript := scripts.SCRIPT_CREATEFS
	createfsScriptPath := "/client.sh"
	t.AddStep(&step.PullImage{
		Image:        cc.GetContainerImage(),
		ExecWithSudo: false,
		ExecInLocal:  true,
	})
	t.AddStep(&step.CreateContainer{
		Image:             cc.GetContainerImage(),
		Envs:              []string{"LD_PRELOAD=/usr/local/lib/libjemalloc.so"},
		Command:           getMountCommand(cc, mountFSName),
		Name:              mountPoint2ContainerName(mountPoint),
		Mount:             fmt.Sprintf(FORMAT_MOUNT_OPTION, mountPoint, containerMountPath),
		Volumes:           getMountVolumes(cc),
		Devices:           []string{"/dev/fuse"},
		SecurityOptions:   []string{"apparmor:unconfined"},
		LinuxCapabilities: []string{"SYS_ADMIN"},
		Ulimits:           []string{"core=-1"},
		Pid:               cc.GetContainerPid(),
		Privileged:        true,
		Out:               &containerId,
		ExecWithSudo:      false,
		ExecInLocal:       true,
		Entrypoint:        "/bin/bash",
	})
	t.AddStep(&step.SyncFile{ // sync service config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  fmt.Sprintf("%s/conf/client.conf", root),
		ContainerDestId:   &containerId,
		ContainerDestPath: fmt.Sprintf("%s/conf/client.conf", prefix),
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecWithSudo:      false,
		ExecInLocal:       true,
	})
	t.AddStep(&step.SyncFile{ // sync tools config
		ContainerSrcId:    &containerId,
		ContainerSrcPath:  fmt.Sprintf("%s/conf/tools.conf", root),
		ContainerDestId:   &containerId,
		ContainerDestPath: topology.GetProjectLayout(topology.KIND_CURVEFS).ToolsConfSystemPath,
		KVFieldSplit:      CLIENT_CONFIG_DELIMITER,
		Mutate:            newToolsMutate(cc, CLIENT_CONFIG_DELIMITER),
		ExecWithSudo:      false,
		ExecInLocal:       true,
	})
	t.AddStep(&step.InstallFile{ // install client.sh shell
		ContainerId:       &containerId,
		ContainerDestPath: createfsScriptPath,
		Content:           &createfsScript,
		ExecWithSudo:      false,
		ExecInLocal:       true,
	})
	t.AddStep(&step.StartContainer{
		ContainerId:  &containerId,
		ExecWithSudo: false,
		ExecInLocal:  true,
	})

	return t, nil
}

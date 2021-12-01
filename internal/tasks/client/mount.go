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

package client

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
)

const (
	REGEX_FORMAT     = "^([^=#]+)=.*$" // key: mu[1]
	kEY_CONTAINER_ID = "CONTAINER_ID"
)

var (
	FUSE_ARGS = []string{
		"-f",
		"-o default_permissions",
		"-o allow_other",
		"-o fsname=%s", // fsname
		"-o fstype=s3",
		"-o user=curvefs",
		"-o conf=%s", // config path
		"%s",         // mount path
	}

	CREATE_CONTAINER_AGRS = []string{
		"--mount type=bind,source=%s,target=%s,bind-propagation=rshared", // host mount path, container mount path
		"%s", // mount volumes (dataDir and logDir)
		"--cap-add SYS_ADMIN",
		"--device=/dev/fuse",
		"--security-opt apparmor:unconfined",
		"--name %s",                    // container name
		"%s --role=client --args='%s'", // container image, start arguments
	}
)

type (
	step2CreateContainer struct {
		mountPoint  string
		mountFSName string
		config      *configure.ClientConfig
	}
	step2SyncConfig struct {
		tempDir string
		config  *configure.ClientConfig
	}
	step2StartContainer struct{}
)

func (s *step2CreateContainer) fuseArgs() string {
	format := strings.Join(FUSE_ARGS, " ")
	return fmt.Sprintf(format, s.mountFSName, s.config.GetProjectConfPath(), s.config.GetProjectMountPath())
}

// -v hostPath1:conatinerPath1 -v hostPath2:conatinerPath2
func (s *step2CreateContainer) volumeArgs(ctx *task.Context) string {
	volumes := []string{}
	config := s.config
	prefix := config.GetClientPrefix()
	logDir := config.GetLogDir()
	dataDir := config.GetDataDir()

	if logDir != "" {
		hostPath := logDir
		containerPath := prefix + "/logs"
		volumes = append(volumes, fmt.Sprintf("-v %s:%s", hostPath, containerPath))
	}

	if dataDir != "" {
		hostPath := dataDir
		containerPath := prefix + "/data"
		volumes = append(volumes, fmt.Sprintf("-v %s:%s", hostPath, containerPath))
	}

	return strings.Join(volumes, " ")
}

func (s *step2CreateContainer) Execute(ctx *task.Context) error {
	status, err := getMountStatus(ctx, s.mountPoint)
	if err != nil {
		return err
	} else if status.Status != STATUS_UNMOUNTED {
		return fmt.Errorf("path mounted, please use 'curveadm umount %s' first", s.mountPoint)
	}

	hostMountPath := s.mountPoint
	containerMountPath := s.config.GetProjectMountPath()
	mountVolumes := s.volumeArgs(ctx)
	containerName := status.ContainerName
	containerImage := s.config.GetContainerImage()
	format := strings.Join(CREATE_CONTAINER_AGRS, " ")
	args := fmt.Sprintf(format, hostMountPath, containerMountPath, mountVolumes, containerName, containerImage, s.fuseArgs())
	if out, err := ctx.Module().LocalShell("sudo docker create %s", args); err != nil {
		return err
	} else {
		ctx.Register().Set(kEY_CONTAINER_ID, tui.TrimContainerId(out))
	}
	return nil
}

func (s *step2CreateContainer) Rollback(ctx *task.Context) {
}

func (s *step2SyncConfig) replace(r *regexp.Regexp, line string) string {
	mu := r.FindStringSubmatch(line)
	if len(mu) == 0 {
		return line
	}

	key := mu[1]
	serviceConfig := s.config.GetServiceConfig()
	if value, ok := serviceConfig[strings.ToLower(key)]; ok {
		return fmt.Sprintf("%s=%s", key, value)
	}

	return line
}

func (s *step2SyncConfig) readConfigFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// regex
	r, err := regexp.Compile(REGEX_FORMAT)
	if err != nil {
		return "", err
	}

	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := s.replace(r, scanner.Text())
		lines = append(lines, line)
	}

	return strings.Join(lines, "\n"), nil
}

func (s *step2SyncConfig) writeConfigFile(filename, text string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}

func (s *step2SyncConfig) Execute(ctx *task.Context) error {
	file, err := os.CreateTemp(s.tempDir, "curvefs_client_*")
	if err != nil {
		return err
	}
	defer file.Close()

	containerId := ctx.Register().Get(kEY_CONTAINER_ID)
	containerSrcPath := s.config.GetCurveFSConfPath()
	containerDstPath := s.config.GetProjectConfPath()
	tempPath := file.Name()
	cmd := fmt.Sprintf("sudo docker cp %s:%s %s", containerId, containerSrcPath, tempPath)
	if _, err := ctx.Module().LocalShell(cmd); err != nil { // copy client config from container
		return err
	} else if lines, err := s.readConfigFile(tempPath); err != nil {
		return err
	} else if err := s.writeConfigFile(tempPath, lines); err != nil {
		return err
	}

	_, err = ctx.Module().LocalShell("sudo docker cp %s %s:%s", tempPath, containerId, containerDstPath)
	return err
}

func (s *step2SyncConfig) Rollback(ctx *task.Context) {
}

func (s *step2StartContainer) Execute(ctx *task.Context) error {
	containerId := ctx.Register().Get(kEY_CONTAINER_ID)
	_, err := ctx.Module().LocalShell("sudo docker start %s", containerId)
	return err
}

func (s *step2StartContainer) Rollback(ctx *task.Context) {
}

func NewMountFSTask(curvradm *cli.CurveAdm, mountPoint, mountFSName string, config *configure.ClientConfig) (*task.Task, error) {
	t := task.NewTask("Mount FileSystem", "", nil)
	t.AddStep(&step2CreateContainer{mountFSName: mountFSName, mountPoint: mountPoint, config: config})
	t.AddStep(&step2SyncConfig{tempDir: curvradm.TempDir(), config: config})
	t.AddStep(&step2StartContainer{})
	return t, nil
}

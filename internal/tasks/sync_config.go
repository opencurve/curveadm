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

package tasks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/log"
	"github.com/opencurve/curveadm/internal/tasks/task"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	ETCD_DSV     = ": "
	DEFAULT_DSV  = "="
	REGEX_FORMAT = "^(([^%s]+)%s\\s*)([^\\s#]*)" // key: mu[2] value: mu[3]

	KEY_LOCAL_PATH         = "LOCAL_PATH"
	KEY_REMOTE_PATH        = "REMOTE_PATH"
	KEY_CONTAINER_SRC_PATH = "CONTAINER_SRC_PATH"
	KEY_CONTAINER_DST_PATH = "CONTAINER_DST_PATH"
)

type (
	syncItem struct {
		tempDir          string
		serviceId        string
		containerId      string
		containerSrcPath string
		containerDstPath string
		configDelimiter  string
	}
)

type (
	step2InitSyncConfig struct {
		tempDir          string
		serviceId        string
		containerSrcPath string
		containerDstPath string
	}
	step2CopyFileFromRemote struct{ containerId string }
	step2RenderingConfig    struct{ delimiter string }
	step2CopyFileToRemote   struct{ containerId string }

	step2CreateDirectory struct {
		containerDstDir string
		containerId     string
	}
)

func (s *step2InitSyncConfig) Execute(ctx *task.Context) error {
	file, err := os.CreateTemp(s.tempDir, fmt.Sprintf("%s_*", s.serviceId))
	if err != nil {
		return err
	}
	defer file.Close()

	localPath := file.Name()
	remotePath := fmt.Sprintf("/tmp/%s", filepath.Base(localPath))
	containerSrcPath := s.containerSrcPath
	containerDstPath := s.containerDstPath
	ctx.Register().Set(KEY_LOCAL_PATH, localPath)
	ctx.Register().Set(KEY_REMOTE_PATH, remotePath)
	ctx.Register().Set(KEY_CONTAINER_SRC_PATH, containerSrcPath)
	ctx.Register().Set(KEY_CONTAINER_DST_PATH, containerDstPath)

	log.Info("SyncConfig",
		log.Field(KEY_LOCAL_PATH, localPath),
		log.Field(KEY_REMOTE_PATH, remotePath),
		log.Field(KEY_CONTAINER_SRC_PATH, containerSrcPath),
		log.Field(KEY_CONTAINER_DST_PATH, containerDstPath))

	return nil
}

func (s *step2InitSyncConfig) Rollback(ctx *task.Context) {
}

func (s *step2CopyFileFromRemote) Execute(ctx *task.Context) error {
	localPath := ctx.Register().Get(KEY_LOCAL_PATH).(string)
	remotePath := ctx.Register().Get(KEY_REMOTE_PATH).(string)
	containerSrcPath := ctx.Register().Get(KEY_CONTAINER_SRC_PATH).(string)
	cmd := fmt.Sprintf("sudo docker cp %s:%s %s", s.containerId, containerSrcPath, remotePath)
	if _, err := ctx.Module().SshShell(cmd); err != nil {
		return err
	} else if err := ctx.Module().Download(remotePath, localPath); err != nil {
		return err
	} else if _, err := ctx.Module().SshShell("sudo rm %s", remotePath); err != nil {
		return err
	}
	return nil
}

func (s *step2CopyFileFromRemote) Rollback(ctx *task.Context) {
}

func (s *step2RenderingConfig) replace(r *regexp.Regexp, line string, dc *configure.DeployConfig) (string, error) {
	mu := r.FindStringSubmatch(line)
	if len(mu) == 0 {
		return line, nil
	}

	key := mu[2]
	value := mu[3]
	serviceConfig := dc.GetServiceConfig()
	if v, ok := serviceConfig[strings.ToLower(key)]; ok { // replace
		value = v
	}

	var err error
	value, err = dc.GetVariables().Rendering(value)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%s", mu[1], value), nil
}

func (s *step2RenderingConfig) readConfigFile(ctx *task.Context) (string, error) {
	localPath := ctx.Register().Get(KEY_LOCAL_PATH).(string)
	file, err := os.Open(localPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// regex
	pattern := fmt.Sprintf(REGEX_FORMAT, s.delimiter, s.delimiter)
	r, err := regexp.Compile(pattern)
	if err != nil {
		return "", err
	}

	// replace line one by one
	lines := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line, err := s.replace(r, scanner.Text(), ctx.Config())
		if err != nil {
			return "", err
		}
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n"), nil
}

func (s *step2RenderingConfig) writeConfigFile(ctx *task.Context, text string) error {
	localPath := ctx.Register().Get(KEY_LOCAL_PATH).(string)
	file, err := os.Create(localPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(text)
	return err
}

func (s *step2RenderingConfig) Execute(ctx *task.Context) error {
	text, err := s.readConfigFile(ctx)
	if err != nil {
		return err
	}

	return s.writeConfigFile(ctx, text)
}

func (s *step2RenderingConfig) Rollback(ctx *task.Context) {
}

func (s *step2CopyFileToRemote) Execute(ctx *task.Context) error {
	localPath := ctx.Register().Get(KEY_LOCAL_PATH).(string)
	remotePath := ctx.Register().Get(KEY_REMOTE_PATH).(string)
	containerDstPath := ctx.Register().Get(KEY_CONTAINER_DST_PATH).(string)
	if err := ctx.Module().Scp(localPath, remotePath); err != nil {
		return err
	}

	_, err := ctx.Module().SshShell("sudo docker cp %s %s:%s", remotePath, s.containerId, containerDstPath)
	return err
}

func (s *step2CopyFileToRemote) Rollback(ctx *task.Context) {
}

func (s *step2CreateDirectory) Execute(ctx *task.Context) error {
	tempDir := fmt.Sprintf("/tmp/%s", utils.RandString(10))
	_, err := ctx.Module().SshShell("mkdir -p %s", tempDir)
	if err != nil {
		return err
	}
	defer func() {
		ctx.Module().SshShell("rm -r %s", tempDir)
	}()

	cmd := fmt.Sprintf("sudo docker cp %s %s:%s", tempDir, s.containerId, s.containerDstDir)
	_, err = ctx.Module().SshShell(cmd)
	return err
}

func (s *step2CreateDirectory) Rollback(ctx *task.Context) {
}

func addSyncItem(t *task.Task, item syncItem) {
	t.AddStep(&step2InitSyncConfig{
		tempDir:          item.tempDir,
		serviceId:        item.serviceId,
		containerSrcPath: item.containerSrcPath,
		containerDstPath: item.containerDstPath,
	})
	t.AddStep(&step2CopyFileFromRemote{containerId: item.containerId})
	t.AddStep(&step2RenderingConfig{delimiter: item.configDelimiter})
	t.AddStep(&step2CopyFileToRemote{containerId: item.containerId})
}

func NewSyncConfigTask(curveadm *cli.CurveAdm, dc *configure.DeployConfig) (*task.Task, error) {
	serviceId := configure.ServiceId(curveadm.ClusterId(), dc.GetId())
	containerId, err := curveadm.Storage().GetContainerId(serviceId)
	if err != nil {
		return nil, err
	} else if containerId == "" {
		return nil, fmt.Errorf("service(id=%s) not found", serviceId)
	}

	subname := fmt.Sprintf("host=%s role=%s containerId=%s",
		dc.GetHost(), dc.GetRole(), tui.TrimContainerId(containerId))
	t := task.NewTask("Sync Config", subname, dc)

	delimiter := DEFAULT_DSV
	if dc.GetRole() == configure.ROLE_ETCD {
		delimiter = ETCD_DSV
	}
	addSyncItem(t, syncItem{
		tempDir:     curveadm.TempDir(),
		serviceId:   serviceId,
		containerId: containerId,
		// ex: /usr/local/curvefs/conf/mds.conf
		containerSrcPath: fmt.Sprintf("%s/conf/%s.conf", dc.GetCurveFSPrefix(), dc.GetRole()),
		// ex: /usr/local/curvefs/mds/conf/mds.conf
		containerDstPath: fmt.Sprintf("%s/conf/%s.conf", dc.GetServicePrefix(), dc.GetRole()),
		configDelimiter:  delimiter,
	})

	t.AddStep(&step2CreateDirectory{ // it's a trick
		containerId:     containerId,
		containerDstDir: "/etc/curvefs",
	})
	addSyncItem(t, syncItem{
		tempDir:          curveadm.TempDir(),
		serviceId:        serviceId,
		containerId:      containerId,
		containerSrcPath: fmt.Sprintf("%s/conf/tools.conf", dc.GetCurveFSPrefix()),
		containerDstPath: "/etc/curvefs/tools.conf",
		configDelimiter:  DEFAULT_DSV,
	})

	return t, nil
}

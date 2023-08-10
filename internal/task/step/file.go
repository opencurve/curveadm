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
 * Created Date: 2021-12-15
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package step

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	TEMP_DIR       = "/tmp"
	REGEX_KV_SPLIT = "^(([^%s]+)%s\\s*)([^\\s#]*)" // key: mu[2] value: mu[3]
)

type (
	ReadFile struct {
		HostSrcPath      string
		ContainerId      string
		ContainerSrcPath string
		Content          *string
		module.ExecOptions
	}

	InstallFile struct {
		Content           *string
		HostDestPath      string
		ContainerId       *string
		ContainerDestPath string
		module.ExecOptions
	}

	Mutate func(string, string, string) (string, error)

	Filter struct {
		KVFieldSplit string
		Mutate       Mutate
		Input        *string
		Output       *string
	}

	SyncFile struct {
		ContainerSrcId    *string
		ContainerSrcPath  string
		ContainerDestId   *string
		ContainerDestPath string
		KVFieldSplit      string
		Mutate            func(string, string, string) (string, error)
		module.ExecOptions
	}

	TrySyncFile struct {
		ContainerSrcId    *string
		ContainerSrcPath  string
		ContainerDestId   *string
		ContainerDestPath string
		KVFieldSplit      string
		Mutate            func(string, string, string) (string, error)
		module.ExecOptions
	}

	DownloadFile struct {
		RemotePath string
		LocalPath  string
		module.ExecOptions
	}
)

func (s *ReadFile) Execute(ctx *context.Context) error {
	// remotePath
	remotePath := s.HostSrcPath
	if len(s.HostSrcPath) > 0 {
		// do nothing
	} else {
		remotePath = utils.RandFilename(TEMP_DIR)
		// defer ctx.Module().Shell().Remove(remotePath).Execute(module.ExecOptions{})
		dockerCli := ctx.Module().DockerCli().CopyFromContainer(s.ContainerId, s.ContainerSrcPath, remotePath)
		_, err := dockerCli.Execute(s.ExecOptions)
		if err != nil {
			return errno.ERR_COPY_FROM_CONTAINER_FAILED.FD("(%s cp CONTAINER:SRC_PATH DEST_PATH)", s.ExecWithEngine).E(err)
		}
	}

	// localPath
	localPath := utils.RandFilename(TEMP_DIR)
	defer os.Remove(localPath)
	if !s.ExecInLocal {
		err := ctx.Module().File().Download(remotePath, localPath)
		if err != nil {
			return errno.ERR_DOWNLOAD_FILE_FROM_REMOTE_BY_SSH_FAILED.E(err)
		}
	} else {
		cmd := ctx.Module().Shell().Rename(remotePath, localPath)
		_, err := cmd.Execute(s.ExecOptions)
		if err != nil {
			return errno.ERR_RENAME_FILE_OR_DIRECTORY_FAILED.E(err)
		}
	}

	data, err := utils.ReadFile(localPath)
	*s.Content = data
	if err != nil {
		return errno.ERR_READ_FILE_FAILED.E(err)
	}
	return nil
}

func (s *InstallFile) Execute(ctx *context.Context) error {
	localPath := utils.RandFilename(TEMP_DIR)
	defer os.Remove(localPath)
	err := utils.WriteFile(localPath, *s.Content, 0644)
	if err != nil {
		return errno.ERR_WRITE_FILE_FAILED.E(err)
	}

	remotePath := utils.RandFilename(TEMP_DIR)
	if !s.ExecInLocal {
		// defer ctx.Module().Shell().Remove(remotePath).Execute(module.ExecOptions{})
		err = ctx.Module().File().Upload(localPath, remotePath)
		if err != nil {
			return errno.ERR_UPLOAD_FILE_TO_REMOTE_BY_SSH_FAILED.E(err)
		}
	} else {
		cmd := ctx.Module().Shell().Rename(localPath, remotePath)
		_, err := cmd.Execute(module.ExecOptions{
			ExecWithSudo:  false, // NOTE: file owner is me
			ExecInLocal:   s.ExecInLocal,
			ExecSudoAlias: s.ExecSudoAlias,
		})
		if err != nil {
			return errno.ERR_RENAME_FILE_OR_DIRECTORY_FAILED.E(err)
		}
	}

	if len(s.HostDestPath) > 0 {
		cmd := ctx.Module().Shell().Rename(remotePath, s.HostDestPath)
		_, err = cmd.Execute(s.ExecOptions)
		if err != nil {
			return errno.ERR_RENAME_FILE_OR_DIRECTORY_FAILED.E(err)
		}
	} else {
		cli := ctx.Module().DockerCli().CopyIntoContainer(remotePath, *s.ContainerId, s.ContainerDestPath)
		_, err = cli.Execute(s.ExecOptions)
		if err != nil {
			return errno.ERR_COPY_INTO_CONTAINER_FAILED.FD(" (%scp SRC_PATH CONTAINER:DEST_PATH)", s.ExecWithEngine).E(err)
		}
	}
	return nil
}

func (s *Filter) kvSplit(line string, key, value *string) error {
	pattern := fmt.Sprintf(REGEX_KV_SPLIT, s.KVFieldSplit, s.KVFieldSplit)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return errno.ERR_BUILD_REGEX_FAILED.E(err)
	}

	mu := regex.FindStringSubmatch(line)
	if len(mu) == 0 {
		*key = ""
		*value = ""
	} else {
		*key = mu[2]
		*value = mu[3]
	}
	return nil
}

func (s *Filter) Execute(ctx *context.Context) error {
	var key, value string
	output := []string{}
	scanner := bufio.NewScanner(strings.NewReader(string(*s.Input)))
	for scanner.Scan() {
		in := scanner.Text()
		if len(s.KVFieldSplit) > 0 {
			err := s.kvSplit(in, &key, &value)
			if err != nil {
				return err
			}
		}

		out, err := s.Mutate(in, key, value)
		if err != nil {
			return err
		}
		output = append(output, out)
	}

	*s.Output = strings.Join(output, "\n")
	return nil
}

func (s *SyncFile) Execute(ctx *context.Context) error {
	var input, output string
	steps := []task.Step{}
	steps = append(steps, &ReadFile{
		ContainerId:      *s.ContainerSrcId,
		ContainerSrcPath: s.ContainerSrcPath,
		Content:          &input,
		ExecOptions:      s.ExecOptions,
	})
	steps = append(steps, &Filter{
		KVFieldSplit: s.KVFieldSplit,
		Mutate:       s.Mutate,
		Input:        &input,
		Output:       &output,
	})
	steps = append(steps, &InstallFile{
		ContainerId:       s.ContainerDestId,
		ContainerDestPath: s.ContainerDestPath,
		Content:           &output,
		ExecOptions:       s.ExecOptions,
	})

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *DownloadFile) Execute(ctx *context.Context) error {
	return ctx.Module().File().Download(s.RemotePath, s.LocalPath)
}

func (s *TrySyncFile) Execute(ctx *context.Context) error {
	var input string
	step := &ReadFile{
		ContainerId:      *s.ContainerSrcId,
		ContainerSrcPath: s.ContainerSrcPath,
		Content:          &input,
		ExecOptions:      s.ExecOptions,
	}
	if err := step.Execute(ctx); err != nil {
		// no this file
		return nil
	}
	sync := SyncFile{
		ContainerSrcId:    s.ContainerSrcId,
		ContainerSrcPath:  s.ContainerSrcPath,
		ContainerDestId:   s.ContainerDestId,
		ContainerDestPath: s.ContainerDestPath,
		KVFieldSplit:      s.KVFieldSplit,
		Mutate:            s.Mutate,
		ExecOptions:       s.ExecOptions,
	}
	return sync.Execute(ctx)
}

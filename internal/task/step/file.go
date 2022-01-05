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

package step

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

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
		ExecWithSudo     bool
		ExecInLocal      bool
	}

	InstallFile struct {
		HostDestPath      string
		ContainerId       *string
		ContainerDestPath string
		Content           *string
		ExecWithSudo      bool
		ExecInLocal       bool
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
		ExecWithSudo      bool
		ExecInLocal       bool
	}
)

func (s *ReadFile) Execute(ctx *context.Context) error {
	// remotePath
	remotePath := s.HostSrcPath
	if len(s.HostSrcPath) > 0 {
		// do nothing
	} else {
		remotePath = utils.RandFilename(TEMP_DIR)
		// defer ctx.Module().Shell().Remove(remotePath).Execute(module.ExecOption{})
		dockerCli := ctx.Module().DockerCli().CopyFromContainer(s.ContainerId, s.ContainerSrcPath, remotePath)
		_, err := dockerCli.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
		if err != nil {
			return err
		}
	}

	// localPath
	localPath := utils.RandFilename(TEMP_DIR)
	defer os.Remove(localPath)
	if !s.ExecInLocal {
		err := ctx.Module().File().Download(remotePath, localPath)
		if err != nil {
			return err
		}
	} else {
		cmd := ctx.Module().Shell().Rename(remotePath, localPath)
		_, err := cmd.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
		if err != nil {
			return err
		}
	}

	data, err := utils.ReadFile(localPath)
	*s.Content = data
	return err
}

func (s *InstallFile) Execute(ctx *context.Context) error {
	localPath := utils.RandFilename(TEMP_DIR)
	defer os.Remove(localPath)
	err := utils.WriteFile(localPath, *s.Content)
	if err != nil {
		return err
	}

	remotePath := utils.RandFilename(TEMP_DIR)
	if !s.ExecInLocal {
		// defer ctx.Module().Shell().Remove(remotePath).Execute(module.ExecOption{})
		err = ctx.Module().File().Upload(localPath, remotePath)
		if err != nil {
			return err
		}
	} else {
		cmd := ctx.Module().Shell().Rename(localPath, remotePath)
		_, err := cmd.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
		if err != nil {
			return err
		}
	}

	if len(s.HostDestPath) > 0 {
		cmd := ctx.Module().Shell().Rename(remotePath, s.HostDestPath)
		_, err = cmd.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
	} else {
		cli := ctx.Module().DockerCli().CopyIntoContainer(remotePath, *s.ContainerId, s.ContainerDestPath)
		_, err = cli.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
	}
	return err
}

func (s *Filter) kvSplit(line string, key, value *string) error {
	pattern := fmt.Sprintf(REGEX_KV_SPLIT, s.KVFieldSplit, s.KVFieldSplit)
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return err
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
		ExecWithSudo:     s.ExecWithSudo,
		ExecInLocal:      s.ExecInLocal,
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
		ExecWithSudo:      s.ExecWithSudo,
		ExecInLocal:       s.ExecInLocal,
	})

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return err
		}
	}
	return nil
}

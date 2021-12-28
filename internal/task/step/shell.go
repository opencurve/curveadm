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
	"strings"

	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	ERR_NOT_MOUNTED = "not mounted"
)

type (
	CreateDirectory struct {
		Paths        []string
		ExecWithSudo bool
		ExecInLocal  bool
	}

	RemoveFile struct {
		Files        []string
		ExecWithSudo bool
		ExecInLocal  bool
	}

	CreateFilesystem struct {
		Device       string
		ExecWithSudo bool
		ExecInLocal  bool
	}

	MountFilesystem struct {
		Source       string
		Directory    string
		ExecWithSudo bool
		ExecInLocal  bool
	}

	UmountFilesystem struct {
		Directory      string
		IgnoreUmounted bool
		ExecWithSudo   bool
		ExecInLocal    bool
	}
)

func (s *CreateDirectory) Execute(ctx *context.Context) error {
	for _, path := range s.Paths {
		if len(path) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Mkdir(path)
		cmd.AddOption("--parents") // no error if existing, make parent directories as needed
		_, err := cmd.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *RemoveFile) Execute(ctx *context.Context) error {
	for _, file := range s.Files {
		if len(file) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Remove(file)
		cmd.AddOption("--force")     // ignore nonexistent files and arguments, never prompt
		cmd.AddOption("--recursive") // remove directories and their contents recursively
		_, err := cmd.Execute(module.ExecOption{
			ExecWithSudo: s.ExecWithSudo,
			ExecInLocal:  s.ExecInLocal,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *CreateFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mkfs(s.Device)
	// force mke2fs to create a filesystem, even if the specified device is not a partition
	// on a block special device, or if other parameters do not make sense
	cmd.AddOption("-F")
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo: s.ExecWithSudo,
		ExecInLocal:  s.ExecInLocal,
	})
	return err
}

func (s *MountFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mount(s.Source, s.Directory)
	_, err := cmd.Execute(module.ExecOption{
		ExecWithSudo: s.ExecWithSudo,
		ExecInLocal:  s.ExecInLocal,
	})
	return err
}

func (s *UmountFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Umount(s.Directory)
	out, err := cmd.Execute(module.ExecOption{
		ExecWithSudo: s.ExecWithSudo,
		ExecInLocal:  s.ExecInLocal,
	})
	if s.IgnoreUmounted && strings.HasPrefix(out, ERR_NOT_MOUNTED) {
		return nil
	}
	return err
}

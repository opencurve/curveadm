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
	"github.com/opencurve/curveadm/internal/task/context"
)

type (
	CreateDirectory struct {
		Paths []string
	}
)

func (s *CreateDirectory) Execute(ctx *context.Context) error {
	for _, path := range s.Paths {
		if len(path) == 0 {
			continue
		}
		out, err := ctx.Sshell().CreateDirectory(path).Execute()
		if err != nil {
			return err
		}
	}
	return nil
}

/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-05-19
 * Author: Jingli Chen (Wine93)
 */

package analyzer

import (
	"fmt"

	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

type ShellAnalyzer struct {
	step *step.Command
}

func NewShellAnalyzer() *ShellAnalyzer {
	return &ShellAnalyzer{
		step: &step.Command{},
	}
}

func (s *ShellAnalyzer) ParseOption(key, value string) error {
	step := s.step
	switch key {
	case "command":
		step.Command = value
	case "exec_with_sudo":
		step.ExecWithSudo = utils.IsTrueStr(value)
	case "exec_in_local":
		step.ExecInLocal = utils.IsTrueStr(value)
	default:
		return fmt.Errorf("unknown option '%s'", key)
	}
	return nil
}

func (s *ShellAnalyzer) BindVariable(param string, variable *Variable) error {
	step := s.step
	switch param {
	case "out":
		step.Out = &variable.Value
	default:
		return fmt.Errorf("param '%s' is not a pointer", param)
	}
	return nil
}

func (s *ShellAnalyzer) Build() (task.Step, error) {
	step := s.step
	if len(step.Command) == 0 {
		return nil, fmt.Errorf("Command is empty")
	}
	return step, nil
}

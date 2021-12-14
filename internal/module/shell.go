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
 * Created Date: 2021-12-14
 * Author: Jingli Chen (Wine93)
 */

package module

import (
	"fmt"
	"strings"
	"text/template"

	ssh "github.com/melbahja/goph"
)

// mkdir [OPTION]... DIRECTORY...
// rmdir [OPTION]... DIRECTORY...
const (
	TEMPLATE_MKDIR = "mkdir {{.options}} {{.directorys}}"
	TEMPLATE_RMDIR = "rmdir {{.options}} {{.directorys}}"
)

type Shell struct {
	sshClient *ssh.Client
	options   []string
	tmpl      *template.Template
	data      map[string]interface{}
}

func NewShell(sshClient *ssh.Client) *Shell {
	return &Shell{
		sshClient: sshClient,
		options:   []string{},
		tmpl:      nil,
		data:      map[string]interface{}{},
	}
}

func (s *Shell) AddOption(format string, args ...interface{}) *Shell {
	s.options = append(s.options, fmt.Sprintf(format, args...))
	return s
}

func (s *Shell) Mkdir(directory ...string) *Shell {
	s.tmpl = template.Must(template.New("mkdir").Parse(TEMPLATE_MKDIR))
	s.data["directorys"] = strings.Join(directory, " ")
	return s
}

func (s *Shell) Rmdir(directory ...string) *Shell {
	s.tmpl = template.Must(template.New("rmdir").Parse(TEMPLATE_RMDIR))
	s.data["directorys"] = strings.Join(directory, " ")
	return s
}

func (s *Shell) Execute(options ExecOption) (string, error) {
	s.data["options"] = strings.Join(s.options, " ")
	return execCommand(s.sshClient, s.tmpl, s.data, options)
}

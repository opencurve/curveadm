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
// rm [OPTION]... [FILE]...
// mv [OPTION]... SOURCE DEST
// umount [options] <directory>
const (
	TEMPLATE_MKDIR       = "mkdir {{.options}} {{.directorys}}"
	TEMPLATE_RMDIR       = "rmdir {{.options}} {{.directorys}}"
	TEMPLATE_REMOVE      = "rm {{.options}} {{.files}}"
	TEMPLATE_RENAME      = "mv {{.options}} {{.source}} {{.dest}}"
	TEMPLATE_MKFS        = "mkfs.ext4 {{.options}} {{.device}}"
	TEMPLATE_CHMOD       = "chmod {{.options}} {{.mode}} {{.file}}"
	TEMPLATE_MOUNT       = "mount {{.options}} {{.source}} {{.directory}}"
	TEMPLATE_UMOUNT      = "umount {{.options}} {{.directory}}"
	TEMPLATE_DISKFREE    = "df {{.options}} {{.files}}"
	TEMPLATE_COMMAND     = "bash -c '{{.command}}'"
	TEMPLATE_EXEC_SCEIPT = "{{.scriptPath}} {{.arguments}}"
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
	s.tmpl = template.Must(template.New("Mkdir").Parse(TEMPLATE_MKDIR))
	s.data["directorys"] = strings.Join(directory, " ")
	return s
}

func (s *Shell) Rmdir(directory ...string) *Shell {
	s.tmpl = template.Must(template.New("Rmdir").Parse(TEMPLATE_RMDIR))
	s.data["directorys"] = strings.Join(directory, " ")
	return s
}

func (s *Shell) Remove(file ...string) *Shell {
	s.tmpl = template.Must(template.New("Remove").Parse(TEMPLATE_REMOVE))
	s.data["files"] = strings.Join(file, " ")
	return s
}

func (s *Shell) Rename(source, dest string) *Shell {
	s.tmpl = template.Must(template.New("Rename").Parse(TEMPLATE_RENAME))
	s.data["source"] = source
	s.data["dest"] = dest
	return s
}

func (s *Shell) Mkfs(device string) *Shell {
	s.tmpl = template.Must(template.New("Mkfs").Parse(TEMPLATE_MKFS))
	s.data["device"] = device
	return s
}

func (s *Shell) Chmod(mode, file string) *Shell {
	s.tmpl = template.Must(template.New("Chmod").Parse(TEMPLATE_CHMOD))
	s.data["mode"] = mode
	s.data["file"] = file
	return s
}

func (s *Shell) Mount(source, directory string) *Shell {
	s.tmpl = template.Must(template.New("Mount").Parse(TEMPLATE_MOUNT))
	s.data["source"] = source
	s.data["directory"] = directory
	return s
}

func (s *Shell) Umount(directory string) *Shell {
	s.tmpl = template.Must(template.New("Umount").Parse(TEMPLATE_UMOUNT))
	s.data["directory"] = directory
	return s
}

func (s *Shell) DiskFree(file ...string) *Shell {
	s.tmpl = template.Must(template.New("DiskFree").Parse(TEMPLATE_DISKFREE))
	s.data["files"] = strings.Join(file, " ")
	return s
}

func (s *Shell) Command(command string) *Shell {
	s.tmpl = template.Must(template.New("Command").Parse(TEMPLATE_COMMAND))
	s.data["command"] = command
	return s
}

func (s *Shell) ExecScript(scriptPath string, args ...string) *Shell {
	s.tmpl = template.Must(template.New("ExecScript").Parse(TEMPLATE_EXEC_SCEIPT))
	s.data["scriptPath"] = scriptPath
	s.data["arguments"] = strings.Join(args, " ")
	return s
}

func (s *Shell) Execute(options ExecOption) (string, error) {
	s.data["options"] = strings.Join(s.options, " ")
	return execCommand(s.sshClient, s.tmpl, s.data, options)
}

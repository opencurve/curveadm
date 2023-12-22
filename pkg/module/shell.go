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

// __SIGN_BY_WINE93__

package module

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const (
	// text
	TEMPLATE_SED = "sed {{.options}} {{.files}}"

	// storage (filesystem/block)
	TEMPLATE_LIST     = "ls {{.options}} {{.files}}"
	TEMPLATE_MKDIR    = "mkdir {{.options}} {{.directorys}}"
	TEMPLATE_RMDIR    = "rmdir {{.options}} {{.directorys}}"
	TEMPLATE_REMOVE   = "rm {{.options}} {{.files}}"
	TEMPLATE_RENAME   = "mv {{.options}} {{.source}} {{.dest}}"
	TEMPLATE_COPY     = "cp {{.options}} {{.source}} {{.dest}}"
	TEMPLATE_CHMOD    = "chmod {{.options}} {{.mode}} {{.file}}"
	TEMPLATE_STAT     = "stat {{.options}} {{.files}}"
	TEMPLATE_CAT      = "cat {{.options}} {{.files}}"
	TEMPLATE_MKFS     = "mkfs.ext4 {{.options}} {{.device}}"
	TEMPLATE_MOUNT    = "mount {{.options}} {{.source}} {{.directory}}"
	TEMPLATE_UMOUNT   = "umount {{.options}} {{.directory}}"
	TEMPLATE_TUNE2FS  = "tune2fs {{.options}} {{.device}}"
	TEMPLATE_FUSER    = "fuser {{.options}} {{.names}}"
	TEMPLATE_DISKFREE = "df {{.options}} {{.files}}"
	TEMPLATE_LSBLK    = "lsblk {{.options}} {{.devices}}"
	TEMPLATE_BLKID    = "blkid {{.options}} {{.device}}"

	// network
	TEMPLATE_SS   = "ss {{.options}} '{{.filter}}'"
	TEMPLATE_PING = "ping {{.options}} {{.destination}}"
	TEMPLATE_CURL = "curl {{.options}} {{.url}}"

	// kernel
	TEMPLATE_WHOAMI   = "whoami"
	TEMPLATE_DATE     = "date {{.options}} {{.format}}"
	TEMPLATE_UNAME    = "uname {{.options}}"
	TEMPLATE_MODPROBE = "modprobe {{.options}} {{.modulename}} {{.arguments}}"
	TEMPLATE_MODINFO  = "modinfo {{.modulename}}"

	// others
	TEMPLATE_TAR  = "tar {{.options}} {{.file}}"
	TEMPLATE_DPKG = "dpkg {{.options}}"
	TEMPLATE_RPM  = "rpm {{.options}}"
	TEMPLATE_SCP  = "scp {{.options}} {{.source}} {{.user}}@{{.host}}:{{.target}}"

	// bash
	TEMPLATE_COMMAND     = "{{.command}}"
	TEMPLATE_BASH_SCEIPT = "bash {{.scriptPath}} {{.arguments}}"
)

// TODO(P1): support command pipe
type Shell struct {
	remoteClient RemoteClient
	options      []string
	tmpl         *template.Template
	data         map[string]interface{}
}

func NewShell(remoteClient RemoteClient) *Shell {
	return &Shell{
		remoteClient: remoteClient,
		options:      []string{},
		tmpl:         nil,
		data:         map[string]interface{}{},
	}
}

func (s *Shell) AddOption(format string, args ...interface{}) *Shell {
	s.options = append(s.options, fmt.Sprintf(format, args...))
	return s
}

func (s *Shell) String() (string, error) {
	buffer := bytes.NewBufferString("")
	s.data["options"] = strings.Join(s.options, " ")
	err := s.tmpl.Execute(buffer, s.data)
	if err != nil {
		return "", err
	}
	return buffer.String(), nil
}

func (s *Shell) Execute(options ExecOptions) (string, error) {
	s.data["options"] = strings.Join(s.options, " ")
	return execCommand(s.remoteClient, s.tmpl, s.data, options)
}

// text
func (s *Shell) Sed(file ...string) *Shell {
	s.tmpl = template.Must(template.New("sed").Parse(TEMPLATE_SED))
	s.data["files"] = strings.Join(file, " ")
	return s
}

// storage(filesystem/block)
func (s *Shell) List(files ...string) *Shell {
	s.tmpl = template.Must(template.New("list").Parse(TEMPLATE_LIST))
	s.data["files"] = strings.Join(files, " ")
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

func (s *Shell) Remove(file ...string) *Shell {
	s.tmpl = template.Must(template.New("remove").Parse(TEMPLATE_REMOVE))
	s.data["files"] = strings.Join(file, " ")
	return s
}

func (s *Shell) Rename(source, dest string) *Shell {
	s.tmpl = template.Must(template.New("rename").Parse(TEMPLATE_RENAME))
	s.data["source"] = source
	s.data["dest"] = dest
	return s
}

func (s *Shell) Copy(source, dest string) *Shell {
	s.tmpl = template.Must(template.New("copy").Parse(TEMPLATE_COPY))
	s.data["source"] = source
	s.data["dest"] = dest
	return s
}

func (s *Shell) Chmod(mode, file string) *Shell {
	s.tmpl = template.Must(template.New("chmod").Parse(TEMPLATE_CHMOD))
	s.data["mode"] = mode
	s.data["file"] = file
	return s
}

func (s *Shell) Stat(files ...string) *Shell {
	s.tmpl = template.Must(template.New("stat").Parse(TEMPLATE_STAT))
	s.data["files"] = strings.Join(files, " ")
	return s
}

func (s *Shell) Cat(files ...string) *Shell {
	s.tmpl = template.Must(template.New("cat").Parse(TEMPLATE_CAT))
	s.data["files"] = strings.Join(files, " ")
	return s
}

func (s *Shell) Mkfs(device string) *Shell {
	s.tmpl = template.Must(template.New("mkfs").Parse(TEMPLATE_MKFS))
	s.data["device"] = device
	return s
}

func (s *Shell) Mount(source, directory string) *Shell {
	s.tmpl = template.Must(template.New("mount").Parse(TEMPLATE_MOUNT))
	s.data["source"] = source
	s.data["directory"] = directory
	return s
}

func (s *Shell) Umount(directory string) *Shell {
	s.tmpl = template.Must(template.New("umount").Parse(TEMPLATE_UMOUNT))
	s.data["directory"] = directory
	return s
}

func (s *Shell) Tune2FS(device string) *Shell {
	s.tmpl = template.Must(template.New("tune2fs").Parse(TEMPLATE_TUNE2FS))
	s.data["device"] = device
	return s
}

func (s *Shell) Fuser(name ...string) *Shell {
	s.tmpl = template.Must(template.New("fuser").Parse(TEMPLATE_FUSER))
	s.data["names"] = strings.Join(name, " ")
	return s
}

func (s *Shell) DiskFree(file ...string) *Shell {
	s.tmpl = template.Must(template.New("diskfree").Parse(TEMPLATE_DISKFREE))
	s.data["files"] = strings.Join(file, " ")
	return s
}

func (s *Shell) LsBlk(device ...string) *Shell {
	s.tmpl = template.Must(template.New("lsblk").Parse(TEMPLATE_LSBLK))
	s.data["devices"] = strings.Join(device, " ")
	return s
}

func (s *Shell) BlkId(device string) *Shell {
	s.tmpl = template.Must(template.New("blkid").Parse(TEMPLATE_BLKID))
	s.data["device"] = device
	return s
}

// network
func (s *Shell) SocketStatistics(filter string) *Shell {
	s.tmpl = template.Must(template.New("ss").Parse(TEMPLATE_SS))
	s.data["filter"] = filter
	return s
}

func (s *Shell) Ping(destination string) *Shell {
	s.tmpl = template.Must(template.New("ping").Parse(TEMPLATE_PING))
	s.data["destination"] = destination
	return s
}

func (s *Shell) Curl(url string) *Shell {
	s.tmpl = template.Must(template.New("curl").Parse(TEMPLATE_CURL))
	s.data["url"] = url
	return s
}

// kernel
func (s *Shell) Whoami() *Shell {
	s.tmpl = template.Must(template.New("whoami").Parse(TEMPLATE_WHOAMI))
	return s
}

func (s *Shell) Date(format string) *Shell {
	s.tmpl = template.Must(template.New("date").Parse(TEMPLATE_DATE))
	s.data["format"] = format
	return s
}

func (s *Shell) UnixName() *Shell {
	s.tmpl = template.Must(template.New("uname").Parse(TEMPLATE_UNAME))
	return s
}

func (s *Shell) ModProbe(modulename string, args ...string) *Shell {
	s.tmpl = template.Must(template.New("modprobe").Parse(TEMPLATE_MODPROBE))
	s.data["modulename"] = modulename
	s.data["arguments"] = strings.Join(args, " ")
	return s
}

func (s *Shell) ModInfo(modulename string) *Shell {
	s.tmpl = template.Must(template.New("modinfo").Parse(TEMPLATE_MODINFO))
	s.data["modulename"] = modulename
	return s
}

// other
func (s *Shell) Tar(file string) *Shell {
	s.tmpl = template.Must(template.New("tar").Parse(TEMPLATE_TAR))
	s.data["file"] = file
	return s
}

func (s *Shell) Dpkg() *Shell {
	s.tmpl = template.Must(template.New("tar").Parse(TEMPLATE_DPKG))
	return s
}

func (s *Shell) Rpm() *Shell {
	s.tmpl = template.Must(template.New("tar").Parse(TEMPLATE_RPM))
	return s
}

func (s *Shell) Scp(source, user, host, target string) *Shell {
	s.tmpl = template.Must(template.New("scp").Parse(TEMPLATE_SCP))
	s.data["source"] = source
	s.data["user"] = user
	s.data["host"] = host
	s.data["target"] = target
	return s
}

func (s *Shell) Command(command string) *Shell {
	s.tmpl = template.Must(template.New("command").Parse(TEMPLATE_COMMAND))
	s.data["command"] = command
	return s
}

func (s *Shell) BashScript(scriptPath string, args ...string) *Shell {
	s.tmpl = template.Must(template.New("bashScript").Parse(TEMPLATE_BASH_SCEIPT))
	s.data["scriptPath"] = scriptPath
	s.data["arguments"] = strings.Join(args, " ")
	return s
}

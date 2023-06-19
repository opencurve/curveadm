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

// __SIGN_BY_WINE93__

package step

import (
	"fmt"
	"os"
	"strings"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/module"
)

const (
	ERR_NOT_MOUNTED          = "not mounted"
	ERR_MOUNTPOINT_NOT_FOUND = "mountpoint not found"
	ERROR_DEVICE_BUSY        = "Device or resource busy"
)

type (
	// text
	Sed struct {
		Files      []string
		Expression *string
		InPlace    bool
		Out        *string
		module.ExecOptions
	}

	// storage (filesystem/block)
	List struct {
		Files []string
		Out   *string
		module.ExecOptions
	}

	CreateDirectory struct {
		Paths   []string
		Success *bool
		Out     *string
		module.ExecOptions
	}

	RemoveFile struct {
		Files []string
		Out   *string
		module.ExecOptions
	}

	CopyFile struct {
		Source    string
		Dest      string
		NoClobber bool // do not overwrite an existing file
		Out       *string
		module.ExecOptions
	}

	Stat struct {
		Files  []string
		Format string
		Out    *string
		module.ExecOptions
	}

	Cat struct {
		Files   []string
		Success *bool
		Out     *string
		module.ExecOptions
	}

	CreateFilesystem struct {
		Device string
		Out    *string
		module.ExecOptions
	}

	MountFilesystem struct {
		Source    string
		Directory string
		Out       *string
		module.ExecOptions
	}

	UmountFilesystem struct {
		Directorys     []string
		IgnoreUmounted bool
		IgnoreNotFound bool
		Out            *string
		module.ExecOptions
	}

	Tune2FS struct {
		Device                   string
		ReservedBlocksPercentage string
		Success                  *bool
		Out                      *string
		module.ExecOptions
	}

	Fuser struct {
		Names []string
		Out   *string
		module.ExecOptions
	}

	// see also: https://linuxize.com/post/how-to-check-disk-space-in-linux-using-the-df-command/#output-format
	ShowDiskFree struct {
		Files  []string
		Format string
		Out    *string
		module.ExecOptions
	}

	ListBlockDevice struct {
		Device     []string
		Format     string
		NoHeadings bool
		Success    *bool
		Out        *string
		module.ExecOptions
	}

	BlockId struct {
		Device   string
		Format   string
		MatchTag string
		Success  *bool
		Out      *string
		module.ExecOptions
	}

	// network
	SocketStatistics struct {
		Filter    string
		Listening bool // display listening sockets
		NoHeader  bool // Suppress header line
		Success   *bool
		Out       *string
		module.ExecOptions
	}

	Ping struct {
		Destination *string
		Count       int
		Timeout     int
		Success     *bool
		Out         *string
		module.ExecOptions
	}

	Curl struct {
		Url      string
		Form     string
		Insecure bool
		Output   string
		Silent   bool
		Success  *bool
		Out      *string
		module.ExecOptions
	}

	// kernel
	Whoami struct {
		Success *bool
		Out     *string
		module.ExecOptions
	}

	Date struct {
		Format string
		Out    *string
		module.ExecOptions
	}

	UnixName struct {
		KernelRelease bool
		Out           *string
		module.ExecOptions
	}

	ModInfo struct {
		Name    string
		Success *bool
		Out     *string
		module.ExecOptions
	}

	ModProbe struct {
		Name    string
		Args    []string
		Success *bool
		Out     *string
		module.ExecOptions
	}

	// other
	Hostname struct {
		Success *bool
		Out     *string
		module.ExecOptions
	}

	Tar struct {
		File            string
		Archive         string // -f
		Create          bool   // -c
		Directory       string // -C
		Extract         bool   // -x
		StripComponents int
		Gzip            bool // -z
		UnGzip          bool // -z
		Verbose         bool // -v
		Success         *bool
		Out             *string
		module.ExecOptions
	}

	Dpkg struct {
		Install string
		Purge   string
		Success *bool
		Out     *string
		module.ExecOptions
	}

	Rpm struct {
		Hash    bool
		Install string
		NoDeps  bool
		Verbose bool
		Success *bool
		Out     *string
		module.ExecOptions
	}

	Scp struct {
		Content    *string
		Mode       int
		RemotePath string
		module.ExecOptions
	}

	Command struct {
		Command string
		Success *bool
		Out     *string
		module.ExecOptions
	}
)

func (s *Sed) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Sed(s.Files...)
	if s.InPlace {
		cmd.AddOption("--in-place")
	}
	if len(*s.Expression) > 0 {
		cmd.AddOption("--expression='%s'", *s.Expression)
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_EDIT_FILE_FAILED)
}

func (s *List) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().List(s.Files...)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_EDIT_FILE_FAILED)
}

func (s *CreateDirectory) Execute(ctx *context.Context) error {
	for _, path := range s.Paths {
		if len(path) == 0 {
			continue
		}

		cmd := ctx.Module().Shell().Mkdir(path)
		cmd.AddOption("--parents") // no error if existing, make parent directories as needed

		out, err := cmd.Execute(s.ExecOptions)
		err = PostHandle(s.Success, s.Out, out, err, errno.ERR_CREATE_DIRECTORY_FAILED)
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

		out, err := cmd.Execute(s.ExecOptions)
		err = PostHandle(nil, s.Out, out, err, errno.ERR_REMOVE_FILES_OR_DIRECTORIES_FAILED)
		// FIXME
		// device busy: maybe directory is mount point
		if err != nil && !strings.Contains(out, ERROR_DEVICE_BUSY) {
			return err
		}
	}
	return nil
}

func (s *CopyFile) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Copy(s.Source, s.Dest)
	if s.NoClobber {
		cmd.AddOption("--no-clobber")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_COPY_FILES_AND_DIRECTORIES_FAILED)
}

func (s *Stat) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Stat(s.Files...)
	if len(s.Format) > 0 {
		cmd.AddOption(fmt.Sprintf("--format='%s'", s.Format))
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_GET_FILE_STATUS_FAILED)
}

func (s *Cat) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Cat(s.Files...)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_CONCATENATE_FILE_FAILED)
}

func (s *CreateFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mkfs(s.Device)
	// force mke2fs to create a filesystem, even if the specified device is not a partition
	// on a block special device, or if other parameters do not make sense
	cmd.AddOption("-F")

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_BUILD_A_LINUX_FILE_SYSTEM_FAILED)
}

func (s *MountFilesystem) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Mount(s.Source, s.Directory)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_MOUNT_A_FILESYSTEM_FAILED)
}

func (s *UmountFilesystem) Execute(ctx *context.Context) error {
	for _, directory := range s.Directorys {
		if len(directory) == 0 {
			continue
		}
		cmd := ctx.Module().Shell().Umount(directory)

		out, err := cmd.Execute(s.ExecOptions)
		err = PostHandle(nil, s.Out, out, err, errno.ERR_UNMOUNT_FILE_SYSTEMS_FAILED)

		if (s.IgnoreUmounted && strings.Contains(out, ERR_NOT_MOUNTED)) ||
			(s.IgnoreNotFound && strings.Contains(out, ERR_MOUNTPOINT_NOT_FOUND)) {
			continue
		} else if err != nil {
			return err
		}
	}
	return nil
}

func (s *Tune2FS) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Tune2FS(s.Device)
	if len(s.ReservedBlocksPercentage) > 0 {
		cmd.AddOption("-m %s", s.ReservedBlocksPercentage)
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_RESERVE_FILESYSTEM_BLOCKS_FAILED)
}

func (s *Fuser) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Fuser(s.Names...)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_FIND_WHICH_PROCESS_USING_FILE_FAILED)
}

func (s *ShowDiskFree) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().DiskFree(s.Files...)
	if len(s.Format) > 0 {
		cmd.AddOption("--output=%s", s.Format)
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_GET_DISK_SPACE_USAGE_FAILED)
}

func (s *ListBlockDevice) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().LsBlk(s.Device...)
	if len(s.Format) > 0 {
		cmd.AddOption("--output=%s", s.Format)
	}
	if s.NoHeadings {
		cmd.AddOption("--noheadings")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_LIST_BLOCK_DEVICES_FAILED)
}

func (s *BlockId) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().BlkId(s.Device)
	if len(s.Format) > 0 {
		cmd.AddOption("--output=%s", s.Format)
	}
	if len(s.MatchTag) > 0 {
		cmd.AddOption("--match-tag=%s", s.MatchTag)
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_BLOCK_DEVICE_UUID_FAILED)
}

// network
func (s *SocketStatistics) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().SocketStatistics(s.Filter)
	if s.Listening {
		cmd.AddOption("--listening")
	}
	if s.NoHeader {
		cmd.AddOption("--no-header")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_CONNECTION_INFORMATION_FAILED)
}

func (s *Ping) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Ping(*s.Destination)
	if s.Count > 0 {
		cmd.AddOption("-c %d", s.Count)
	}
	if s.Timeout > 0 {
		cmd.AddOption("-W %d", s.Timeout)
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_SEND_ICMP_ECHO_REQUEST_TO_HOST_FAILED)
}

func (s *Curl) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Curl(s.Url)
	if len(s.Form) > 0 {
		cmd.AddOption("--form %s", s.Form)
	}
	if s.Insecure {
		cmd.AddOption("--insecure")
	}
	if len(s.Output) > 0 {
		cmd.AddOption("--output %s", s.Output)
	}
	if s.Silent {
		cmd.AddOption("--silent")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_TRANSFERRING_DATA_FROM_OR_TO_SERVER_FAILED)
}

// kernel
func (s *Whoami) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Whoami()
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_SYSTEM_TIME_FAILED)
}

func (s *Date) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Date(s.Format)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_GET_SYSTEM_TIME_FAILED)
}

func (s *UnixName) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().UnixName()
	if s.KernelRelease {
		cmd.AddOption("--kernel-release")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(nil, s.Out, out, err, errno.ERR_GET_SYSTEM_INFORMATION_FAILED)
}

func (s *ModInfo) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().ModInfo(s.Name)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_KERNEL_MODULE_INFO_FAILED)
}

func (s *ModProbe) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().ModProbe(s.Name, s.Args...)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_ADD_MODUDLE_FROM_LINUX_KERNEL_FAILED)
}

// other
func (s *Hostname) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Command("hostname")
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_GET_HOSTNAME_FAILED)
}

func (s *Tar) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Tar(s.File)
	if len(s.Archive) > 0 {
		cmd.AddOption("--file=%s", s.Archive)
	}
	if s.Create {
		cmd.AddOption("--create")
	}
	if len(s.Directory) > 0 {
		cmd.AddOption("--directory=%s", s.Directory)
	}
	if s.Extract {
		cmd.AddOption("--extract")
	}
	if s.StripComponents > 0 {
		cmd.AddOption("--strip-components=%d", s.StripComponents)
	}
	if s.UnGzip {
		cmd.AddOption("--ungzip")
	}
	if s.Gzip {
		cmd.AddOption("--gzip")
	}
	if s.Verbose {
		cmd.AddOption("--verbose")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_STORES_AND_EXTRACTS_FILES_FAILED)
}

func (s *Dpkg) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Dpkg()
	if len(s.Install) > 0 {
		cmd.AddOption("--install %s", s.Install)
	}
	if len(s.Purge) > 0 {
		cmd.AddOption("--purge %s", s.Purge)
	}
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_INSTALL_OR_REMOVE_DEBIAN_PACKAGE_FAILED)
}

func (s *Rpm) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Rpm()
	if s.Hash {
		cmd.AddOption("--hash")
	}
	if len(s.Install) > 0 {
		cmd.AddOption("--install %s", s.Install)
	}
	if s.NoDeps {
		cmd.AddOption("--nodeps")
	}
	if s.Verbose {
		cmd.AddOption("--verbose")
	}

	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_INSTALL_OR_REMOVE_RPM_PACKAGE_FAILED)
}

func (s *Scp) Execute(ctx *context.Context) error {
	localPath := utils.RandFilename(TEMP_DIR)
	defer os.Remove(localPath)
	mode := 0644
	if s.Mode > 0 {
		mode = s.Mode
	}
	err := utils.WriteFile(localPath, *s.Content, mode)
	if err != nil {
		return errno.ERR_WRITE_FILE_FAILED.E(err)
	}

	config := ctx.SSHClient().Config()
	cmd := ctx.Module().Shell().Scp(localPath, config.User, config.Host, s.RemotePath)
	cmd.AddOption("-P %d", config.Port)
	if !config.ForwardAgent {
		cmd.AddOption("-i %s", config.PrivateKeyPath)
	}

	options := s.ExecOptions
	options.ExecWithSudo = false
	options.ExecInLocal = true
	out, err := cmd.Execute(options)
	return PostHandle(nil, nil, out, err, errno.ERR_SECURE_COPY_FILE_TO_REMOTE_FAILED)
}

func (s *Command) Execute(ctx *context.Context) error {
	cmd := ctx.Module().Shell().Command(s.Command)
	out, err := cmd.Execute(s.ExecOptions)
	return PostHandle(s.Success, s.Out, out, err, errno.ERR_RUN_A_BASH_COMMAND_FAILED)
}

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
 * Created Date: 2022-08-08
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"path"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

/*
 * pfsd-v1.2.tar.gz"
 * ├── pfsd.deb
 * └── pfsd.rpm
 */
const (
	URL_POLARFS_PACKAGE = "https://curveadm.nos-eastchina1.126.net/package/pfsd-v1.2.tar.gz"

	PACKAGE_ITEM_NAME_PFSD_DEB = "pfsd.deb"
	PACKAGE_ITEM_NAME_PFSD_RPM = "pfsd.rpm"
)

type step2InstallPackage struct {
	root     string
	release  string
	curveadm *cli.CurveAdm
}

func (s *step2InstallPackage) Execute(ctx *context.Context) error {
	steps := []task.Step{}
	curveadm := s.curveadm
	release := s.release
	if release == comm.OS_RELEASE_DEBIAN ||
		release == comm.OS_RELEASE_UBUNTU {
		steps = append(steps, &step.Dpkg{
			Install:     path.Join(s.root, PACKAGE_ITEM_NAME_PFSD_DEB),
			ExecOptions: curveadm.ExecOptions(),
		})
	} else if release == comm.OS_RELEASE_CENTOS {
		steps = append(steps, &step.Rpm{
			Hash:        true,
			Install:     path.Join(s.root, PACKAGE_ITEM_NAME_PFSD_RPM),
			NoDeps:      true,
			Verbose:     true,
			ExecOptions: curveadm.ExecOptions(),
		})
	} else {
		return errno.ERR_UNSUPPORT_LINUX_OS_REELASE.
			F("os release: %s", release)
	}

	for _, step := range steps {
		err := step.Execute(ctx)
		if err != nil {
			return errno.ERR_INSTALL_PFSD_PACKAGE_FAILED.E(err)
		}
	}
	return nil
}

func getRelease(curveadm *cli.CurveAdm) string {
	v := curveadm.MemStorage().Get(comm.KEY_OS_RELEASE)
	if v == nil {
		return comm.OS_RELEASE_UNKNOWN
	}
	return v.(string)
}

func NewInstallPolarFSTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	host := curveadm.MemStorage().Get(comm.KEY_POLARFS_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	release := getRelease(curveadm)
	subname := fmt.Sprintf("host=%s release=%s", host, release)
	t := task.NewTask("Install PolarFS", subname, hc.GetSSHConfig())

	// add step to task
	var input, output string
	randStr := utils.RandString(10)
	tarball := fmt.Sprintf("/tmp/pfsd-%s.tar.gz", randStr)
	root := fmt.Sprintf("/tmp/pfsd-%s", randStr)

	t.AddStep(&step.CreateDirectory{
		Paths:       []string{root},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Curl{
		Url:         URL_POLARFS_PACKAGE,
		Insecure:    true,
		Output:      tarball,
		Silent:      true,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Tar{
		Archive:         tarball,
		Directory:       root,
		Extract:         true,
		StripComponents: 1,
		UnGzip:          true,
		Verbose:         true,
		ExecOptions:     curveadm.ExecOptions(),
	})
	t.AddStep(&step2InstallPackage{
		root:     root,
		release:  release,
		curveadm: curveadm,
	})
	t.AddStep(&step.ReadFile{
		HostSrcPath: "/etc/curve/conf/client.conf.template",
		Content:     &input,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Filter{
		KVFieldSplit: CLIENT_CONFIG_DELIMITER,
		Mutate:       newMutate(cc, CLIENT_CONFIG_DELIMITER),
		Input:        &input,
		Output:       &output,
	})
	t.AddStep(&step.InstallFile{
		Content:      &output,
		HostDestPath: "/etc/curve/client.conf",
		ExecOptions:  curveadm.ExecOptions(),
	})
	t.AddPostStep(&step.RemoveFile{
		Files: []string{tarball, root},
	})

	return t, nil
}

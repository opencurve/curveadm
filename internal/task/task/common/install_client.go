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

package common

import (
	"fmt"
	"path"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

/*
 * curvebs-v1.2.tar.gz"
 * ├── curvebs.deb
 * └── curvebs.rpm
 *
 * curvefs-v2.3.tar.gz"
 * ├── curvefs.deb
 * └── curvefs.rpm
 */
const (
	URL_CURVEBS_PACKAGE = "https://curveadm.nos-eastchina1.126.net/package/curvebs-v1.2.tar.gz"
	URL_CURVEFS_PACKAGE = "https://curveadm.nos-eastchina1.126.net/package/curvefs-v2.3.tar.gz"

	PACKAGE_ITEM_NAME_CLIENT_DEB = "client.deb"
	PACKAGE_ITEM_NAME_CLIENT_RPM = "client.rpm"

	KIND_CURVEBS = topology.KIND_CURVEBS

	CLIENT_CONFIG_DELIMITER = "="
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
			Install:     path.Join(s.root, PACKAGE_ITEM_NAME_CLIENT_DEB),
			ExecOptions: curveadm.ExecOptions(),
		})
	} else if release == comm.OS_RELEASE_CENTOS {
		steps = append(steps, &step.Rpm{
			Hash:        true,
			Install:     path.Join(s.root, PACKAGE_ITEM_NAME_CLIENT_RPM),
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

func newClientMutate(cc *configure.ClientConfig, delimiter string) step.Mutate {
	serviceConfig := cc.GetServiceConfig()
	return func(in, key, value string) (out string, err error) {
		if len(key) == 0 {
			out = in
			return
		}

		// replace config
		v, ok := serviceConfig[strings.ToLower(key)]
		if ok {
			value = v
		}

		// replace variable
		value, err = cc.GetVariables().Rendering(value)
		if err != nil {
			return
		}

		out = fmt.Sprintf("%s%s%s", key, delimiter, value)
		return
	}
}

func getRelease(curveadm *cli.CurveAdm) string {
	v := curveadm.MemStorage().Get(comm.KEY_OS_RELEASE)
	if v == nil {
		return comm.OS_RELEASE_UNKNOWN
	}
	return v.(string)
}

func NewInstallClientTask(curveadm *cli.CurveAdm, cc *configure.ClientConfig) (*task.Task, error) {
	host := curveadm.MemStorage().Get(comm.KEY_CLIENT_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	kind := cc.GetKind()
	release := getRelease(curveadm)
	subname := fmt.Sprintf("host=%s release=%s", host, release)
	name := utils.Choose(kind == KIND_CURVEBS, "CurveBS", "CurveFS")
	t := task.NewTask(fmt.Sprintf("Install %s Client", name), subname, hc.GetConnectConfig())

	// add step to task
	var input, output string
	randStr := utils.RandString(10)
	tarball := fmt.Sprintf("/tmp/pfsd-%s.tar.gz", randStr)
	root := fmt.Sprintf("/tmp/pfsd-%s", randStr)
	url := utils.Choose(kind == KIND_CURVEBS, URL_CURVEBS_PACKAGE, URL_CURVEFS_PACKAGE)
	name = utils.Choose(kind == KIND_CURVEBS, "curve", "curvefs")

	t.AddStep(&step.CreateDirectory{
		Paths:       []string{root},
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Curl{
		Url:         url,
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
		HostSrcPath: fmt.Sprintf("/etc/%s/conf/client.conf.template", name),
		Content:     &input,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step.Filter{
		KVFieldSplit: CLIENT_CONFIG_DELIMITER,
		Mutate:       newClientMutate(cc, CLIENT_CONFIG_DELIMITER),
		Input:        &input,
		Output:       &output,
	})
	t.AddStep(&step.InstallFile{
		Content:      &output,
		HostDestPath: fmt.Sprintf("/etc/%s/client.conf", name),
		ExecOptions:  curveadm.ExecOptions(),
	})
	t.AddPostStep(&step.RemoveFile{
		Files: []string{tarball, root},
	})

	return t, nil
}

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
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/os"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	REGEX_OS_RELEASE = "^ID=(.*)$"
)

type step2ParseOSRelease struct {
	host     string
	success  *bool
	out      *string
	curveadm *cli.CurveAdm
}

func (s *step2ParseOSRelease) Execute(ctx *context.Context) error {
	if !*s.success {
		return errno.ERR_CONCATENATE_FILE_FAILED.S(*s.out)
	}

	lines := strings.Split(*s.out, "\n")
	pattern := regexp.MustCompile(REGEX_OS_RELEASE)
	for _, line := range lines {
		mu := pattern.FindStringSubmatch(line)
		if len(mu) != 0 {
			s.curveadm.MemStorage().Set(comm.KEY_OS_RELEASE, mu[1])
			return nil
		}
	}
	return errno.ERR_GET_OS_REELASE_FAILED.S(*s.out)
}

func NewDetectOSReleaseTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	host := curveadm.MemStorage().Get(comm.KEY_POLARFS_HOST).(string)
	hc, err := curveadm.GetHost(host)
	if err != nil {
		return nil, err
	}

	// new task
	var success bool
	var out string
	subname := fmt.Sprintf("host=%s", host)
	t := task.NewTask("Detect OS Release", subname, hc.GetConnectConfig())

	// add step to task
	t.AddStep(&step.Cat{
		Files:       []string{os.GetOSReleasePath()},
		Success:     &success,
		Out:         &out,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2ParseOSRelease{
		host:     host,
		success:  &success,
		out:      &out,
		curveadm: curveadm,
	})

	return t, nil
}

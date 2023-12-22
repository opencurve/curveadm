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
 * Created Date: 2022-02-09
 * Author: Jingli Chen (Wine93)
 */

package bs

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
	"github.com/opencurve/curveadm/internal/utils"
)

const (
	DEFAULT_TGTD_LISTEN_PORT = 3260
)

type (
	step2FormatTarget struct {
		host       string
		hostname   string
		output     *string
		memStorage *utils.SafeMap
	}

	Target struct {
		Host   string
		Tid    string
		Name   string
		Store  string
		Portal string
	}
)

func addTarget(memStorage *utils.SafeMap, id string, target *Target) {
	memStorage.TX(func(kv *utils.SafeMap) error {
		m := map[string]*Target{}
		v := kv.Get(comm.KEY_ALL_TARGETS)
		if v != nil {
			m = v.(map[string]*Target)
		}
		m[id] = target
		kv.Set(comm.KEY_ALL_TARGETS, m)
		return nil
	})
}

/*
Output Example:
Target 3: iqn.2022-02.com.opencurve:curve.wine93/test03

	...
	LUN information:
	    LUN: 0
	        ...
	    LUN: 1
	        ...
	        Backing store path: cbd:pool//test03_wine93_
*/
func (s *step2FormatTarget) Execute(ctx *context.Context) error {
	output := *s.output
	lines := strings.Split(output, "\n")

	var target *Target
	titlePattern := regexp.MustCompile("^Target ([0-9]+): (.+)$")
	storePattern := regexp.MustCompile("Backing store path: (cbd:pool//.+)$")
	for _, line := range lines {
		mu := titlePattern.FindStringSubmatch(line)
		if len(mu) > 0 {
			target = &Target{
				Host:   s.host,
				Tid:    mu[1],
				Name:   mu[2],
				Store:  "-",
				Portal: fmt.Sprintf("%s:%d", s.hostname, DEFAULT_TGTD_LISTEN_PORT),
			}
			addTarget(s.memStorage, mu[1], target)
			continue
		}

		mu = storePattern.FindStringSubmatch(line)
		if len(mu) > 0 {
			target.Store = mu[1]
		}
	}

	return nil
}

func NewListTargetsTask(curveadm *cli.CurveAdm, v interface{}) (*task.Task, error) {
	options := curveadm.MemStorage().Get(comm.KEY_TARGET_OPTIONS).(TargetOption)
	hc, err := curveadm.GetHost(options.Host)
	if err != nil {
		return nil, err
	}

	subname := fmt.Sprintf("host=%s", hc.GetHostname())
	t := task.NewTask("List Targets", subname, hc.GetConnectConfig())

	// add step
	var output string
	containerId := DEFAULT_TGTD_CONTAINER_NAME

	t.AddStep(&step.ListContainers{
		ShowAll:     true,
		Format:      "'{{.ID}} {{.Status}}'",
		Filter:      fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2CheckTgtdStatus{
		output: &output,
	})
	t.AddStep(&step.ContainerExec{
		ContainerId: &containerId,
		Command:     fmt.Sprintf("tgtadm --lld iscsi --mode target --op show"),
		Out:         &output,
		ExecOptions: curveadm.ExecOptions(),
	})
	t.AddStep(&step2FormatTarget{
		host:       options.Host,
		hostname:   hc.GetHostname(),
		output:     &output,
		memStorage: curveadm.MemStorage(),
	})

	return t, nil
}

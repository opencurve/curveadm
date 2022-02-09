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
	"github.com/opencurve/curveadm/internal/utils"
	"regexp"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	client "github.com/opencurve/curveadm/internal/configure/client/bs"
	"github.com/opencurve/curveadm/internal/task/context"
	"github.com/opencurve/curveadm/internal/task/step"
	"github.com/opencurve/curveadm/internal/task/task"
)

const (
	DEFAULT_TGTD_LISTEN_PORT = 3260
)

type (
	step2FormatTarget struct {
		output     *string
		config     *client.ClientConfig
		memStorage *utils.SafeMap
	}

	Target struct {
		Tid    string
		Name   string
		Store  string
		Portal string
	}
)

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
				Tid:    mu[1],
				Name:   mu[2],
				Store:  "-",
				Portal: fmt.Sprintf("%s:%d", s.config.GetHost(), DEFAULT_TGTD_LISTEN_PORT),
			}
			s.memStorage.Set(mu[1], target)
			continue
		}

		mu = storePattern.FindStringSubmatch(line)
		if len(mu) > 0 {
			target.Store = mu[1]
		}
	}

	return nil
}

func NewListTargetsTask(curveadm *cli.CurveAdm, cc *client.ClientConfig) (*task.Task, error) {
	subname := fmt.Sprintf("hostname=%s", cc.GetHost())
	t := task.NewTask("List Targets", subname, cc.GetSSHConfig())

	// add step
	var output string
	containerId := DEFAULT_TGTD_CONTAINER_NAME

	t.AddStep(&step.ListContainers{
		ShowAll:       true,
		Format:        "'{{.ID}} {{.Status}}'",
		Quiet:         true,
		Filter:        fmt.Sprintf("name=%s", DEFAULT_TGTD_CONTAINER_NAME),
		Out:           &output,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2CheckTgtdStatus{
		output: &output,
	})
	t.AddStep(&step.ContainerExec{
		ContainerId:   &containerId,
		Command:       fmt.Sprintf("tgtadm --lld iscsi --mode target --op show"),
		Out:           &output,
		ExecWithSudo:  true,
		ExecInLocal:   false,
		ExecSudoAlias: curveadm.SudoAlias(),
	})
	t.AddStep(&step2FormatTarget{
		output:     &output,
		config:     cc,
		memStorage: curveadm.MemStorage(),
	})

	return t, nil
}

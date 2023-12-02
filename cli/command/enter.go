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
 * Created Date: 2021-10-15
 * Author: Jingli Chen (Wine93)
 */

// __SIGN_BY_WINE93__

package command

import (
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	ATTACH_LEADER_OR_RANDOM_CONTAINER = []int{playbook.ATTACH_LEADER_OR_RANDOM_CONTAINER}
)

type enterOptions struct {
	id string
}

func NewEnterCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options enterOptions

	cmd := &cobra.Command{
		Use:   "enter ID",
		Short: "Enter service container",
		Args:  utils.RequiresMaxArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return nil
			}
			options.id = args[0]
			return curveadm.CheckId(options.id)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runEnter(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func genLeaderOrRandomPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig) (*playbook.Playbook, error) {
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := ATTACH_LEADER_OR_RANDOM_CONTAINER
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar:  true,
				SilentMainBar: true,
				SkipError:     true,
			},
		})
	}
	return pb, nil
}

func checkOrGetId(curveadm *cli.CurveAdm, dcs []*topology.DeployConfig, options enterOptions) (string, error) {
	id := options.id
	if id != "" {
		return id, nil
	}
	pb, err := genLeaderOrRandomPlaybook(curveadm, dcs)
	if err != nil {
		return "", err
	}
	// run playground
	err = pb.Run()
	if err != nil {
		return "", err
	}
	// get leader or random container id
	value := curveadm.MemStorage().Get(comm.LEADER_OR_RANDOM_ID)
	if value == nil {
		return "", errno.ERR_NO_LEADER_OR_RANDOM_CONTAINER_FOUND
	}
	id = value.(common.Leader0rRandom).Id
	return id, nil
}

func runEnter(curveadm *cli.CurveAdm, options enterOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

	// 2) check id options
	id, err := checkOrGetId(curveadm, dcs, options)
	if err != nil {
		return err
	}

	// 3) filter service
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   id,
		Role: "*",
		Host: "*",
	})
	if len(dcs) == 0 {
		return errno.ERR_NO_SERVICES_MATCHED
	}

	// 4) get container id
	dc := dcs[0]
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return err
	}

	// 5) attach remote container
	home := dc.GetProjectLayout().ServiceRootDir
	return tools.AttachRemoteContainer(curveadm, dc.GetHost(), containerId, home)
}

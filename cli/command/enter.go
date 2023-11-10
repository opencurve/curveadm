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
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
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

func genStatusForLeaderPlaybook(curveadm *cli.CurveAdm,
	dcs []*topology.DeployConfig,
	options statusOptions) (*playbook.Playbook, error) {
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})
	if len(dcs) == 0 {
		return nil, errno.ERR_NO_SERVICES_MATCHED
	}

	steps := []int{playbook.INIT_SERVIE_STATUS, playbook.GET_MDS_LEADER}
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: dcs,
			ExecOptions: playbook.ExecOptions{
				//Concurrency:   10,
				SilentSubBar:  true,
				SilentMainBar: step == playbook.INIT_SERVIE_STATUS,
				SkipError:     true,
			},
		})
	}
	return pb, nil
}

func runEnter(curveadm *cli.CurveAdm, options enterOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}
	var containerId string
	var dc *topology.DeployConfig
	if options.id != "" {
		// If there is an ID, execute the following
		// 2) filter service
		dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
			Id:   options.id,
			Role: "*",
			Host: "*",
		})
		if len(dcs) == 0 {
			return errno.ERR_NO_SERVICES_MATCHED
		}
		// 3) get container id
		dc = dcs[0]
	} else {
		// If no id parameter, execute the following
		// 2) generate get status playbook
		statusForLeaderOptions := statusOptions{id: "*", role: ROLE_MDS, host: "*"}
		pb, err := genStatusForLeaderPlaybook(curveadm, dcs, statusForLeaderOptions)
		if err != nil {
			return err
		}
		// 3) run playground
		err = pb.Run()
		if err != nil {
			return err
		}
		// 4) display service status for leader
		statuses := []task.LeaderServiceStatus{}
		value := curveadm.MemStorage().Get(comm.KEY_LEADER_SERVICE_STATUS)
		if value != nil {
			m := value.(map[string]task.LeaderServiceStatus)
			for _, status := range m {
				statuses = append(statuses, status)
			}
		}
		for _, status := range statuses {
			if status.IsLeader {
				dc = status.Config
				break
			}
		}
		// 5) get leader container id
		if dc == nil {
			return errno.ERR_NO_LEADER_CONTAINER_FOUND
		}
	}
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err = curveadm.GetContainerId(serviceId)
	if err != nil {
		return err
	}

	// attach remote container
	home := dc.GetProjectLayout().ServiceRootDir
	return tools.AttachRemoteContainer(curveadm, dc.GetHost(), containerId, home)
}

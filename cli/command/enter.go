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
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type enterOptions struct {
	id            string
	role          string
	host          string
	verbose       bool
	showInstances bool
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

func runEnter(curveadm *cli.CurveAdm, options enterOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}
	statusOptions1 := statusOptions{id: "*", role: "*", host: "*"}
	pb, err := genStatusPlaybook(curveadm, dcs, statusOptions1)
	if err != nil {
		return err
	}
	// 3) run playground
	err = pb.Run()

	var containerId string
	var dc *topology.DeployConfig
	//如果有ID执行如下

	if options.id != "" {
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
		serviceId := curveadm.GetServiceId(dc.GetId())
		containerId, err = curveadm.GetContainerId(serviceId)
		if err != nil {
			return err
		}
	} else {
		statuses := []task.ServiceStatus{}
		value := curveadm.MemStorage().Get(comm.KEY_ALL_SERVICE_STATUS)
		if value != nil {
			m := value.(map[string]task.ServiceStatus)
			for _, status := range m {
				statuses = append(statuses, status)
			}
		}
		for _, status := range statuses {
			if !status.IsLeader {
				continue
			}
			dc = status.Config
		}
		serviceId := curveadm.GetServiceId(dc.GetId())
		containerId, err = curveadm.GetContainerId(serviceId)
		if err != nil {
			return err
		}
	}

	// 4) attch remote container
	home := dc.GetProjectLayout().ServiceRootDir
	return tools.AttachRemoteContainer(curveadm, dc.GetHost(), containerId, home)
}

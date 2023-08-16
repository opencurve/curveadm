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
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/spf13/cobra"
)

type execOptions struct {
	id  string
	cmd string
}

func NewExecCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options execOptions

	cmd := &cobra.Command{
		Use:   "exec ID [OPTIONS]",
		Short: "Exec a cmd in service container",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.id = args[0]
			options.cmd = strings.Join(args[1:], " ")
			return curveadm.CheckId(options.id)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExec(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runExec(curveadm *cli.CurveAdm, options execOptions) error {
	// 1) parse cluster topology
	dcs, err := curveadm.ParseTopology()
	if err != nil {
		return err
	}

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
	dc := dcs[0]
	serviceId := curveadm.GetServiceId(dc.GetId())
	containerId, err := curveadm.GetContainerId(serviceId)
	if err != nil {
		return err
	}

	// 4) exec cmd in remote container
	return tools.ExecCmdInRemoteContainer(curveadm, dc.GetHost(), containerId, options.cmd)
}

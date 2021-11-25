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

package command

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type enterOptions struct {
	id string
}

const (
	FORMAT_ENTER_CMD = "ssh -tt %s@%s -p %d -i %s " +
		"sudo docker exec -it %s /bin/bash -c \"cd %s; /bin/bash\""
)

func NewEnterCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options enterOptions

	cmd := &cobra.Command{
		Use:   "enter ID",
		Short: "Enter service container",
		Args:  utils.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.id = args[0]
			return runEnter(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func connectContainer(curveadm *cli.CurveAdm, dc *configure.DeployConfig, containerId string) error {
	user := dc.GetUser()
	host := dc.GetHost()
	sshPort := dc.GetSshPort()
	privateKeyFile := dc.GetPrivateKeyFile()

	cmd := utils.NewCommand(FORMAT_ENTER_CMD, user, host, sshPort, privateKeyFile, containerId, dc.GetServicePrefix())
	cmd.Stdout = curveadm.Out()
	cmd.Stderr = curveadm.Err()
	cmd.Stdin = curveadm.In()
	return cmd.Run()
}

func runEnter(curveadm *cli.CurveAdm, options enterOptions) error {
	dcs, err := configure.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	dcs = configure.FilterDeployConfig(dcs, configure.FilterOption{
		Id:   options.id,
		Host: "*",
		Role: "*",
	})

	if len(dcs) != 1 {
		return fmt.Errorf("invalid service id")
	} else if containerId, err := curveadm.Storage().GetContainerId(options.id); err != nil {
		return curveadm.NewPromptError(err, "")
	} else if err := connectContainer(curveadm, dcs[0], containerId); err != nil {
		return err
	}
	return nil
}

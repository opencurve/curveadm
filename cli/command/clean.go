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
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui/common"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	cleanExample = `Examples:
  $ curveadm clean                               # Clean everything for all services
  $ curveadm clean --only='log,data'             # Clean log and data for all services
  $ curveadm clean --role=etcd --only=container  # Clean container for etcd services`

	supportOnlyFlag = map[string]bool{
		task.ITEM_LOG:       true,
		task.ITEM_DATA:      true,
		task.ITEM_CONTAINER: true,
	}
)

type cleanOptions struct {
	id             string
	role           string
	host           string
	only           []string
	withoutRecycle bool
}

func NewCleanCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options cleanOptions

	cmd := &cobra.Command{
		Use:     "clean [OPTIONS]",
		Short:   "Clean service's environment",
		Args:    cliutil.NoArgs,
		Example: cleanExample,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			for _, v := range options.only {
				if !supportOnlyFlag[v] {
					return fmt.Errorf("unknown flag '%s' in only option", v)
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runClean(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.id, "id", "", "*", "Specify service id")
	flags.StringVarP(&options.role, "role", "", "*", "Specify service role")
	flags.StringVarP(&options.host, "host", "", "*", "Specify service host")
	flags.StringSliceVarP(&options.only, "only", "o", []string{"log", "data", "container"}, "Specify clean item")
	flags.BoolVarP(&options.withoutRecycle, "no-recycle", "", false, "Remove data directory directly instead of recycle chunks")

	return cmd
}

func runClean(curveadm *cli.CurveAdm, options cleanOptions) error {
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	}

	id := options.id
	role := options.role
	host := options.host
	only := options.only
	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   id,
		Role: role,
		Host: host,
	})

	if len(dcs) == 0 {
		curveadm.WriteOut("Clean nothing\n")
		return nil
	}

	if pass := tui.ConfirmYes(tui.PromptCleanService(role, host, only)); !pass {
		curveadm.WriteOut("Clean canceled\n")
		return nil
	}

	curveadm.WriteOut("\n")
	memStorage := curveadm.MemStorage()
	memStorage.Set(task.KEY_CLEAN_ITEMS, only)
	memStorage.Set(task.KEY_RECYCLE, options.withoutRecycle == false)
	err = tasks.ExecTasks(tasks.CLEAN_SERVICE, curveadm, dcs)
	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	return err
}

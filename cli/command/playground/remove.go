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
 * Created Date: 2022-06-23
 * Author: Jingli Chen (Wine93)
 */

package playground

import (
	"github.com/opencurve/curveadm/cli/cli"
	pg "github.com/opencurve/curveadm/internal/configure/playground"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/log"
	"github.com/spf13/cobra"
)

type removeOptions struct {
	name string
}

func NewRemoveCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options removeOptions

	cmd := &cobra.Command{
		Use:     "rm PLAYGROUND",
		Aliases: []string{"delete"},
		Short:   "Remove playground",
		Args:    cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runRemove(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	return cmd
}

func runRemove(curveadm *cli.CurveAdm, options removeOptions) error {
	name := options.name
	playgrounds, err := curveadm.Storage().GetPlaygrounds(name)
	if err != nil {
		return err
	} else if len(playgrounds) == 0 {
		curveadm.WriteOutln("Playground '%s' not exist.", name)
		return nil
	}

	playground := playgrounds[0]
	err = tasks.ExecTasks(tasks.REMOVE_PLAYGROUND, curveadm, &pg.PlaygroundConfig{
		Name:       name,
		Mountpoint: playground.MountPoint,
	})
	if err == nil {
		err = curveadm.Storage().DeletePlayground(options.name)
		if err != nil {
			log.Error("DeletePlayground", log.Field("error", err))
		}
	}

	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	curveadm.WriteOutln("Playground '%s' removed.", name)
	return nil
}

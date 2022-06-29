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
	"fmt"
	"path/filepath"
	"time"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	pg "github.com/opencurve/curveadm/internal/configure/playground"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/task/tasks"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/log"
	"github.com/spf13/cobra"
)

const (
	FORMAT_PLAYGROUND_NAME = "playground-%s-%d" // playground-curvebs-1656035415

	KIND_CURVEBS = topology.KIND_CURVEBS
	KIND_CURVEFS = topology.KIND_CURVEFS
)

var (
	supportKind = map[string]bool{
		KIND_CURVEBS: true,
		KIND_CURVEFS: true,
	}
)

type runOptions struct {
	kind           string
	mountPoint     string
	containerImage string
}

func checkOptions(options runOptions) error {
	kind := options.kind
	mountPoint := options.mountPoint
	if !supportKind[kind] {
		return fmt.Errorf("unsupport kind '%s'", options.kind)
	} else if kind == KIND_CURVEFS && len(mountPoint) == 0 {
		return fmt.Errorf("you must specify mountpoint for CurveFS")
	} else if kind == KIND_CURVEFS && !utils.PathExist(mountPoint) {
		return fmt.Errorf("mountpoint '%s' not exist", mountPoint)
	} else if kind == KIND_CURVEFS && !filepath.IsAbs(mountPoint) {
		return fmt.Errorf("mountpoint '%s' must be an absolute path", mountPoint)
	}
	return nil
}

func NewRunCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options runOptions

	cmd := &cobra.Command{
		Use:     "run [OPTIONS]",
		Aliases: []string{"create"},
		Short:   "Run playground",
		Args:    cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := checkOptions(options); err != nil {
				return err
			}
			return runRun(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.kind, "kind", "k", "curvefs", "Specify the type of playground (curvebs/curvefs)")
	flags.StringVarP(&options.mountPoint, "mountpoint", "m", "", "Specify the mountpoint for CurveFS playground")
	flags.StringVarP(&options.containerImage, "container_image", "c", "", "Specify the playground container image")

	return cmd
}

func runRun(curveadm *cli.CurveAdm, options runOptions) error {
	kind := options.kind
	name := fmt.Sprintf(FORMAT_PLAYGROUND_NAME, kind, time.Now().Unix())
	err := curveadm.Storage().InsertPlayground(name, options.mountPoint, "STOPPED")
	if err != nil {
		return err
	}

	err = tasks.ExecTasks(tasks.RUN_PLAYGROUND, curveadm, &pg.PlaygroundConfig{
		Kind:           options.kind,
		Name:           name,
		ContainerImage: options.containerImage,
		Mountpoint:     options.mountPoint,
	})
	if err == nil {
		err = curveadm.Storage().SetPlaygroundStatus(name, "RUNNING")
		if err != nil {
			log.Error("SetPlaygroundStatus", log.Field("error", err))
		}
	}

	if err != nil {
		return curveadm.NewPromptError(err, "")
	}
	curveadm.WriteOutln(color.GreenString("Playground '%s' successfully deployed ^_^", name))
	return nil
}

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
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
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

	RUN_PLAYGROUND_PLAYBOOK_STEPS = []int{
		playbook.RUN_PLAYGROUND,
	}
)

type runOptions struct {
	name           string
	kind           string
	mountPoint     string
	containerImage string
}

func checkRunOptions(curveadm *cli.CurveAdm, options runOptions) error {
	kind := options.kind
	mountPoint := options.mountPoint
	if !supportKind[kind] {
		return errno.ERR_UNSUPPORT_PLAYGROUND_KIND.
			F("kind=%s", kind)
	} else if kind == KIND_CURVEFS && len(mountPoint) == 0 {
		return errno.ERR_MUST_SPECIFY_MOUNTPOINT_FOR_CURVEFS_PLAYGROUND
	} else if kind == KIND_CURVEFS && !filepath.IsAbs(mountPoint) {
		return errno.ERR_PLAYGROUND_MOUNTPOINT_REQUIRE_ABSOLUTE_PATH.
			F("mountPoint=%s", mountPoint)
	} else if kind == KIND_CURVEFS && !utils.PathExist(mountPoint) {
		return errno.ERR_PLAYGROUND_MOUNTPOINT_NOT_EXIST.
			F("mountPoint=%s", mountPoint)
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
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkRunOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.kind, "kind", "k", "curvefs", "Specify the type of playground (curvebs/curvefs)")
	flags.StringVar(&options.mountPoint, "mountpoint", "m", "Specify the mountpoint for CurveFS playground")
	flags.StringVar(&options.containerImage, "container_image", "i", "Specify the playground container image")

	return cmd
}

func genRunPlaybook(curveadm *cli.CurveAdm,
	options runOptions) (*playbook.Playbook, error) {
	steps := RUN_PLAYGROUND_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type: step,
			Configs: &configure.PlaygroundConfig{
				Kind:       options.kind,
				Name: options.name,
				Mountpoint: options.mountPoint,
			},
		})
	}
	return pb, nil
}

func runRun(curveadm *cli.CurveAdm, options runOptions) error {
	// 1) insert a new playground
	kind := options.kind
	options.name = fmt.Sprintf(FORMAT_PLAYGROUND_NAME, kind, time.Now().Unix())
	err := curveadm.Storage().InsertPlayground(options.name, options.mountPoint, "STOPPED")
	if err != nil {
		return errno.ERR_INSERT_PLAYGROUND_FAILED.E(err)
	}

	// 2) generate remove playground
	pb, err := genRunPlaybook(curveadm, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln(color.GreenString("Playground '%s' successfully deployed ^_^",
		options.name))
	return nil
}

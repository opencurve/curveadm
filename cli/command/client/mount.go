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

package client

import (
	"strings"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/fs"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	MOUNT_EXAMPLE = `Examples:
  $ curveadm mount /s3_001     /path/to/mount --host machine -c client.yaml [--fstype s3]    # Mount a s3 CurveFS '/s3_001' to '/path/to/mount'
  $ curveadm mount /volume_001 /path/to/mount --host machine -c client.yaml --fstype volume  # Mount a volume CurveFS '/volume_001' to '/path/to/mount'`
)

var (
	MOUNT_PLAYBOOK_STEPS = []int{
		// TODO(P0): create filesystem
		playbook.CHECK_KERNEL_MODULE,
		playbook.CHECK_CLIENT_S3,
		playbook.MOUNT_FILESYSTEM,
	}
)

type mountOptions struct {
	host        string
	mountFSName string
	mountFSType string
	mountPoint  string
	filename    string
}

func checkMountOptions(curveadm *cli.CurveAdm, options mountOptions) error {
	if !strings.HasPrefix(options.mountPoint, "/") {
		return errno.ERR_FS_MOUNTPOINT_REQUIRE_ABSOLUTE_PATH.
			F("mount point: %s", options.mountPoint)
	}
	return nil
}

func NewMountCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options mountOptions

	cmd := &cobra.Command{
		Use:     "mount NAME_OF_CURVEFS MOUNT_POINT [OPTIONS]",
		Short:   "Mount filesystem",
		Args:    utils.ExactArgs(2),
		Example: MOUNT_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.mountFSName = args[0]
			options.mountPoint = args[1]
			return checkMountOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.mountFSName = args[0]
			options.mountPoint = args[1]
			return runMount(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")
	flags.StringVar(&options.mountFSType, "fstype", "s3", "Specify fs data backend")

	return cmd
}

func genMountPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options mountOptions) (*playbook.Playbook, error) {
	steps := MOUNT_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: ccs,
			Options: map[string]interface{}{
				comm.KEY_MOUNT_OPTIONS: fs.MountOptions{
					Host:        options.host,
					MountFSName: options.mountFSName,
					MountFSType: options.mountFSType,
					MountPoint:  utils.TrimSuffixRepeat(options.mountPoint, "/"),
				},
				comm.KEY_CLIENT_HOST:              options.host, // for checker
				comm.KEY_CHECK_KERNEL_MODULE_NAME: comm.KERNERL_MODULE_FUSE,
			},
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: step == playbook.CHECK_CLIENT_S3,
			},
		})
	}
	return pb, nil
}

func runMount(curveadm *cli.CurveAdm, options mountOptions) error {
	// 1) parse client configure
	cc, err := configure.ParseClientConfig(options.filename)
	if err != nil {
		return err
	} else if cc.GetKind() != topology.KIND_CURVEFS {
		return errno.ERR_REQUIRE_CURVEFS_KIND_CLIENT_CONFIGURE_FILE.
			F("kind: %s", cc.GetKind())
	}

	// 2) generate mount playbook
	pb, err := genMountPlaybook(curveadm, []*configure.ClientConfig{cc}, options)
	if err != nil {
		return err
	}

	// 3) run playground
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print success prompt
	curveadm.WriteOutln("")
	curveadm.WriteOutln(color.GreenString("Mount %s to %s (%s) success ^_^"),
		options.mountFSName, options.mountPoint, options.host)
	return nil
}

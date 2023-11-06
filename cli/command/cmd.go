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
	"fmt"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/cli/command/client"
	"github.com/opencurve/curveadm/cli/command/cluster"
	"github.com/opencurve/curveadm/cli/command/config"
	"github.com/opencurve/curveadm/cli/command/disks"
	"github.com/opencurve/curveadm/cli/command/hosts"
	"github.com/opencurve/curveadm/cli/command/http"
	"github.com/opencurve/curveadm/cli/command/install"
	"github.com/opencurve/curveadm/cli/command/monitor"
	"github.com/opencurve/curveadm/cli/command/pfs"
	"github.com/opencurve/curveadm/cli/command/playground"
	"github.com/opencurve/curveadm/cli/command/target"
	"github.com/opencurve/curveadm/cli/command/website"
	"github.com/opencurve/curveadm/internal/errno"
	tools "github.com/opencurve/curveadm/internal/tools/upgrade"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var curveadmExample = `Examples:
  $ curveadm playground run --kind curvebs  # Run a CurveBS playground quickly
  $ curveadm cluster add c1                 # Add a cluster named 'c1'
  $ curveadm deploy                         # Deploy current cluster
  $ curveadm stop                           # Stop current cluster service
  $ curveadm clean                          # Clean current cluster
  $ curveadm enter 6ff561598c6f             # Enter specified service container
  $ curveadm -u                             # Upgrade curveadm itself to the latest version`

type rootOptions struct {
	debug   bool
	upgrade bool
}

func addSubCommands(cmd *cobra.Command, curveadm *cli.CurveAdm) {
	cmd.AddCommand(
		client.NewClientCommand(curveadm),         // curveadm client
		cluster.NewClusterCommand(curveadm),       // curveadm cluster ...
		config.NewConfigCommand(curveadm),         // curveadm config ...
		hosts.NewHostsCommand(curveadm),           // curveadm hosts ...
		disks.NewDisksCommand(curveadm),           // curveadm disks ...
		playground.NewPlaygroundCommand(curveadm), // curveadm playground ...
		target.NewTargetCommand(curveadm),         // curveadm target ...
		pfs.NewPFSCommand(curveadm),               // curveadm pfs ...
		monitor.NewMonitorCommand(curveadm),       // curveadm monitor ...
		http.NewHttpCommand(curveadm),             // curveadm http
		website.NewWebsiteCommand(curveadm),       // curveadm website ...
		install.NewInstallCommand(curveadm),       // curveadm install

		NewAuditCommand(curveadm),      // curveadm audit
		NewCleanCommand(curveadm),      // curveadm clean
		NewCompletionCommand(curveadm), // curveadm completion
		NewDeployCommand(curveadm),     // curveadm deploy
		NewEnterCommand(curveadm),      // curveadm enter
		NewExecCommand(curveadm),       // curveadm exec
		NewFormatCommand(curveadm),     // curveadm format
		NewMigrateCommand(curveadm),    // curveadm migrate
		NewPrecheckCommand(curveadm),   // curveadm precheck
		NewReloadCommand(curveadm),     // curveadm reload
		NewRestartCommand(curveadm),    // curveadm restart
		NewScaleOutCommand(curveadm),   // curveadm scale-out
		NewStartCommand(curveadm),      // curveadm start
		NewStatusCommand(curveadm),     // curveadm status
		NewStopCommand(curveadm),       // curveadm stop
		NewSupportCommand(curveadm),    // curveadm support
		NewUpgradeCommand(curveadm),    // curveadm upgrade
		// commonly used shorthands
		hosts.NewSSHCommand(curveadm),      // curveadm ssh
		hosts.NewPlaybookCommand(curveadm), // curveadm playbook
		client.NewMapCommand(curveadm),     // curveadm map
		client.NewMountCommand(curveadm),   // curveadm mount
		client.NewUnmapCommand(curveadm),   // curveadm unmap
		client.NewUmountCommand(curveadm),  // curveadm umount
	)
}

func setupRootCommand(cmd *cobra.Command, curveadm *cli.CurveAdm) {
	cmd.SetVersionTemplate("CurveAdm v{{.Version}}\n")
	cliutil.SetFlagErrorFunc(cmd)
	cliutil.SetHelpTemplate(cmd)
	cliutil.SetUsageTemplate(cmd)
	cliutil.SetErr(cmd, curveadm)
}

func NewCurveAdmCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options rootOptions

	cmd := &cobra.Command{
		Use:     "curveadm [OPTIONS] COMMAND [ARGS...]",
		Short:   "Deploy and manage CurveBS/CurveFS cluster",
		Version: cli.Version,
		Example: curveadmExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if options.debug {
				return errno.List()
			} else if options.upgrade {
				return tools.Upgrade2Latest(cli.Version)
			} else if len(args) == 0 {
				return cliutil.ShowHelp(curveadm.Err())(cmd, args)
			}

			return fmt.Errorf("curveadm: '%s' is not a curveadm command.\n"+
				"See 'curveadm --help'", args[0])
		},
		SilenceUsage:          true, // silence usage when an error occurs
		DisableFlagsInUseLine: true,
	}

	cmd.Flags().BoolP("version", "v", false, "Print version information and quit")
	cmd.PersistentFlags().BoolP("help", "h", false, "Print usage")
	cmd.Flags().BoolVarP(&options.debug, "debug", "d", false, "Print debug information")
	cmd.Flags().BoolVarP(&options.upgrade, "upgrade", "u", false, "Upgrade curveadm itself to the latest version")

	addSubCommands(cmd, curveadm)
	setupRootCommand(cmd, curveadm)

	return cmd
}

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
 * Created Date: 2021-12-27
 * Author: Jingli Chen (Wine93)
 */

package command

import (
	"fmt"

	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	tuicomm "github.com/opencurve/curveadm/internal/tui/common"
	tui "github.com/opencurve/curveadm/internal/tui/format"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	FORMAT_EXAMPLE = `Examples:
  $ curveadm format -f /path/to/format.yaml           # Format chunkfile pool with specified configure file
  $ curveadm format --status -f /path/to/format.yaml  # Display formatting status
  $ curveadm format --stop   -f /path/to/format.yaml  # Stop formatting progress
  $ curveadm format --debug  -f /path/to/format.yaml  # Format chunkfile with debug mode
  $ curveadm format --clean  -f /path/to/format.yaml  # clean the container left by debug mode`
)

var (
	FORMAT_PLAYBOOK_STEPS = []int{
		playbook.FORMAT_CHUNKFILE_POOL,
	}

	FORMAT_STATUS_PLAYBOOK_STEPS = []int{
		playbook.GET_FORMAT_STATUS,
	}
	// FORMAT_STOP_PLAYBOOK_STEPS stop formatting step
	FORMAT_STOP_PLAYBOOK_STEPS = []int{
		playbook.STOP_FORMAT,
	}

	FORMAT_CLEAN_PLAYBOOK_STEPS = []int{
		playbook.CLEAN_FORMAT,
	}
)

type formatOptions struct {
	filename   string
	showStatus bool
	stopFormat bool
	debug      bool
	clean      bool
}

func checkFormatOptions(options formatOptions) error {
	opts := []bool{
		options.showStatus,
		options.stopFormat,
		options.debug,
		options.clean,
	}

	trueCount := 0
	for _, opt := range opts {
		if opt {
			trueCount++
			if trueCount > 1 {
				return errno.ERR_UNSUPPORT_CONFIGURE_VALUE_TYPE
			}
		}
	}

	return nil
}

func NewFormatCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options formatOptions

	cmd := &cobra.Command{
		Use:     "format [OPTIONS]",
		Short:   "Format chunkfile pool",
		Args:    cliutil.NoArgs,
		Example: FORMAT_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			return checkFormatOptions(options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFormat(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.filename, "formatting", "f", "format.yaml", "Specify the configure file for formatting chunkfile pool")
	flags.BoolVar(&options.showStatus, "status", false, "Show formatting status")
	flags.BoolVar(&options.stopFormat, "stop", false, "Stop formatting progress")
	flags.BoolVar(&options.debug, "debug", false, "Debug formatting progress")
	flags.BoolVar(&options.clean, "clean", false, "Clean the Container")

	return cmd
}

func genFormatPlaybook(curveadm *cli.CurveAdm,
	fcs []*configure.FormatConfig,
	options formatOptions) (*playbook.Playbook, error) {
	if len(fcs) == 0 {
		return nil, errno.ERR_NO_DISK_FOR_FORMATTING
	}

	showStatus := options.showStatus
	stopFormat := options.stopFormat
	debug := options.debug
	clean := options.clean

	steps := FORMAT_PLAYBOOK_STEPS
	if showStatus {
		steps = FORMAT_STATUS_PLAYBOOK_STEPS
	}
	if stopFormat {
		steps = FORMAT_STOP_PLAYBOOK_STEPS
	}
	if clean {
		steps = FORMAT_CLEAN_PLAYBOOK_STEPS
	}

	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		// options
		options := map[string]interface{}{}
		if step == playbook.FORMAT_CHUNKFILE_POOL {
			options[comm.DEBUG_MODE] = debug
		}
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: fcs,
			ExecOptions: playbook.ExecOptions{
				SilentSubBar: showStatus,
			},
			Options: options,
		})
	}
	return pb, nil
}

func displayFormatStatus(curveadm *cli.CurveAdm) string {
	statuses := []bs.FormatStatus{}
	v := curveadm.MemStorage().Get(comm.KEY_ALL_FORMAT_STATUS)
	if v != nil {
		m := v.(map[string]bs.FormatStatus)
		for _, status := range m {
			statuses = append(statuses, status)
		}
	}
	return tui.FormatStatus(statuses)
}

func runFormat(curveadm *cli.CurveAdm, options formatOptions) error {
	var err error
	var fcs []*configure.FormatConfig
	diskRecords := curveadm.DiskRecords()
	debug := options.debug
	if debug {
		curveadm.SetDebugLevel()
	}

	// 1) parse format config from yaml file or database
	if len(diskRecords) == 0 {
		fcs, err = configure.ParseFormat(options.filename)
		if err != nil {
			return err
		}
	} else {
		for _, dr := range diskRecords {
			containerImage := configure.DEFAULT_CONTAINER_IMAGE
			if len(dr.ContainerImage) > 0 {
				containerImage = dr.ContainerImage
			}
			disk := fmt.Sprintf("%s:%s:%d", dr.Device, dr.MountPoint, dr.FormatPercent)
			fc, err := configure.NewFormatConfig(containerImage, dr.Host, disk)
			if err != nil {
				return err
			}

			fc.FromDiskRecord = true

			// "ServiceMountDevice=0" means write disk UUID into /etc/fstab for host mounting.
			// "ServiceMountDevice=1" means not to update /etc/fstab, the disk UUID will be wrote
			// into the config of service(chunkserver) container for disk automatic mounting.
			if dr.ServiceMountDevice != 0 {
				fc.ServiceMountDevice = true
			}

			if dr.ChunkServerID != comm.DISK_DEFAULT_NULL_CHUNKSERVER_ID {
				// skip formatting the disk with nonempty chunkserver id
				continue
			}
			fcs = append(fcs, fc)
		}
	}

	// 2) generate start playbook
	pb, err := genFormatPlaybook(curveadm, fcs, options)
	if err != nil {
		return err
	}

	// 3) run playbook
	err = pb.Run()
	if err != nil {
		return err
	}

	// 4) print status or prompt
	if options.showStatus {
		output := displayFormatStatus(curveadm)
		curveadm.WriteOutln("")
		curveadm.WriteOut("%s", output)
	} else {
		tuicomm.PromptFormat()
	}
	return nil
}

// for http service
func Format(curveadm *cli.CurveAdm, status bool) (string, error) {
	var output string
	var err error
	var fcs []*configure.FormatConfig
	diskRecords, err := curveadm.Storage().GetDisk(comm.DISK_FILTER_ALL)
	if err != nil {
		return output, errno.ERR_GET_DISK_RECORDS_FAILED.E(err)
	}

	if len(diskRecords) != 0 {
		for _, dr := range diskRecords {
			containerImage := configure.DEFAULT_CONTAINER_IMAGE
			if len(dr.ContainerImage) > 0 {
				containerImage = dr.ContainerImage
			}
			disk := fmt.Sprintf("%s:%s:%d", dr.Device, dr.MountPoint, dr.FormatPercent)
			fc, err := configure.NewFormatConfig(containerImage, dr.Host, disk)
			if err != nil {
				return output, err
			}
			fc.FromDiskRecord = true
			if dr.ServiceMountDevice != 0 {
				fc.ServiceMountDevice = true
			}

			if dr.ChunkServerID != comm.DISK_DEFAULT_NULL_CHUNKSERVER_ID {
				// skip formatting the disk with nonempty chunkserver id
				continue
			}
			fcs = append(fcs, fc)
		}
	}
	// gen playbook
	pb, err := genFormatPlaybook(curveadm, fcs, formatOptions{showStatus: status})
	if err != nil {
		return output, err
	}
	// run playbook
	err = pb.Run()
	if err != nil {
		return output, err
	}
	output = displayFormatStatus(curveadm)
	return output, nil
}

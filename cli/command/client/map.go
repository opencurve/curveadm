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
 * Created Date: 2022-01-09
 * Author: Jingli Chen (Wine93)
 */

package client

import (
	"strconv"
	"strings"

	"github.com/dustin/go-humanize"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	comm "github.com/opencurve/curveadm/internal/common"
	"github.com/opencurve/curveadm/internal/configure"
	"github.com/opencurve/curveadm/internal/configure/topology"
	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/playbook"
	"github.com/opencurve/curveadm/internal/task/task/bs"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

const (
	MAP_EXAMPLE = `Examples:
  $ curveadm map user:/volume --host machine1 --create                  # Map volume which created by automatic
  $ curveadm map user:/volume --host machine1 --size=10GiB --create     # Map volume which size is 10GiB and created by automatic
  $ curveadm map user:/volume --host machine1 --create --poolset ssd    # Map volume created by automatic in poolset 'ssd'
  $ curveadm map user:/volume --host machine1 -c /path/to/client.yaml   # Map volume with specified configure file`
)

var (
	MAP_PLAYBOOK_STEPS = []int{
		// TODO(P0): check client time greater than mds
		//playbook.CHECK_MDS_ADDRESS, // TODO(P0)
		playbook.CHECK_KERNEL_MODULE,
		playbook.START_NEBD_SERVICE,
		playbook.CREATE_VOLUME,
		playbook.MAP_IMAGE,
	}
)

type mapOptions struct {
	image       string
	host        string
	size        string
	create      bool
	filename    string
	noExclusive bool
	poolset     string
}

func ParseImage(image string) (user, name string, err error) {
	items := strings.Split(image, ":")
	if len(items) != 2 || len(items[0]) == 0 || len(items[1]) == 0 {
		err = errno.ERR_INVALID_VOLUME_FORMAT.
			F("volume: %s", image)
		return
	}

	user, name = items[0], items[1]
	if user == "root" {
		err = errno.ERR_ROOT_VOLUME_USER_NOT_ALLOWED.
			F("volume user: %s", user)
	} else if !strings.HasPrefix(name, "/") {
		err = errno.ERR_VOLUME_NAME_MUST_START_WITH_SLASH_PREFIX.
			F("volume name: %s", name)
	} else if strings.Contains(name, "_") {
		err = errno.ERR_VOLUME_NAME_CAN_NOT_CONTAIN_UNDERSCORE.
			F("volume name: %s", name)
	}
	return
}

func ParseSize(size string) (int, error) {
	if !strings.HasSuffix(size, "GiB") {
		return 0, errno.ERR_VOLUME_SIZE_MUST_END_WITH_GiB_SUFFIX.
			F("size: %s", size)
	}

	size = strings.TrimSuffix(size, "GiB")
	n, err := strconv.Atoi(size)
	if err != nil || n <= 0 {
		return 0, errno.ERR_VOLUME_SIZE_REQUIRES_POSITIVE_INTEGER.
			F("size: %sGiB", size)
	} else if n%10 != 0 {
		return 0, errno.ERR_VOLUME_SIZE_MUST_BE_MULTIPLE_OF_10_GiB.
			F("size: %sGiB", size)
	}

	return n, nil
}

func ParseBlockSize(blocksize string) (uint64, error) {
	if !strings.HasSuffix(blocksize, "B") {
		return 0, errno.ERR_VOLUME_BLOCKSIZE_MUST_END_WITH_BYTE_SUFFIX.
			F("blocksize: %s", blocksize)
	}
	blocksize = strings.TrimSuffix(blocksize, "B")
	m, err := humanize.ParseBytes(blocksize)
	if err != nil || m <= 0 {
		return 0, errno.ERR_VOLUME_BLOCKSIZE_REQUIRES_POSITIVE_INTEGER.
			F("blocksize: %s", humanize.IBytes(m))
	} else if m%512 != 0 {
		return 0, errno.ERR_VOLUME_BLOCKSIZE_BE_MULTIPLE_OF_512.
			F("blocksize: %s", humanize.IBytes(m))
	}
	return m, nil
}
func checkMapOptions(curveadm *cli.CurveAdm, options mapOptions) error {
	if _, _, err := ParseImage(options.image); err != nil {
		return err
	} else if _, err = ParseSize(options.size); err != nil {
		return err
	} else if !utils.PathExist(options.filename) {
		return errno.ERR_CLIENT_CONFIGURE_FILE_NOT_EXIST.
			F("file path: %s", utils.AbsPath(options.filename))
	}
	return nil
}

func NewMapCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options mapOptions

	cmd := &cobra.Command{
		Use:     "map USER:VOLUME [OPTIONS]",
		Short:   "Map a volume to nbd device",
		Args:    utils.ExactArgs(1),
		Example: MAP_EXAMPLE,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.image = args[0]
			return checkMapOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMap(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVar(&options.host, "host", "localhost", "Specify target host")
	flags.BoolVar(&options.create, "create", false, "Create volume iff not exist")
	flags.BoolVar(&options.noExclusive, "no-exclusive", false, "Map volume non exclusive")
	flags.StringVar(&options.size, "size", "10GiB", "Specify volume size")
	flags.StringVarP(&options.filename, "conf", "c", "client.yaml", "Specify client configuration file")
	flags.StringVar(&options.poolset, "poolset", "", "Specify the poolset")
	return cmd
}

func genMapPlaybook(curveadm *cli.CurveAdm,
	ccs []*configure.ClientConfig,
	options mapOptions) (*playbook.Playbook, error) {
	user, name, _ := ParseImage(options.image)
	size, _ := ParseSize(options.size)
	steps := MAP_PLAYBOOK_STEPS
	pb := playbook.NewPlaybook(curveadm)
	for _, step := range steps {
		pb.AddStep(&playbook.PlaybookStep{
			Type:    step,
			Configs: ccs,
			Options: map[string]interface{}{
				comm.KEY_MAP_OPTIONS: bs.MapOptions{
					Host:        options.host,
					User:        user,
					Volume:      name,
					Size:        size,
					Create:      options.create,
					NoExclusive: options.noExclusive,
					Poolset:     options.poolset,
				},
				comm.KEY_CLIENT_HOST:              options.host, // for checker
				comm.KEY_CHECK_KERNEL_MODULE_NAME: comm.KERNERL_MODULE_NBD,
			},
		})
	}
	return pb, nil
}

// TODO(P1): unmap by id
func runMap(curveadm *cli.CurveAdm, options mapOptions) error {
	// 1) parse client configure
	cc, err := configure.ParseClientConfig(options.filename)
	if err != nil {
		return err
	} else if cc.GetKind() != topology.KIND_CURVEBS {
		return errno.ERR_REQUIRE_CURVEBS_KIND_CLIENT_CONFIGURE_FILE.
			F("kind: %s", cc.GetKind())
	}

	// 2) generate map playbook
	pb, err := genMapPlaybook(curveadm, []*configure.ClientConfig{cc}, options)
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
	curveadm.WriteOutln(color.GreenString("Map %s to %s nbd device success ^_^"),
		options.image, options.host)
	return nil
}

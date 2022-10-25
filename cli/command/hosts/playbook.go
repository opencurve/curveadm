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
 * Created Date: 2022-10-25
 * Author: Jingli Chen (Wine93)
 */

package hosts

// NOTE: playbook under beta version
import (
	"fmt"
	"path"
	"strings"
	"sync"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/hosts"
	"github.com/opencurve/curveadm/internal/tools"
	"github.com/opencurve/curveadm/internal/utils"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

var (
	retC chan result
	wg   sync.WaitGroup
)

type result struct {
	host string
	out  string
	err  error
}

type playbookOptions struct {
	filepath string
	labels   []string
}

func checkPlaybookOptions(curveadm *cli.CurveAdm, options playbookOptions) error {
	// TODO: added error code
	if !utils.PathExist(options.filepath) {
		return fmt.Errorf("%s: no such file", options.filepath)
	}
	return nil
}

func NewPlaybookCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options playbookOptions

	cmd := &cobra.Command{
		Use:   "playbook [OPTIONS] PLAYBOOK",
		Short: "Execute playbook",
		Args:  cliutil.ExactArgs(1),
		PreRunE: func(cmd *cobra.Command, args []string) error {
			options.filepath = args[0]
			return checkPlaybookOptions(curveadm, options)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			options.filepath = args[0]
			return runPlaybook(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringSliceVarP(&options.labels, "labels", "l", []string{}, "Specify the host labels")

	return cmd
}

func execute(curveadm *cli.CurveAdm, hc *hosts.HostConfig, source string) {
	defer func() { wg.Done() }()
	name := hc.GetHost()
	target := path.Join("/tmp", utils.RandString(8))
	err := tools.Scp(curveadm, name, source, target)
	if err != nil {
		retC <- result{host: name, err: err}
		return
	}

	defer func() {
		command := fmt.Sprintf("rm -rf %s", target)
		tools.ExecuteRemoteCommand(curveadm, name, command)
	}()

	items := []string{}
	for _, env := range hc.GetEnvs() {
		items = append(items, env)
	}
	items = append(items, fmt.Sprintf("bash %s", target))
	command := strings.Join(items, " ")
	out, err := tools.ExecuteRemoteCommand(curveadm, name, command)
	retC <- result{host: name, out: out, err: err}
}

func output(curveadm *cli.CurveAdm, total int) {
	curveadm.WriteOutln("TOTAL: %d hosts", total)
	for {
		select {
		case ret := <-retC:
			curveadm.WriteOutln("")
			out, err := ret.out, ret.err
			curveadm.WriteOutln("--- %s [%s]", ret.host,
				utils.Choose(err == nil, color.GreenString("SUCCESS"), color.RedString("FAIL")))
			if err != nil {
				curveadm.WriteOut(out)
				curveadm.WriteOutln(err.Error())
			} else if len(out) > 0 {
				curveadm.WriteOut(out)
			}
		}
	}
}

func runPlaybook(curveadm *cli.CurveAdm, options playbookOptions) error {
	var hcs []*hosts.HostConfig
	var err error
	hosts := curveadm.Hosts()
	if len(hosts) > 0 {
		hcs, err = filter(hosts, options.labels) // filter hosts
		if err != nil {
			return err
		}
	}

	retC = make(chan result)
	wg.Add(len(hcs))
	go output(curveadm, len(hcs))
	for _, hc := range hcs {
		go execute(curveadm, hc, options.filepath)
	}
	wg.Wait()
	return nil
}

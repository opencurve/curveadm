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
 * Created Date: 2022-05-17
 * Author: Jingli Chen (Wine93)
 */

package plugin

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/opencurve/curveadm/cli/cli"
	config "github.com/opencurve/curveadm/internal/configure/plugin"
	"github.com/opencurve/curveadm/internal/task/tasks"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type runOptions struct {
	pluginName string
	hosts      string
	inventory  string
	args       []string
}

func NewRunCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options runOptions

	cmd := &cobra.Command{
		Use:   "run PLUGIN",
		Short: "Run plugin",
		Args:  cliutil.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			options.pluginName = args[0]
			return runRun(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.hosts, "hosts", "", "", "Specify the targets")
	flags.StringVarP(&options.inventory, "inventory", "i", "", "Specify the path of inventory file")
	flags.StringArrayVarP(&options.args, "arg", "a", []string{}, "Specify plugin run options")

	return cmd
}

func getTargets(curveadm *cli.CurveAdm, options runOptions) ([]config.Target, error) {
	inventoryPath := curveadm.InventoryPath()
	if len(options.inventory) > 0 {
		inventoryPath = options.inventory
	}

	targets, err := config.ParseInventory(inventoryPath)
	if err != nil {
		return nil, err
	}

	m := map[string]config.Target{}
	for _, target := range targets {
		m[target.Name] = target
	}

	// filter target
	targets = []config.Target{}
	items := strings.Split(options.hosts, ":")
	for _, name := range items {
		target, ok := m[name]
		if !ok {
			return nil, fmt.Errorf("Host '%s' not found", name)
		}
		targets = append(targets, target)
	}

	return targets, nil
}

func displayOutput(curveadm *cli.CurveAdm, targets []config.Target) {
	memStorage := curveadm.MemStorage()
	for _, target := range targets {
		title := fmt.Sprintf("%s (%s):", target.Name, target.Host)
		curveadm.WriteOutln("")
		if output := memStorage.Get(target.Host); output != nil {
			curveadm.WriteOutln("%s\n%s", color.BlueString(title), output.(string))
		} else {
			curveadm.WriteOutln("%s %s", color.BlueString(title), color.GreenString("OK"))
		}
	}
}

func runRun(curveadm *cli.CurveAdm, options runOptions) error {
	targets, err := getTargets(curveadm, options)
	if err != nil {
		return err
	}

	plugin, err := curveadm.PluginManager().Load(options.pluginName)
	if err != nil {
		return err
	}

	playbook, err := config.ParsePlaybook(plugin.EntrypointPath)
	if err != nil {
		return err
	}

	pcs, err := playbook.Build(options.pluginName, options.args, targets)
	if err != nil {
		return err
	}

	if err := tasks.ExecTasks(tasks.RUN_PLUGIN, curveadm, pcs); err != nil {
		return curveadm.NewPromptError(err, "")
	}

	displayOutput(curveadm, targets)
	return nil
}

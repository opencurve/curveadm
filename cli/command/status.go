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
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/configure/topology"
	task "github.com/opencurve/curveadm/internal/task/task/common"
	"github.com/opencurve/curveadm/internal/task/tasks"
	tui "github.com/opencurve/curveadm/internal/tui/service"
	cliutil "github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/cobra"
)

type statusOptions struct {
	id          string
	role        string
	host        string
	vebose      bool
	showReplica bool
}

func NewStatusCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options statusOptions

	cmd := &cobra.Command{
		Use:   "status [OPTIONS]",
		Short: "Display service status",
		Args:  cliutil.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.id, "id", "", "*", "Specify service id")
	flags.StringVarP(&options.role, "role", "", "*", "Specify service role")
	flags.StringVarP(&options.host, "host", "", "*", "Specify service host")
	flags.BoolVarP(&options.vebose, "verbose", "v", false, "Verbose output for status")
	flags.BoolVarP(&options.showReplica, "show-replica", "s", false, "Display replica service")

	return cmd
}

func isLeader(addr string, kind string) bool {
	url := "http://" + addr
	if kind == topology.KIND_CURVEBS {
		url = url + "/vars/mds_status?console=1"
	} else {
		url = url + "/vars/curvefs_mds_status?console=1"
	}
	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return false
	}
	if resp.StatusCode == 200 {
		if strings.Contains(string(body), "leader") {
			return true
		}
	}
	return false
}

func getClusterMdsAddr(dcs []*topology.DeployConfig) string {
	mdsList := []string{}
	for _, dc := range dcs {
		role := dc.GetRole()
		if role != "mds" {
			continue
		}
		kind := dc.GetKind()
		ip := dc.GetListenIp()
		port := dc.GetListenPort()
		dummyPort := dc.GetListenDummyPort()
		dummyAddr := string(ip) + ":" + strconv.Itoa(dummyPort)
		leader := isLeader(dummyAddr, kind)

		if leader {
			mdsList = append(mdsList, string(ip)+":"+strconv.Itoa(port)+"(leader)")
		} else {
			mdsList = append(mdsList, string(ip)+":"+strconv.Itoa(port))
		}
	}
	if len(mdsList) > 0 {
		return strings.Join(mdsList, ",")
	} else {
		return "-"
	}
}

func runStatus(curveadm *cli.CurveAdm, options statusOptions) error {
	if curveadm.ClusterId() == -1 {
		curveadm.WriteOut("No cluster, please add a cluster firstly\n")
		return nil
	}
	dcs, err := topology.ParseTopology(curveadm.ClusterTopologyData())
	if err != nil {
		return err
	} else if len(dcs) == 0 {
		curveadm.WriteOut("No service, please check your cluster topology\n")
		return nil
	}

	dcs = curveadm.FilterDeployConfig(dcs, topology.FilterOption{
		Id:   options.id,
		Role: options.role,
		Host: options.host,
	})

	if err := tasks.ExecTasks(tasks.GET_SERVICE_STATUS, curveadm, dcs); err != nil {
		return curveadm.NewPromptError(err, "Did you deploy the cluster?")
	}

	// display status
	statuses := []task.ServiceStatus{}
	m := curveadm.MemStorage().Map
	for _, v := range m {
		status := v.(task.ServiceStatus)
		statuses = append(statuses, status)
	}

	output := tui.FormatStatus(statuses, options.vebose, options.showReplica)
	curveadm.WriteOut("\n")
	curveadm.WriteOut("cluster name    : %s\n", curveadm.ClusterName())
	curveadm.WriteOut("cluster kind    : %s\n", dcs[0].GetKind())
	curveadm.WriteOut("cluster mds addr: %s\n", getClusterMdsAddr(dcs))
	curveadm.WriteOut("\n")
	curveadm.WriteOut("%s", output)
	return nil
}

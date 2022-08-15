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
 * Created Date: 2021-11-24
 * Author: Jingli Chen (Wine93)
 */

package cluster

import (
	"fmt"
	"os"

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/utils"
	log "github.com/opencurve/curveadm/pkg/log/glg"
	"github.com/spf13/cobra"
)

const (
	CLUSTER_NAME        = 0x01
	CLUSTER_DESCRIPTION = 0x02
	CLUSTER_CREATETIME  = 0x03
	CLUSTER_TOPOLOGY    = 0x04
	SERVICE             = 0x10
)

var (
	exportExample = `Examples:
  $ curveadm cluster export my-cluster                     # Export cluster 'my-cluster' 
  $ curveadm cluster export my-cluster -o /path/to/dbfile  # Export cluster 'my-cluster' to specified file`
)

type exportOptions struct {
	name    string
	outfile string
}

func NewExportCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options exportOptions

	cmd := &cobra.Command{
		Use:     "export CLUSTER [OPTIONS]",
		Short:   "Export cluster",
		Args:    utils.ExactArgs(1),
		Example: exportExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runExport(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.outfile, "output", "o", "curveadm.db", "Output to specified database file")

	return cmd
}

func writeItem(file *os.File, id int, value string) error {
	key := fmt.Sprintf("--- %04d %d\n", id, len(value)+1)
	if _, err := file.WriteString(key); err != nil {
		return err
	} else if _, err := file.WriteString(value + "\n"); err != nil {
		return err
	}
	return nil
}

func newMonitorWrite(file *os.File) (func(int, string) bool, func() error) {
	var err error
	return func(id int, value string) bool {
			if err != nil {
				return false
			}
			err = writeItem(file, id, value)
			return err == nil
		},
		func() error {
			return err
		}
}

func exportCluster(cluster storage.Cluster, services []storage.Service, filename string) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	mw, me := newMonitorWrite(file)

	// dump cluster
	if succ := mw(CLUSTER_NAME, cluster.Name) &&
		mw(CLUSTER_DESCRIPTION, cluster.Description) &&
		mw(CLUSTER_CREATETIME, cluster.CreateTime.Format("2006-01-02 15:04:05")) &&
		mw(CLUSTER_TOPOLOGY, cluster.Topology); !succ {
		return me()
	}

	// dump service
	for _, service := range services {
		value := fmt.Sprintf("%s %s", service.Id, service.ContainerId)
		if succ := mw(SERVICE, value); !succ {
			return me()
		}
	}

	return nil
}

func runExport(curveadm *cli.CurveAdm, options exportOptions) error {
	name := options.name
	storage := curveadm.Storage()
	clusters, err := storage.GetClusters(name)
	if err != nil {
		log.Error("GetClusters", log.Field("error", err))
		return err
	} else if len(clusters) == 0 {
		return fmt.Errorf("cluster %s not exist", name)
	} else if services, err := storage.GetServices(clusters[0].Id); err != nil {
		log.Error("GetServices", log.Field("error", err))
		return err
	} else if err = exportCluster(clusters[0], services, options.outfile); err != nil {
		return err
	}

	curveadm.WriteOut("Export cluster '%s' to '%s' success\n", name, options.outfile)
	return nil
}

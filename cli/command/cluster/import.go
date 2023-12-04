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

	"github.com/opencurve/curveadm/cli/cli"
	"github.com/opencurve/curveadm/internal/storage"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/log/zaplog"
	"github.com/spf13/cobra"
)

const (
	MAX_VALUE_BYETS = 1024 * 1024 // 1MB
)

var (
	importExample = `Examples:
  $ curveadm cluster import my-cluster                     # Import cluster 'my-cluster' with curveadm.db
  $ curveadm cluster import my-cluster -f /path/to/dbfile  # Import cluster 'my-cluster' with specified database file`
)

type importOptions struct {
	name   string
	dbfile string
}

func NewImportCommand(curveadm *cli.CurveAdm) *cobra.Command {
	var options importOptions

	cmd := &cobra.Command{
		Use:     "import CLUSTER [OPTIONS]",
		Short:   "Import cluster",
		Args:    utils.ExactArgs(1),
		Example: importExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			options.name = args[0]
			return runImport(curveadm, options)
		},
		DisableFlagsInUseLine: true,
	}

	flags := cmd.Flags()
	flags.StringVarP(&options.dbfile, "database", "f", "curveadm.db", "Specify the path of database file")

	return cmd
}

func readDB(filepath, name string) (*storage.Cluster, []storage.Service, error) {
	dbUrl := fmt.Sprintf("sqlite://%s", filepath)
	s, err := storage.NewStorage(dbUrl)
	if err != nil {
		return nil, nil, err
	}

	clusters, err := s.GetClusters(name)
	if err != nil {
		return nil, nil, err
	} else if len(clusters) == 0 {
		return nil, nil, fmt.Errorf("cluster '%s' not found", name)
	} else if len(clusters) > 1 {
		return nil, nil, fmt.Errorf("cluster '%s' is duplicate", name)
	}

	cluster := clusters[0]
	services, err := s.GetServices(cluster.Id)
	if err != nil {
		return nil, nil, err
	}
	return &cluster, services, nil
}

func importCluster(storage *storage.Storage, dbfile, name string) error {
	// read database file
	cluster, services, err := readDB(dbfile, name)
	if err != nil {
		return err
	}

	// insert cluster
	err = storage.InsertCluster(name, cluster.UUId, cluster.Description, cluster.Topology, cluster.Type)
	if err != nil {
		return err
	}

	// insert service
	clusters, err := storage.GetClusters(name)
	if err != nil {
		return err
	}
	clusterId := clusters[0].Id
	for _, service := range services {
		err := storage.InsertService(clusterId, service.Id, service.ContainerId)
		if err != nil {
			return err
		}
	}
	return nil
}

func runImport(curveadm *cli.CurveAdm, options importOptions) error {
	name := options.name
	storage := curveadm.Storage()
	clusters, err := storage.GetClusters(name)
	if err != nil {
		zaplog.Error("GetClusters", zaplog.Field("error", err))
		return err
	} else if len(clusters) != 0 { // TODO: let user enter a new cluster name
		return fmt.Errorf("cluster %s already exist", name)
	} else if err := importCluster(storage, options.dbfile, name); err != nil {
		return err
	}

	curveadm.WriteOut("Cluster '%s' imported\n", name)
	return nil
}

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
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

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
  $ curveadm cluster import my-cluster                     # Import a cluster named 'my-cluster' 
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

func readItem(file *os.File) (int, string, error) {
	// key: --- $ID $LENGTH
	buffer := []byte{}
	bytes := make([]byte, 1)
	for {
		_, err := file.Read(bytes)
		if err == io.EOF {
			if len(buffer) == 0 {
				return 0, "", io.EOF
			} else {
				return 0, "", fmt.Errorf("invalid curveadm database, line: %s", string(buffer))
			}
		} else if err != nil {
			return 0, "", err
		} else if bytes[0] == '\n' {
			break
		}
		buffer = append(buffer, bytes[0])
	}

	key := string(buffer)
	pattern := regexp.MustCompile("^--- ([0-9]+) ([0-9]+)$")
	mu := pattern.FindStringSubmatch(key)
	if len(mu) == 0 {
		return 0, "", fmt.Errorf("invalid curveadm database, line: %s", key)
	}

	id, _ := strconv.Atoi(mu[1])
	nbytes, _ := strconv.Atoi(mu[2])
	if nbytes > MAX_VALUE_BYETS {
		return 0, "", fmt.Errorf("too big value int curveadm database")
	}

	// value
	bytes = make([]byte, nbytes)
	nread, err := file.Read(bytes)
	if err == io.EOF || nread != nbytes {
		return 0, "", fmt.Errorf("value broken in database")
	} else if err != nil {
		return 0, "", err
	}

	return id, string(bytes[:nbytes-1]), nil
}

func readDatabase(filename string) (storage.Cluster, []storage.Service, error) {
	cluster := storage.Cluster{}
	services := []storage.Service{}
	file, err := os.Open(filename)
	if err != nil {
		return cluster, services, err
	}
	defer file.Close()

	for {
		id, value, err := readItem(file)
		if err == io.EOF {
			break
		} else if err != nil {
			return cluster, services, err
		}

		switch id {
		case CLUSTER_DESCRIPTION:
			cluster.Description = value
		case CLUSTER_TOPOLOGY:
			cluster.Topology = value
		case SERVICE:
			items := strings.Split(value, " ")
			if len(items) != 2 {
				return cluster, services, fmt.Errorf("invalid service, line: %s", value)
			}
			service := storage.Service{
				Id:          items[0],
				ContainerId: items[1],
			}
			services = append(services, service)
		}
	}

	return cluster, services, nil
}

func importCluster(storage *storage.Storage, name, dbfile string) error {
	// read database file
	cluster, services, err := readDatabase(dbfile)
	if err != nil {
		return err
	}

	// insert cluster
	err = storage.InsertCluster(name, cluster.Description, cluster.Topology)
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
		if err := storage.InsertService(clusterId, service.Id, service.ContainerId); err != nil {
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
	} else if len(clusters) != 0 {
		return fmt.Errorf("cluster %s already exist", name)
	} else if err := importCluster(storage, name, options.dbfile); err != nil {
		storage.DeleteCluster(name)
		return err
	}

	curveadm.WriteOut("Cluster '%s' imported\n", name)
	return nil
}

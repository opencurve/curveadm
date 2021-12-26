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

package topology

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/opencurve/curveadm/pkg/variable"
)

const (
	SELECT_LISTEN_PORT = iota
	SELECT_LISTEN_CLIENT_PORT
	SELECT_LISTEN_DUMMY_PORT
)

type Var struct {
	name     string
	kind     []string // kind limit for register variable
	role     []string // role limit for register variable
	resolved bool
}

/*
 * built-in variables:
 *
 * service:
 *   ${user}                      "curve"
 *   ${prefix}                    "/curvebs/{etcd,mds,chunkserver}"
 *   ${service_id}                "c690bde11d1a"
 *   ${service_role}              "mds"
 *   ${service_host}              "10.0.0.1"
 *   ${service_host_sequence}     "1"
 *   ${service_replica_sequence}  "1"
 *   ${service_addr}              "10.0.0.1"
 *   ${service_port}              "6666"
 *   ${service_client_port}       "2379" (etcd)
 *   ${service_dummy_port}        "6667" (mds)
 *   ${service_external_addr}     "10.0.10.1" (chunkserver/metaserver)
 *   ${log_dir}                   "/data/logs"
 *   ${data_dir}                  "/data"
 *   ${random_uuid}               "6fa8f01c411d7655d0354125c36847bb"
 *
 * cluster:
 *   ${cluster_etcd_http_addr}    "etcd1=http://10.0.10.1:2380,etcd2=http://10.0.10.2:2380,etcd3=http://10.0.10.3:2380"
 *   ${cluster_etcd_addr}         "10.0.10.1:2380,10.0.10.2:2380,10.0.10.3:2380"
 *   ${cluster_mds_addr}          "10.0.10.1:6666,10.0.10.2:6666,10.0.10.3:6666"
 *   ${cluster_mds_dummy_addr}    "10.0.10.1:6667,10.0.10.2:6667,10.0.10.3:6667"
 *   ${cluster_mds_dummy_port}    "6667,6668,6669"
 *   ${cluster_chunkserver_addr}  "10.0.10.1:6800,10.0.10.2:6800,10.0.10.3:6800"
 *   ${cluster_metaserver_addr}   "10.0.10.1:6701,10.0.10.2:6701,10.0.10.3:6701"
 */
var (
	serviceVars = []Var{
		{name: "user"},
		{name: "prefix"},
		{name: "service_id"},
		{name: "service_role"},
		{name: "service_host"},
		{name: "service_host_sequence"},
		{name: "service_replica_sequence"},
		{name: "service_addr"},
		{name: "service_port"},
		{name: "service_client_port", role: []string{ROLE_ETCD}},
		{name: "service_dummy_port", role: []string{ROLE_MDS}},
		{name: "service_external_addr", role: []string{ROLE_CHUNKSERVER, ROLE_METASERVER}},
		{name: "log_dir"},
		{name: "data_dir"},
		{name: "random_uuid"},
	}

	// NOTE: we don't support cluster variable exist in topology
	clusterVars = []Var{
		{name: "cluster_etcd_http_addr"},
		{name: "cluster_etcd_addr"},
		{name: "cluster_mds_addr"},
		{name: "cluster_mds_dummy_addr"},
		{name: "cluster_mds_dummy_port"},
		{name: "cluster_chunkserver_addr", kind: []string{KIND_CURVEBS}},
		{name: "cluster_metaserver_addr", kind: []string{KIND_CURVEFS}},
	}
)

func skip(dc *DeployConfig, v Var) bool {
	role := dc.GetRole()
	kind := dc.GetKind()
	if len(v.kind) != 0 && !utils.Slice2Map(v.kind)[kind] {
		return true
	} else if len(v.role) != 0 && !utils.Slice2Map(v.role)[role] {
		return true
	}

	return false
}

func addVariables(dcs []*DeployConfig, idx int, vars []Var) error {
	dc := dcs[idx]
	for _, v := range vars {
		if skip(dc, v) == true {
			continue
		}

		err := dc.GetVariables().Register(variable.Variable{
			Name:  v.name,
			Value: getValue(v.name, dcs, idx),
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func AddServiceVariables(dcs []*DeployConfig, idx int) error {
	return addVariables(dcs, idx, serviceVars)
}

func AddClusterVariables(dcs []*DeployConfig, idx int) error {
	return addVariables(dcs, idx, clusterVars)
}

/*
 * interface for get variable value
 */
func joinEtcdPeer(dcs []*DeployConfig) string {
	peers := []string{}
	for _, dc := range dcs {
		if dc.GetRole() != ROLE_ETCD {
			continue
		}

		hostSequence := dc.GetHostSequence()
		replicaSquence := dc.GetReplicaSequence()
		peerHost := dc.GetListenIp()
		peerPort := dc.GetListenPort()
		peer := fmt.Sprintf("etcd%d%d=http://%s:%d", hostSequence, replicaSquence, peerHost, peerPort)
		peers = append(peers, peer)
	}
	return strings.Join(peers, ",")
}

func joinPeer(dcs []*DeployConfig, selectRole string, selectPort int) string {
	peers := []string{}
	for _, dc := range dcs {
		if dc.GetRole() != selectRole {
			continue
		}

		peerHost := dc.GetListenIp()
		peerPort := dc.GetListenPort()
		if selectPort == SELECT_LISTEN_CLIENT_PORT {
			peerPort = dc.GetListenClientPort()
		} else if selectPort == SELECT_LISTEN_DUMMY_PORT {
			peerPort = dc.GetListenDummyPort()
		}

		peer := fmt.Sprintf("%s:%d", peerHost, peerPort)
		peers = append(peers, peer)
	}
	return strings.Join(peers, ",")
}

func joinMDSDummyPort(dcs []*DeployConfig) string {
	ports := []string{}
	for _, dc := range dcs {
		if dc.GetRole() == ROLE_MDS {
			ports = append(ports, strconv.Itoa(dc.GetListenDummyPort()))
		}
	}
	return strings.Join(ports, ",")
}

func getValue(name string, dcs []*DeployConfig, idx int) string {
	dc := dcs[idx]
	switch name {
	case "user":
		return dc.GetUser()
	case "prefix":
		return dc.GetProjectLayout().ServiceRootDir
	case "service_id":
		return dc.GetId()
	case "service_role":
		return dc.GetRole()
	case "service_host":
		return dc.GetHost()
	case "service_host_sequence":
		return strconv.Itoa(dc.GetHostSequence())
	case "service_replica_sequence":
		return strconv.Itoa(dc.GetReplicaSequence())
	case "service_addr":
		return utils.Atoa(dc.get(CONFIG_LISTEN_IP))
	case "service_port":
		return utils.Atoa(dc.get(CONFIG_LISTEN_PORT))
	case "service_client_port": // etcd
		return utils.Atoa(dc.get(CONFIG_LISTEN_CLIENT_PORT))
	case "service_dummy_port": // mds
		return utils.Atoa(dc.get(CONFIG_LISTEN_DUMMY_PORT))
	case "service_external_addr": // chunkserver, metaserver
		return utils.Atoa(dc.get(CONFIG_LISTEN_EXTERNAL_IP))
	case "log_dir":
		return dc.GetProjectLayout().ServiceLogDir
	case "data_dir":
		return dc.GetProjectLayout().ServiceDataDir
	case "cluster_etcd_http_addr":
		return joinEtcdPeer(dcs)
	case "cluster_etcd_addr":
		return joinPeer(dcs, ROLE_ETCD, SELECT_LISTEN_CLIENT_PORT)
	case "cluster_mds_addr":
		return joinPeer(dcs, ROLE_MDS, SELECT_LISTEN_PORT)
	case "cluster_mds_dummy_addr":
		return joinPeer(dcs, ROLE_MDS, SELECT_LISTEN_DUMMY_PORT)
	case "cluster_mds_dummy_port":
		return joinMDSDummyPort(dcs)
	case "cluster_chunkserver_addr":
		return joinPeer(dcs, ROLE_CHUNKSERVER, SELECT_LISTEN_PORT)
	case "cluster_metaserver_addr":
		return joinPeer(dcs, ROLE_METASERVER, SELECT_LISTEN_PORT)
	}

	return ""
}

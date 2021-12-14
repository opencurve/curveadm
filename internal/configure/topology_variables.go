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

package configure

import (
	"fmt"
	"strconv"
	"strings"
)

/*
 * built-in variables:
 *
 * service:
 *   ${prefix}                   "/usr/local/curvefs/{etcd,mds,metaserver}"
 *   ${user}                     "curve"
 *   ${service_id}               "1_mds_10.0.0.1_1"
 *   ${service_sequence}         "1"
 *   ${service_role}             "mds"
 *   ${service_host}             "10.0.0.1"
 *   ${service_addr}             "10.0.0.1"
 *   ${service_port}             "6666"
 *   ${service_client_port}      "2379" (etcd)
 *   ${service_dummy_port}       "6667" (mds)
 *   ${service_external_addr}    "10.0.10.1" (metaserver)
 *   ${log_dir}                  "${prefix}/logs"
 *   ${data_dir}                 "${prefix}/data"
 *
 * cluster:
 *   ${cluster_etcd_http_addr}   "etcd1=http://10.0.10.1:2380,etcd2=http://10.0.10.2:2380,etcd3=http://10.0.10.3:2380"
 *   ${cluster_etcd_addr}        "10.0.10.1:2380,10.0.10.2:2380,10.0.10.3:2380"
 *   ${cluster_mds_addr}         "10.0.10.1:6666,10.0.10.2:6666,10.0.10.3:6666"
 *   ${cluster_mds_dummy_addr}   "10.0.10.1:6766,10.0.10.2:6766,10.0.10.3:6766"
 *   ${cluster_metaserver_addr}  "10.0.10.1:6701,10.0.10.2:6701,10.0.10.3:6701"
 */
var (
	commonVariables = []Variable{
		{"prefix", "", "", false},
		{"user", "", "", false},
		{"service_id", "", "", false},
		{"service_sequence", "", "", false},
		{"service_role", "", "", false},
		{"service_host", "", "", false},
		{"service_addr", "", "", false},
		{"service_port", "", "", false},
		{"log_dir", "", "", false},
		{"data_dir", "", "", false},
	}

	serviceVariables = map[string][]Variable{
		ROLE_ETCD: []Variable{
			Variable{"service_client_port", "", "", false},
		},
		ROLE_MDS: []Variable{
			Variable{"service_dummy_port", "", "", false},
		},
		ROLE_METASERVER: []Variable{
			Variable{"service_external_addr", "", "", false},
		},
	}

	// NOTE: we don't support cluster variable exist in topology
	clusterVariables = []Variable{
		{"cluster_etcd_http_addr", "", "", true},
		{"cluster_etcd_addr", "", "", true},
		{"cluster_mds_addr", "", "", true},
		{"cluster_mds_dummy_addr", "", "", true},
		{"cluster_metaserver_addr", "", "", true},
	}
)

func joinEtcdPeer(dcs []*DeployConfig) string {
	peers := []string{}
	for _, dc := range dcs {
		if dc.GetRole() != ROLE_ETCD {
			continue
		}
		peerHost := dc.GetConfig(KEY_LISTEN_IP)
		peerPort := dc.GetConfig(KEY_LISTEN_PORT)
		peer := fmt.Sprintf("etcd%d=http://%s:%s", dc.GetSequence(), peerHost, peerPort)
		peers = append(peers, peer)
	}
	return strings.Join(peers, ",")
}

func joinPeer(dcs []*DeployConfig, role string) string {
	peers := []string{}
	for _, dc := range dcs {
		if dc.GetRole() != role {
			continue
		}
		peerHost := dc.GetConfig(KEY_LISTEN_IP)
		peerPort := dc.GetConfig(KEY_LISTEN_PORT)
		if role == ROLE_ETCD {
			peerPort = dc.GetConfig(KEY_LISTEN_CLIENT_PORT)
		}
		peer := fmt.Sprintf("%s:%s", peerHost, peerPort)
		peers = append(peers, peer)
	}
	return strings.Join(peers, ",")
}

func joinMdsDummyPeer(dcs []*DeployConfig) string {
	peers := []string{}
	for _, dc := range dcs {
		if dc.GetRole() != ROLE_MDS {
			continue
		}
		peerHost := dc.GetConfig(KEY_LISTEN_IP)
		peerPort := dc.GetConfig(KEY_LISTEN_DUMMY_PORT)
		peer := fmt.Sprintf("%s:%s", peerHost, peerPort)
		peers = append(peers, peer)
	}
	return strings.Join(peers, ",")
}

func getValue(name string, dcs []*DeployConfig, idx int) string {
	dc := dcs[idx]
	switch name {
	case "user":
		return dc.GetConfig(KEY_USER)
	case "prefix":
		return dc.GetServicePrefix()
	case "service_id":
		return dc.GetId()
	case "service_sequence":
		return strconv.Itoa(dc.GetSequence())
	case "service_role":
		return dc.GetRole()
	case "service_host":
		return dc.GetHost()
	case "service_addr":
		return dc.GetConfig(KEY_LISTEN_IP)
	case "service_port":
		return dc.GetConfig(KEY_LISTEN_PORT)
	case "service_client_port": // etcd
		return dc.GetConfig(KEY_LISTEN_CLIENT_PORT)
	case "service_dummy_port": // mds
		return dc.GetConfig(KEY_LISTEN_DUMMY_PORT)
	case "service_external_addr": // metaserver
		return dc.GetConfig(KEY_LISTEN_EXTERNAL_IP)
	case "log_dir":
		return dc.GetConfig(KEY_LOG_DIR)
	case "data_dir":
		return dc.GetConfig(KEY_DATA_DIR)
	case "cluster_etcd_http_addr":
		return joinEtcdPeer(dcs)
	case "cluster_etcd_addr":
		return joinPeer(dcs, ROLE_ETCD)
	case "cluster_mds_addr":
		return joinPeer(dcs, ROLE_MDS)
	case "cluster_mds_dummy_addr":
		return joinMdsDummyPeer(dcs)
	case "cluster_metaserver_addr":
		return joinPeer(dcs, ROLE_METASERVER)
	}

	return ""
}

func addVariables(dcs []*DeployConfig, idx int, vars []Variable) error {
	dc := dcs[idx]
	for _, v := range vars {
		v.Value = getValue(v.Name, dcs, idx)
		err := dc.GetVariables().Register(v)
		if err != nil {
			return err
		}
	}

	return nil
}

func addServiceVariables(dcs []*DeployConfig, idx int) error {
	role := dcs[idx].GetRole()
	vars := append(commonVariables, serviceVariables[role]...)
	return addVariables(dcs, idx, vars)
}

func addClusterVariables(dcs []*DeployConfig, idx int) error {
	vars := clusterVariables
	return addVariables(dcs, idx, vars)
}

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
 * Created Date: 2021-12-23
 * Author: Jingli Chen (Wine93)
 */

package pool

import (
	"encoding/json"
	"fmt"

	"github.com/opencurve/curveadm/internal/configure/topology"
)

const (
	KIND_CURVEBS     = topology.KIND_CURVEBS
	KIND_CURVEFS     = topology.KIND_CURVEFS
	ROLE_CHUNKSERVER = topology.ROLE_CHUNKSERVER
	ROLE_METASERVER  = topology.ROLE_METASERVER

	DEFAULT_REPLICAS_PER_COPYSET = 3
	DEFAULT_ZONES_PER_POOL       = 3
	DEFAULT_TYPE                 = 0
	DEFAULT_SCATTER_WIDTH        = 0
)

type (
	LogicalPool struct {
		Name         string `json:"name"`
		Replicas     int    `json:"replicasnum"`
		Zones        int    `json:"zonenum"`
		Copysets     int    `json:"copysetnum"`
		Type         int    `json:"type"`         // curvebs
		ScatterWidth int    `json:"scatterwidth"` // curvebs
		PhysicalPool string `json:"physicalpool"` // curvebs
	}

	Server struct {
		Name         string `json:"name"`
		InternalIp   string `json:"internalip"`
		InternalPort int    `json:"internalport"`
		ExternalIp   string `json:"externalip"`
		ExternalPort int    `json:"externalport"`
		Zone         string `json:"zone"`
		PhysicalPool string `json:"physicalpool,omitempty"` // curvebs
		Pool         string `json:"pool,omitempty"`         // curvefs
	}

	CurveClusterTopo struct {
		Servers      []Server      `json:"servers"`
		LogicalPools []LogicalPool `json:"logicalpools,omitempty"` // curvebs
		Pools        []LogicalPool `json:"pools,omitempty"`        // curvefs
	}
)

/*
 * curvebs_cluster_topo:
 *   servers:
 *     - name: server1
 *       internalip: 127.0.0.1
 *       internalport: 16701
 *       externalip: 127.0.0.1
 *       externalport: 16701
 *       zone: zone1
 *       physicalpool: pool1
 *    ...
 *   logicalpools:
 *     - name: pool1
 *       physicalpool: pool1
 *       replicasnum: 3
 *       copysetnum: 100
 *       zonenum: 3
 *       type: 0
 *       scatterwidth: 0
 *     ...
 *
 *
 * curvefs_cluster_topo:
 *   servers:
 *     - name: server1
 *       internalip: 127.0.0.1
 *       internalport: 16701
 *       externalip: 127.0.0.1
 *       externalport: 16701
 *       zone: zone1
 *       pool: pool1
 *     ...
 *   pools:
 *     - name: pool1
 *       replicasnum: 3
 *       copysetnum: 100
 *       zonenum: 3
 */
func genNextZone(zones int) func() string {
	idx := 0
	return func() string {
		idx++
		return fmt.Sprintf("zone%d", (idx-1)%zones+1)
	}
}

func createLogicalPool(dcs []*topology.DeployConfig, logicalPool string) (LogicalPool, []Server) {
	copysets := 0
	servers := []Server{}
	zones := DEFAULT_ZONES_PER_POOL
	nextZone := genNextZone(zones)
	physicalPool := logicalPool
	kind := dcs[0].GetKind()
	for _, dc := range dcs {
		role := dc.GetRole()
		if (role == ROLE_CHUNKSERVER && kind == KIND_CURVEBS) ||
			(role == ROLE_METASERVER && kind == KIND_CURVEFS) {
			server := Server{
				Name:         fmt.Sprintf("%s_%s_%d", dc.GetHost(), dc.GetName(), dc.GetReplicaSequence()),
				InternalIp:   dc.GetListenIp(),
				InternalPort: dc.GetListenPort(),
				ExternalIp:   dc.GetListenExternalIp(),
				ExternalPort: dc.GetListenPort(),
				Zone:         nextZone(),
			}
			if kind == KIND_CURVEBS {
				server.PhysicalPool = physicalPool
			} else {
				server.Pool = logicalPool
			}
			copysets += dc.GetCopysets()
			servers = append(servers, server)
		}
	}

	// copysets
	copysets = (int)(copysets / DEFAULT_REPLICAS_PER_COPYSET)
	if copysets == 0 {
		copysets = 1
	}

	// logical pool
	lpool := LogicalPool{
		Name:     logicalPool,
		Copysets: copysets,
		Zones:    zones,
		Replicas: DEFAULT_REPLICAS_PER_COPYSET,
	}
	if kind == KIND_CURVEBS {
		lpool.ScatterWidth = DEFAULT_SCATTER_WIDTH
		lpool.Type = DEFAULT_TYPE
		lpool.PhysicalPool = physicalPool
	}

	return lpool, servers
}

func GenerateClusterPool(data string) (string, error) {
	dcs, err := topology.ParseTopology(data)
	if err != nil {
		return "", err
	}

	lpool, servers := createLogicalPool(dcs, "defaultPool")
	topo := CurveClusterTopo{Servers: servers}
	if dcs[0].GetKind() == KIND_CURVEBS {
		topo.LogicalPools = []LogicalPool{lpool}
	} else {
		topo.Pools = []LogicalPool{lpool}
	}

	bytes, err := json.Marshal(&topo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

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

package configure

import (
	"fmt"
	"sort"

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
	MigrateServer struct {
		From *topology.DeployConfig
		To   *topology.DeployConfig
	}
	Poolset struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
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
		Poolset      string `json:"poolset,omitempty"`      // curvebs
	}

	CurveClusterTopo struct {
		Servers      []Server      `json:"servers"`
		LogicalPools []LogicalPool `json:"logicalpools,omitempty"` // curvebs
		Pools        []LogicalPool `json:"pools,omitempty"`        // curvefs
		Poolsets     []Poolset     `json:"poolsets"`               // curvebs
		NPools       int           `json:"npools"`
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
 *       poolset: ssd_poolset1
 *    ...
 *   logicalpools:
 *     - name: pool1
 *       physicalpool: pool1
 *       replicasnum: 3
 *       copysetnum: 100
 *       zonenum: 3
 *       type: 0
 *       scatterwidth: 0
 *   poolsets:
 *     - name: ssd_poolset1
 *       type: ssd
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

// we should sort the "dcs" for generate correct zone number
func SortDeployConfigs(dcs []*topology.DeployConfig) {
	sort.Slice(dcs, func(i, j int) bool {
		dc1, dc2 := dcs[i], dcs[j]
		if dc1.GetRole() == dc2.GetRole() {
			if dc1.GetHostSequence() == dc2.GetHostSequence() {
				return dc1.GetReplicasSequence() < dc2.GetReplicasSequence()
			}
			return dc1.GetHostSequence() < dc2.GetHostSequence()
		}
		return dc1.GetRole() < dc2.GetRole()
	})
}

func formatName(dc *topology.DeployConfig) string {
	return fmt.Sprintf("%s_%s_%d", dc.GetHost(), dc.GetName(), dc.GetReplicasSequence())
}

func createLogicalPool(dcs []*topology.DeployConfig, logicalPool string, poolset string) (LogicalPool, []Server) {
	var zone string
	copysets := 0
	servers := []Server{}
	zones := DEFAULT_ZONES_PER_POOL
	nextZone := genNextZone(zones)
	physicalPool := logicalPool
	kind := dcs[0].GetKind()
	SortDeployConfigs(dcs)
	for _, dc := range dcs {
		role := dc.GetRole()
		if (role == ROLE_CHUNKSERVER && kind == KIND_CURVEBS) ||
			(role == ROLE_METASERVER && kind == KIND_CURVEFS) {
			if dc.GetParentId() == dc.GetId() {
				zone = nextZone()
			}

			// NOTE: if we deploy chunkservers with replica feature
			// and the value of replica greater than 1, we should
			// set internal port and external port to 0 for let MDS
			// attribute them as services on the same machine.
			// see issue: https://github.com/opencurve/curve/issues/1441
			internalPort := dc.GetListenPort()
			externalPort := dc.GetListenExternalPort()
			if dc.GetReplicas() > 1 {
				internalPort = 0
				externalPort = 0
			}

			server := Server{
				Name:         formatName(dc),
				InternalIp:   dc.GetListenIp(),
				InternalPort: internalPort,
				ExternalIp:   dc.GetListenExternalIp(),
				ExternalPort: externalPort,
				Zone:         zone,
			}
			if kind == KIND_CURVEBS {
				server.PhysicalPool = physicalPool
				server.Poolset = poolset
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

func generateClusterPool(dcs []*topology.DeployConfig, poolName string, poolset, diskType string) CurveClusterTopo {
	lpool, servers := createLogicalPool(dcs, poolName, poolset)
	topo := CurveClusterTopo{Servers: servers, NPools: 1}
	if dcs[0].GetKind() == KIND_CURVEBS {
		topo.LogicalPools = []LogicalPool{lpool}
		//Poolset
		poolset := Poolset{
			Name: poolset,
			Type: diskType,
		}
		//topo.Poolsets = []Poolset{poolset}
		topo.Poolsets = append(topo.Poolsets, poolset)
	} else {
		topo.Pools = []LogicalPool{lpool}
	}
	return topo
}

func ScaleOutClusterPool(old *CurveClusterTopo, dcs []*topology.DeployConfig, poolset, diskType string) {
	npools := old.NPools
	topo := generateClusterPool(dcs, fmt.Sprintf("pool%d", npools+1), poolset, diskType)
	if dcs[0].GetKind() == KIND_CURVEBS {
		for _, pool := range topo.LogicalPools {
			old.LogicalPools = append(old.LogicalPools, pool)
		}
		for _, newPst := range topo.Poolsets {
			isExist := false
			for _, oldPst := range old.Poolsets {
				if oldPst.Name == newPst.Name {
					isExist = true
				}
			}
			if !isExist {
				old.Poolsets = append(old.Poolsets, newPst)
			}
		}
	} else {
		for _, pool := range topo.Pools {
			old.Pools = append(old.Pools, pool)
		}
	}
	for _, server := range topo.Servers {
		old.Servers = append(old.Servers, server)
	}
	old.NPools = old.NPools + 1
}

func MigrateClusterServer(old *CurveClusterTopo, migrates []*MigrateServer) {
	m := map[string]*topology.DeployConfig{} // key: from.Name, value: to.DeployConfig
	for _, migrate := range migrates {
		m[formatName(migrate.From)] = migrate.To
	}

	for i, server := range old.Servers {
		dc, ok := m[server.Name]
		if !ok {
			continue
		}

		server.InternalIp = dc.GetListenIp()
		server.ExternalIp = dc.GetListenExternalIp()
		server.Name = formatName(dc)
		if server.InternalPort != 0 && server.ExternalPort != 0 {
			server.InternalPort = dc.GetListenPort()
			server.ExternalPort = dc.GetListenExternalPort()
		}
		old.Servers[i] = server
	}
}

func GenerateDefaultClusterPool(dcs []*topology.DeployConfig, poolset, diskType string) (topo CurveClusterTopo, err error) {
	topo = generateClusterPool(dcs, "pool1", poolset, diskType)
	return
}

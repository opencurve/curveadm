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
	"encoding/json"
	"fmt"
	"strconv"
)

/*
 * curvefs_topology:
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
type (
	Server struct {
		Name         string `json:"name"`
		InternalIp   string `json:"internalip"`
		InternalPort int    `json:"internalport"`
		ExternalIp   string `json:"externalip"`
		ExternalPort int    `json:"externalport"`
		Zone         string `json:"zone"`
		Pool         string `json:"pool"`
	}

	Pool struct {
		Name        string `json:"name"`
		ReplicasNum int    `json:"replicasnum"`
		CopysetNum  int    `json:"copysetnum"`
		ZonNum      int    `json:"zonenum"`
	}

	CurveFSTopology struct {
		Servers []Server `json:"servers"`
		Pools   []Pool   `json:"pools"`
	}
)

// TODO(@Wine93): support generate topology by customer
func GenerateCurveFSTopology(data string) (string, error) {
	dcs, err := ParseTopology(data)
	if err != nil {
		return "", err
	}

	idx := 1
	servers := []Server{}
	pools := []Pool{{"pool1", 3, 100, 3}}
	for _, dc := range dcs {
		if dc.GetRole() != ROLE_METASERVER {
			continue
		}

		ip := dc.GetConfig(KEY_LISTEN_IP)
		port, _ := strconv.Atoi(dc.GetConfig(KEY_LISTEN_PORT))
		server := Server{dc.GetId(), ip, port, ip, port,
			fmt.Sprintf("zone%d", (idx-1)%3+1), "pool1"}
		servers = append(servers, server)
		idx += 1
	}

	topo := CurveFSTopology{servers, pools}
	bytes, err := json.Marshal(&topo)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

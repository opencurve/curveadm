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
	"bytes"
	"fmt"
	"strings"

	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

type (
	Deploy struct {
		Host   string                 `mapstructure:"host"`
		Name   string                 `mapstructure:"name"`
		Config map[string]interface{} `mapstructure:"config"`
	}

	Service struct {
		Config map[string]interface{} `mapstructure:"config"`
		Deploy []Deploy               `mapstructure:"deploy"`
	}

	Topology struct {
		Kind string `mapstructure:"kind"`

		Global map[string]interface{} `mapstructure:"global"`

		EtcdServices        Service `mapstructure:"etcd_services"`
		MdsServices         Service `mapstructure:"mds_services"`
		MetaserverServices  Service `mapstructure:"metaserver_services"`
		ChunkserverServices Service `mapstructure:"chunkserver_services"`

		Pools []Pool `mapstructure:"pools"`
	}
)

func merge(parent, child map[string]interface{}, deep int) {
	for k, v := range parent {
		if child[k] == nil {
			child[k] = v
		} else if k == KEY_VARIABLE && deep < 2 &&
			!utils.IsString(v) && !utils.IsInt(v) { // variable map
			subparent := parent[k].(map[string]interface{})
			subchild := child[k].(map[string]interface{})
			merge(subparent, subchild, deep+1)
		}
	}
}

func newIfNil(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		config = map[string]interface{}{}
	}
	return config
}

func ParseTopology(data string) ([]*DeployConfig, error) {
	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigType("yaml")
	err := parser.ReadConfig(bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, err
	}

	topology := &Topology{}
	err = parser.Unmarshal(topology)
	if err != nil {
		return nil, err
	}

	kind := topology.Kind
	if topology.Kind != KIND_CURVEBS ||

	dcs := []*DeployConfig{}
	globalConfig := newIfNil(topology.Global)
	for _, role := range []string{ROLE_ETCD, ROLE_MDS, ROLE_METASERVER, ROLE_CHUNKSERVER} {
		services := Service{}
		switch role {
		case ROLE_ETCD:
			services = topology.EtcdServices
		case ROLE_MDS:
			services = topology.MdsServices
		case ROLE_METASERVER:
			services = topology.MetaserverServices
		case ROLE_CHUNKSERVER:
			services = topology.ChunkserverServices
		}

		// merge global config into services config
		servicesConfig := newIfNil(services.Config)
		merge(globalConfig, servicesConfig, 1)

		for i, deploy := range services.Deploy {
			// merge services config into deploy config
			deployConfig := newIfNil(deploy.Config)
			merge(servicesConfig, deployConfig, 1)

			dc, err := NewDeployConfig(role, deploy.Host, deploy.Name, i+1, deployConfig)
			if err != nil {
				return nil, err
			}
			dcs = append(dcs, dc)
		}
	}

	exist := map[string]bool{}
	for idx, dc := range dcs {
		if err = addServiceVariables(dcs, idx); err != nil {
			return nil, err
		} else if err = dc.Build(); err != nil {
			return nil, err
		} else if exist[dc.GetId()] {
			return nil, fmt.Errorf("service id(%s) is duplicate", dc.GetId())
		}
	}

	for idx, dc := range dcs {
		if err = addClusterVariables(dcs, idx); err != nil {
			return nil, err
		}
		dc.GetVariables().Debug()
	}

	return dcs, nil
}

func ServiceId(clusterId int, dcId string) string {
	return fmt.Sprintf("%d_%s", clusterId, dcId)
}

func ExtractDcId(serviceId string) string {
	items := strings.Split(serviceId, "_")[1:]
	return strings.Join(items, "_")
}

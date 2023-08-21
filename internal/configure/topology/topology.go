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

// __SIGN_BY_WINE93__

package topology

import (
	"bytes"

	"github.com/opencurve/curveadm/internal/errno"
	"github.com/opencurve/curveadm/internal/utils"
	"github.com/spf13/viper"
)

type (
	Deploy struct {
		Host     string                 `mapstructure:"host"`
		Name     string                 `mapstructure:"name"`
		Replica  int                    `mapstructure:"replica"`  // old version
		Replicas int                    `mapstructure:"replicas"` // old version
		Instance int                    `mapstructure:"instance"`
		Config   map[string]interface{} `mapstructure:"config"`
	}

	Service struct {
		Config map[string]interface{} `mapstructure:"config"`
		Deploy []Deploy               `mapstructure:"deploy"`
	}

	Topology struct {
		Kind string `mapstructure:"kind"`

		Global map[string]interface{} `mapstructure:"global"`

		EtcdServices          Service `mapstructure:"etcd_services"`
		MdsServices           Service `mapstructure:"mds_services"`
		MetaserverServices    Service `mapstructure:"metaserver_services"`
		ChunkserverServices   Service `mapstructure:"chunkserver_services"`
		SnapshotcloneServices Service `mapstructure:"snapshotclone_services"`
	}
)

var (
	CURVEBS_ROLES = []string{
		ROLE_ETCD,
		ROLE_MDS,
		ROLE_CHUNKSERVER,
		ROLE_SNAPSHOTCLONE,
	}
	CURVEFS_ROLES = []string{
		ROLE_ETCD,
		ROLE_MDS,
		ROLE_METASERVER,
	}
)

func merge(parent, child map[string]interface{}, deep int) {
	for k, v := range parent {
		if child[k] == nil {
			child[k] = v
		} else if k == CONFIG_VARIABLE.Key() && deep < 2 &&
			!utils.IsString(v) && !utils.IsInt(v) { // variable map
			subparent := parent[k].(map[string]interface{})
			subchild := child[k].(map[string]interface{})
			merge(subparent, subchild, deep+1)
		}
	}
}

func newIfNil(config map[string]interface{}) map[string]interface{} {
	if config == nil {
		return map[string]interface{}{}
	}
	return config
}

func ParseTopology(data string, ctx *Context) ([]*DeployConfig, error) {
	if len(data) == 0 {
		return nil, errno.ERR_EMPTY_CLUSTER_TOPOLOGY
	}

	parser := viper.NewWithOptions(viper.KeyDelimiter("::"))
	parser.SetConfigType("yaml")
	err := parser.ReadConfig(bytes.NewBuffer([]byte(data)))
	if err != nil {
		return nil, errno.ERR_PARSE_TOPOLOGY_FAILED.E(err)
	}

	topology := &Topology{}
	err = parser.Unmarshal(topology)
	if err != nil {
		return nil, errno.ERR_PARSE_TOPOLOGY_FAILED.E(err)
	}

	// check topology kind
	kind := topology.Kind
	roles := []string{}
	if kind == KIND_CURVEBS {
		roles = append(roles, CURVEBS_ROLES...)
	} else if kind == KIND_CURVEFS {
		roles = append(roles, CURVEFS_ROLES...)
	} else {
		return nil, errno.ERR_UNSUPPORT_CLUSTER_KIND
	}

	dcs := []*DeployConfig{}
	globalConfig := newIfNil(topology.Global)
	for _, role := range roles {
		services := Service{}
		switch role {
		case ROLE_ETCD:
			services = topology.EtcdServices
		case ROLE_MDS:
			services = topology.MdsServices
		case ROLE_CHUNKSERVER:
			services = topology.ChunkserverServices
		case ROLE_SNAPSHOTCLONE:
			services = topology.SnapshotcloneServices
		case ROLE_METASERVER:
			services = topology.MetaserverServices
		}

		// merge global config into services config
		servicesConfig := newIfNil(services.Config)
		merge(globalConfig, servicesConfig, 1)

		for hostSequence, deploy := range services.Deploy {
			// merge services config into deploy config
			deployConfig := newIfNil(deploy.Config)
			merge(servicesConfig, deployConfig, 1)

			// create deploy config
			instance := 1
			if deploy.Replicas < 0 {
				return nil, errno.ERR_REPLICAS_REQUIRES_POSITIVE_INTEGER.
					F("replicas: %d", deploy.Replicas)
			} else if deploy.Replica < 0 {
				return nil, errno.ERR_REPLICAS_REQUIRES_POSITIVE_INTEGER.
					F("replica: %d", deploy.Replica)
			} else if deploy.Instance < 0 {
				return nil, errno.ERR_REPLICAS_REQUIRES_POSITIVE_INTEGER.
					F("replica: %d", deploy.Instance)
			} else if deploy.Instance > 0 {
				instance = deploy.Instance
			} else if deploy.Replicas > 0 {
				instance = deploy.Replicas
			} else if deploy.Replica > 0 {
				instance = deploy.Replica
			}

			for replicasSequence := 0; replicasSequence < instance; replicasSequence++ {
				dc, err := NewDeployConfig(ctx, kind,
					role, deploy.Host, deploy.Name, instance,
					hostSequence, replicasSequence, utils.DeepCopy(deployConfig))
				if err != nil {
					return nil, err // already is error code
				}
				dcs = append(dcs, dc)
			}
		}
	}

	// add service variables
	exist := map[string]bool{}
	for idx, dc := range dcs {
		if err = dc.ResolveHost(); err != nil {
			return nil, err
		} else if err = AddServiceVariables(dcs, idx); err != nil {
			return nil, err // already is error code
		} else if err = dc.Build(); err != nil {
			return nil, err // already is error code
		} else if exist[dc.GetId()] {
			// actually the dc.GetId() return configure id
			return nil, errno.ERR_DUPLICATE_SERVICE_ID.
				F("service id: %s", dc.GetId())
		}
	}

	// add cluster variables
	for idx, dc := range dcs {
		if err = AddClusterVariables(dcs, idx); err != nil {
			return nil, err // already is error code
		} else if err = dc.GetVariables().Build(); err != nil {
			return nil, errno.ERR_RESOLVE_VARIABLE_FAILED.E(err)
		}

		dc.GetVariables().Debug()
	}

	return dcs, nil
}

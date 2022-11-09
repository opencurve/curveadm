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
 * Created Date: 2022-06-24
 * Author: Jingli Chen (Wine93)
 */

package configure

import (
	"github.com/opencurve/curveadm/internal/configure/topology"
)

const (
	DEFAULT_CURVEBS_CONTAINER_IMAGE = "opencurvedocker/curvebs-playground:v1.2"
	DEFAULT_CURVEFS_CONTAINER_IMAGE = "opencurvedocker/curvefs-playground:v2.3"
)

type (
	PlaygroundConfig struct {
		Kind           string
		Name           string
		ContainerImage string
		Mountpoint     string

		DeployConfigs []*topology.DeployConfig
		ClientConfig  *ClientConfig
	}
)

func (cfg *PlaygroundConfig) GetKind() string                            { return cfg.Kind }
func (cfg *PlaygroundConfig) GetName() string                            { return cfg.Name }
func (cfg *PlaygroundConfig) GetMointpoint() string                      { return cfg.Mountpoint }
func (cfg *PlaygroundConfig) GetDeployConfigs() []*topology.DeployConfig { return cfg.DeployConfigs }
func (cfg *PlaygroundConfig) GetClientConfig() *ClientConfig             { return cfg.ClientConfig }

func (cfg *PlaygroundConfig) GetContainIamge() string {
	if len(cfg.ContainerImage) > 0 {
		return cfg.ContainerImage
	} else if cfg.Kind == topology.KIND_CURVEBS {
		return DEFAULT_CURVEBS_CONTAINER_IMAGE
	}
	return DEFAULT_CURVEFS_CONTAINER_IMAGE
}

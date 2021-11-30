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
 * Created Date: 2021-11-30
 * Author: Jingli Chen (Wine93)
 */

package configure

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	CURVEBS_TOPOLOGY = `
kind: curvebs
hello: "world"
global:
  user: curve
  ssh_port: 22
  private_key_file: /home/curve/.ssh/id_rsa
  data_dir: /home/curve/curvefs/data/${service_role}
  log_dir: /home/curve/curvefs/logs/${service_role}
  container_image: opencurvedocker/curvebs:latest
  variable:
    machine1: 10.0.1.1
    machine2: 10.0.1.2
    machine3: 10.0.1.3

etcd_services:
  config:
    listen.ip: ${service_host}
    listen.port: 2380
    listen.client_port: 2379
  deploy:
    - host: ${machine1}
    - host: ${machine2}
    - host: ${machine3}

mds_services:
  config:
    listen.ip: ${service_host}
    listen.port: 6700
    listen.dummy_port: 7700
  deploy:
    - host: ${machine1}
    - host: ${machine2}
    - host: ${machine3}

chunkserver_services:
  config:
    listen.ip: ${service_host}
    listen.port: 6701
    metaserver.loglevel: 0
  deploy:
    - host: ${machine1}
    - host: ${machine2}
    - host: ${machine3}
      config:
        metaserver.loglevel: 3
`
)

func TestParseTopology(t *testing.T) {
	assert := assert.New(t)

	dcs, err := ParseTopology(CURVEBS_TOPOLOGY)
	assert.Nil(err)
	assert.Len(dcs, 9)
}

/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-04-23
* Author: wanghai (SeanHai)
 */

package scripts

var PROMETHEUS_YML = `
global:
  scrape_interval: 3s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'prometheus'
    static_configs:
    - targets: ['localhost:%d']

  - job_name: 'curve_metrics'
    file_sd_configs:
    - files: ['target.json']

  - job_name: 'node'
    static_configs:
      - targets: %s
`

var GRAFANA_DATA_SOURCE = `
datasources:
- name: 'Prometheus'
  type: 'prometheus'
  access: 'proxy'
  org_id: 1
  url: 'http://%s:%d'
  is_default: true
  version: 1
  editable: true
`

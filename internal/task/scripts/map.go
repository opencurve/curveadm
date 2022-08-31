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
 * Created Date: 2022-01-10
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: map USER VOLUME CREATE SIZE
 * Example: map curve test true 10
 */
var MAP = `
#!/usr/bin/env bash

g_user=$1
g_volume=$2
g_options=$3
g_stderr=/tmp/__curveadm_map__

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab
curve-nbd map --nbds_max=16 ${g_options} cbd:pool/${g_volume}_${g_user}_ > ${g_stderr} 2>&1
if [ $? -ne 0 ]; then
  cat ${g_stderr}
  exit 1
else
  echo "SUCCESS"
fi
`

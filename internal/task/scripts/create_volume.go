/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-07-31
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: create_volume USER VOLUME SIZE
 * Example: create_volume curve test 10
 */
var CREATE_VOLUME = `
#!/usr/bin/env bash

g_user=$1
g_volume=$2
g_size=$3
g_poolset=$4

output=$(curve_ops_tool create -userName=$g_user -fileName=$g_volume -fileLength=$g_size -poolset=$g_poolset 2>dev/null)
if [ $? -ne 0 ]; then
  if [ "$output" = "CreateFile fail with errCode: 101" ]; then
     echo "EXIST"
  elif echo ${output} | grep -q "kAuthFailed"; then
     echo "AuthFailed"
  elif echo ${output} | grep -q "auth info fail"; then
     echo "AUTH_KEY_NOT_EXIST"
  else
     echo "FAILED"
  fi
else
  echo "SUCCESS"
fi
`

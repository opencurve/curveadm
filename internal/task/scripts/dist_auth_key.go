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
 * Project: CurveAdm
 * Created Date: 2023-07-07
 * Author: caoxianfei1
 */

package scripts

var DIST_AUTH_KEY = `
user=$1
role=$2
key=$3
authkey=$4
toolsUser=$5
toolsRole=$6
toolsKey=$7

curve_ops_tool auth-key-add -user ${user} -role ${role} -key ${key} -authkey ${authkey} &&
curve_ops_tool auth-key-add -user ${toolsUser} -role ${toolsRole} -key ${toolsKey} -authkey ${authkey}
if [ $? -ne 0 ]; then
  exit 1
else 
  echo "SUCCESS"
fi
`

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
 * Created Date: 2022-03-09
 * Author: aspirer
 */

package scripts

/*
 * Usage: no args, just run it in bash
 */
var WAIT_CHUNKSERVERS = `
wait=0
while ((wait<20))
do
    status=$(curve_ops_tool chunkserver-status |grep "offline")
    total=$(echo ${status} | grep -c "total num = 0")
    offline=$(echo ${status} | grep -c "offline = 0")
    if [ ${total} -eq 0 ] && [ ${offline} -eq 1 ]; then
        exit 0
    fi
    sleep 0.5s
    wait=$(expr ${wait} + 1)
done
echo "wait chunkservers online timeout"
exit 1
`

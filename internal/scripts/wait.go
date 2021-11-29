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
 * Created Date: 2021-11-25
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: wait ADDR...
 * Example: wait 10.0.10.1:2379 10.0.10.2:2379
 */
var WAIT = `
[[ -z $(which curl) ]] && apt-get install -y curl
wait=0
while ((wait<20))
do
    for addr in "$@"
    do
        curl $addr -Iso /dev/null
        if [ $? == 0 ]; then
           exit 0
        fi
    done
    sleep 0.5s
    wait=$(expr $wait + 1)
done
exit 1
`

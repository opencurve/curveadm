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
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: target USER VOLUME CREATE SIZE
 * Example: target curve test true 10
 * See Also: https://linux.die.net/man/8/tgtadm
 */

var START_SPDK = `
#!/usr/bin/env bash

g_binarypath="/usr/local/bin/spdk_tgt"
g_cpumask=0x3
g_sockname="/var/tmp/spdk.sock"
g_iscsi_log="/var/log/spdk.log"

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab

process_name="spdk_tgt"

if ps aux | grep -v grep | grep "$process_name" >/dev/null; then
   echo "spdk iscsi_tgt has already been started, now exit!"
   exit 1
fi

nohup ${g_binarypath} -m ${g_cpumask} -r ${g_sockname} > ${g_iscsi_log} 2>&1 &

i=1
while true
do
    result=$(cat $VSDK_LOG_FILE | grep "please run rpc")
    if [[ "$result" != "" ]]; then
        i=0
        break
    else
        sleep 0.1
        i++
    fi
    if [ "$i" -ge 50 ]; then
        break
    fi
done

if [ "$i" -eq 0]; then
    echo "spdk spdk_tgt started success!"
else
    echo "wait for 5s, spdk iscsi_tgt has not launched, please try again!"
    exit 1
fi

exit 0

`

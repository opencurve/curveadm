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
g_create=$3
g_size=$4

if [ $g_create == "true" ]; then
    output=$(curve_ops_tool create -userName=$g_user -fileName=$g_volume -fileLength=$g_size)
    if [ $? -ne 0 ]; then
        if [ "$output" != "CreateFile fail with errCode: 101" ]; then
            echo ${output}
            exit 1
        fi
    fi
fi

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab
curve-nbd map --nbds_max=16 cbd:pool/${g_volume}_${g_user}_
[[ ! -z $(pgrep curve-nbd) ]] && tail --pid=$(pgrep curve-nbd) -f /dev/null
`

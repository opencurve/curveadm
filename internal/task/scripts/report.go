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
 * Created Date: 2021-12-06
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: report KIND UUID ROLE
 * Example: report curvebs abcdef01234567890 metaserver
 */
var REPORT = `
function rematch() {
    local s=$1 regex=$2
    if [[ $s =~ $regex ]]; then
        echo "${BASH_REMATCH[1]}"
    fi
}

function fs_usage() {
    curvefs_tool usage-metadata 2>/dev/null | awk 'BEGIN {
        BYTES["KB"] = 1024
        BYTES["MB"] = BYTES["KB"] * 1024
        BYTES["GB"] = BYTES["MB"] * 1024
        BYTES["TB"] = BYTES["GB"] * 1024
    }
    { 
        if ($0 ~ /all cluster/) {
            printf ("%0.f", $8 * BYTES[$9])
        }
    }'
}

function bs_usage() {
    local message=$(curve_ops_tool space | grep physical)
    local used=$(rematch "$message" "used = ([0-9]+)GB")
    echo $(($used*1024*1024*1024))
}

[[ -z $(which curl) ]] && apt-get install -y curl
g_kind=$1
g_uuid=$2
g_role=$3
g_usage=$(([[ $g_kind = "curvebs" ]] && bs_usage) || fs_usage)
curl -XPOST http://curveadm.aspirer.wang:19302/ \
    -d "kind=$g_kind" \
    -d "uuid=$g_uuid" \
    -d "role=$g_role" \
    -d "usage=$g_usage"
`

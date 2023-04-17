#!/usr/bin/env bash

# Usage: report KIND UUID ROLE
# Example: report curvebs abcdef01234567890 metaserver
# Created Date: 2021-12-06
# Author: Jingli Chen (Wine93)


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

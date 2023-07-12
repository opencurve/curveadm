#!/usr/bin/env bash

# Usage: create_volume USER VOLUME SIZE
# Example: create_volume curve test 10
# Created Date: 2022-07-31
# Author: Jingli Chen (Wine93)


g_user=$1
g_volume=$2
g_size=$3
g_poolset=$4
g_create_opts=(
    "-userName=$g_user"
    "-fileName=$g_volume"
    -fileLength="$g_size"
)
if [ -n "$g_poolset" ]; then
    g_create_opts+=("-poolset=$g_poolset")
fi

output=$(curve_ops_tool create "${g_create_opts[@]}")
if [ "$?" -ne 0 ]; then
    if [ "$output" = "CreateFile fail with errCode: kFileExists" ]; then
        echo "EXIST"
    else
        echo "FAILED"
    fi
else
    echo "SUCCESS"
fi

#!/usr/bin/env bash

# Usage: wait_chunkserver 3
# Created Date: 2022-03-09
# Author: aspirer


g_total="$1"

wait=0
while ((wait<60))
do
    online=$(curve_ops_tool chunkserver-status | sed -nr 's/.*online = ([0-9]+).*/\1/p')
    if [[ $online -eq $g_total ]]; then
        exit 0
    fi

    sleep 0.5s
    wait=$((wait+1))
done

echo "wait all chunkserver online timeout, total=$g_total, online=$online"
exit 1

#!/usr/bin/env bash

# Usage: no args, just run it in bash
# Created Date: 2022-03-09
# Author: aspirer


# FIXME(P0): wait not works, return -12
wait=0
while ((wait<60))
do
    status=$(curve_ops_tool chunkserver-status |grep "offline")
    total=$(echo ${status} | grep -c "total num = 0")
    offline=$(echo ${status} | grep -c "offline = 0")
    if [ ${total} -eq 0 ] && [ ${offline} -eq 1 ]; then
        echo "CURVEADM_SUCCESS"
        exit 0
    fi
    sleep 0.5s
    wait=$(expr ${wait} + 1)
done
echo "CURVEADM_TIMEOUT"
exit 1

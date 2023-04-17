#!/usr/bin/env bash

# Usage: wait ADDR...
# Example: wait 10.0.10.1:2379 10.0.10.2:2379
# Created Date: 2021-11-25
# Author: Jingli Chen (Wine93)


[[ -z $(which curl) ]] && apt-get install -y curl
wait=0
while ((wait<20))
do
    for addr in "$@"
    do
        curl --connect-timeout 3 --max-time 10 $addr -Iso /dev/null
        if [ $? == 0 ]; then
           exit 0
        fi
    done
    sleep 0.5s
    wait=$(expr $wait + 1)
done
echo "wait timeout"
exit 1

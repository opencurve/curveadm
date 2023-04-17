#!/usr/bin/env bash

# Usage: recycle SOURCE DESTINATION SIZE
# Example: recycle '/data/chunkserver0/copysets /data/chunkserver0/recycler' /data/chunkserver0/chunkfilepool 16781312
# Created Date: 2022-01-13
# Author: Jingli Chen (Wine93)


g_source=$1
g_dest=$2
g_size=$3
chunkid=$(ls -vr ${g_dest} | head -n 1)
chunkid=${chunkid%.clean}
for file in $(find $g_source -type f -size ${g_size}c -printf '%p\n'); do
    chunkid=$((chunkid+1))
    mv $file $g_dest/$chunkid
done

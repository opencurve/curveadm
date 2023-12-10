#!/usr/bin/env bash

# Usage: bash mark_server_pendding.sh CHUNKIP CHUNKPORT
# Example: bash mark_server_pendding.sh 127.0.0.1 18200
# Created Date: 2023-12-04
# Author: Caoxianfei(caoxianfei1)

IP=$1
PORT=$2

CHUNKSERVER_ID=$(curve_ops_tool chunkserver-list | awk -v ip="$IP" -v port="$PORT" '
BEGIN { FS = ", "; OFS = "\n" }
$0 ~ /chunkServerID/ {
	split($3, arr1, " = ")
	split($4, arr2, " = ")
	if (arr1[2] == ip && arr2[2] == port) {
		split($1, arr3, " = ")
		print arr3[2]
		exit
	}
}')

if [ -z "$CHUNKSERVER_ID" ]; then
    echo "chunkserver $IP:$PORT not found"
    exit 1
fi

/curvebs/tools/sbin/curvebs-tool -op=set_chunkserver -chunkserver_id=$CHUNKSERVER_ID -chunkserver_status=pendding 
if [ $? -ne 0 ]; then
    echo "failed to set chunkserver $IP:$PORT pendding status"
    exit 1
fi
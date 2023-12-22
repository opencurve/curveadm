#!/usr/bin/env bash

# Usage: 
# Example: 
# Created Date: 2023-12-15
# Author: Caoxianfei

etcdctl=$1
endpoints=$2
old_name=$3
new_name=$4
new_peer_url=$5

${etcdctl} --endpoints=${endpoints} member add ${new_name} --peer-urls ${new_peer_url} > add_etcd.log 2>&1
if [ $? -ne 0 ]; then
    if cat add_etcd.log | grep -q "Peer URLs already exists"; then 
        exit 0
    else
        exit 1
    fi
fi

# output=$(${etcdctl} --endpoints=${endpoints} member list)
# if [ $? -ne 0 ]; then 
#     echo "failed to list all etcd members"
#     exit 1
# fi

# id=$(echo "$output" | awk -v name="$old_name" -F ', ' '$3 == name {print $1}')
# if [ -z "${id}" ]; then
#     echo "failed to get id of member ${old_name}"
#     exit 1
# fi

# ${etcdctl} --endpoints=${endpoints} member remove ${id}
# if [ $? -ne 0 ]; then
#     echo "failed to remove member ${old_name}"
#     exit 1
# fi





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

tmplog=/tmp/_curveadm_add_etcd_

output=$(${etcdctl} --endpoints=${endpoints} member list)
if [ $? -ne 0 ]; then 
    echo "failed to list all etcd members"
    exit 1
fi

# if member has added, then skip
id=$(echo "$output" | awk -v name="$new_name" -F ', ' '$3 == name {print $1}')
if [ -z "${id}" ]; then
    echo "EXIST"
    exit 0
fi

${etcdctl} --endpoints=${endpoints} member add ${new_name} --peer-urls ${new_peer_url} > ${tmplog} 2>&1
if [ $? -ne 0 ]; then
    if cat ${tmplog} | grep -q "Peer URLs already exists"; then 
        exit 0
    else
        exit 1
    fi
fi


# ${etcdctl} --endpoints=${endpoints} member remove ${id}
# if [ $? -ne 0 ]; then
#     echo "failed to remove member ${old_name}"
#     exit 1
# fi





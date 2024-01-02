#!/usr/bin/env bash

# Usage: 
# Example: 
# Created Date: 2023-12-15
# Author: Caoxianfei

etcdctl=$1
endpoints=$2
old_name=$3

output=$(${etcdctl} --endpoints=${endpoints} member list)
if [ $? -ne 0 ]; then 
    echo "failed to list all etcd members"
    exit 1
fi

id=$(echo "$output" | awk -v name="$old_name" -F ', ' '$3 == name {print $1}')
# if not found the name then exit 0
if [ -z "${id}" ]; then
    echo "NOTEXIST"
    exit 0
fi

${etcdctl} --endpoints=${endpoints} member remove ${id}
if [ $? -ne 0 ]; then
    echo "failed to remove member ${old_name}"
    exit 1
fi





#!/usr/bin/env bash

# Usage: map USER VOLUME CREATE SIZE
# Example: map curve test true 10
# Created Date: 2022-01-10
# Author: Jingli Chen (Wine93)


g_user=$1
g_volume=$2
g_options=$3
g_stderr=/tmp/__curveadm_map__

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab
curve-nbd map --nbds_max=16 ${g_options} cbd:pool/${g_volume}_${g_user}_ > ${g_stderr} 2>&1
if [ $? -ne 0 ]; then
  cat ${g_stderr}
  exit 1
else
  echo "SUCCESS"
fi

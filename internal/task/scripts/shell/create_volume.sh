#!/usr/bin/env bash

# Usage: create_volume USER VOLUME SIZE
# Example: create_volume curve test 10
# Created Date: 2022-07-31
# Author: Jingli Chen (Wine93)


g_user=$1
g_volume=$2
g_size=$3

output=$(curve_ops_tool create -userName="$g_user" -fileName="$g_volume" -fileLength="$g_size")
if [ $? -ne 0 ]; then
  if [ "$output" = "CreateFile fail with errCode: 101" ]; then
     echo "EXIST"
  else
     echo "FAILED"
  fi
else
  echo "SUCCESS"
fi

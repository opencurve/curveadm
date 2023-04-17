#!/usr/bin/env bash

# Usage: create_fs USER VOLUME SIZE
# Example: create_fs curve test 10
# Created Date: 2022-01-04
# Author: chengyi01


g_curvefs_tool="curvefs_tool"
g_curvefs_tool_operator="create-fs"
g_rpc_timeout_ms="-rpcTimeoutMs=10000"
g_fsname="-fsName="
g_fstype="-fsType="
g_entrypoint="/entrypoint.sh"

function createfs() {
    g_fsname=$g_fsname$1
    g_fstype=$g_fstype$2

    $g_curvefs_tool $g_curvefs_tool_operator "$g_fsname" "$g_fstype" $g_rpc_timeout_ms
}

createfs "$@"

ret=$?
if [ $ret -eq 0 ]; then
    $g_entrypoint "$@"
    ret=$?
    exit $ret
else
    echo "CREATEFS FAILED"
    exit 1
fi

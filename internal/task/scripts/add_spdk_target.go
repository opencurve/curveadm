/*
 *  Copyright (c) 2021 NetEase Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

/*
 * Project: CurveAdm
 * Created Date: 2022-02-08
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: target USER VOLUME CREATE SIZE
 * Example: target curve test true 10
 * See Also: https://linux.die.net/man/8/tgtadm
 */

var ADD_SPDK_TARGET = `
#!/usr/bin/env bash

g_sockname="/var/tmp/spdk.sock"
g_rpcpath="/usr/libexec/spdk/scripts/rpc.py"
g_isexclusive=0
g_targetip=10.0.0.1
g_targetport=3260
g_initiatorip=10.0.0.0
g_initiatormask=8
g_volumename=disk1
g_volume=$1
g_devno=$2
g_originbdev=$3
g_blocksize=$4
g_user=$5
g_create=$6
g_size=$7

g_cbdpath=//${g_devno}_${g_user}_
g_curvebdev=cbd_${g_devno}
g_ocf=ocf_${g_devno}
g_aio=aio_${g_devno}

mkdir -p /curvebs/nebd/data/lock
touch /etc/curve/curvetab

if [ $g_create == "true" ]; then
    output=$(curve_ops_tool create -userName=$g_user -fileName=$g_volume -fileLength=$g_size)
    if [ $? -ne 0 ]; then
        if [ "$output" != "CreateFile fail with errCode: 101" ]; then
            exit 1
        fi
    fi
fi

if sudo ${g_rpcpath} -s ${g_sockname} iscsi_get_target_nodes >/dev/null | grep "$g_curvebdev" | grep name >/dev/null; then
    echo "spdk target is already created, please try anther name"
    exit 1
fi

sudo ${g_rpcpath} -s ${g_sockname} bdev_aio_create ${g_aio} ${g_originbdev} 512 >/dev/null
sudo ${g_rpcpath} -s ${g_sockname} bdev_cdb_create -b ${g_curvebdev} --cbd ${g_cbdpath} --exclusive=${g_isexclusive} --blocksize=${g_blocksize} >/dev/null
sudo ${g_rpcpath} -s ${g_sockname} bdev_ocf_create $g_ocf wb $g_aio $g_curvebdev > /dev/null

sudo ${g_rpcpath} -s ${g_sockname} iscsi_create_portal_group 1 ${g_targetip}:${g_targetport} > /dev/null
sudo ${g_rpcpath} -s ${g_sockname} iscsi_create_initiator_group 2 ANY ${g_initiatorip}/${g_initiatormask} > /dev/null

sudo ${g_rpcpath} -s ${g_sockname} iscsi_create_target_node ${g_volumename} "Data Disk1" "g_ocf:0 ${g_aio}:1" 1:2 64 -d > /dev/null

sudo iscsiadm --mode discovery -t sendtargets --portal 10.0.0.1:3260
exit 0

`

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

var TARGET = `
#!/usr/bin/env bash

g_user=$1
g_volume=$2
g_create=$3
g_size=$4
g_blocksize=$5
g_devno=$6
g_originbdev=$7
g_spdk=$8

g_tid=1
g_sockname="/var/tmp/spdk.sock"
g_rpcpath="/usr/libexec/spdk/scripts/rpc.py"
g_isexclusive=0
g_targetip=10.0.0.1
g_targetport=3260
g_initiatorip=10.0.0.0
g_initiatormask=8
g_volumename=disk1

g_cbdpath=//${g_devno}_${g_user}_
g_curvebdev=cbd_${g_devno}
g_ocf=ocf_${g_devno}
g_aio=aio_${g_devno}

g_image=cbd:pool/${g_volume}_${g_user}_
g_image_md5=$(echo -n ${g_image} | md5sum | awk '{ print $1 }')
g_targetname=iqn.$(date +"%Y-%m").com.opencurve:curve.${g_image_md5}

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

if [ $g_spdk == "true" ]; then
	for ((i=1;;i++)); do
	    tgtadm --lld iscsi --mode target --op show --tid $i 1>/dev/null 2>&1
	    if [ $? -ne 0 ]; then
	        g_tid=$i
	        break
	    fi
	done

	tgtadm --lld iscsi \
	   --mode target \
	   --op new \
	   --tid ${g_tid} \
	   --targetname ${g_targetname}
	if [ $? -ne 0 ]; then
	   echo "tgtadm target new failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode logicalunit \
	    --op new \
	    --tid ${g_tid} \
	    --lun 1 \
	    --bstype curve \
	    --backing-store ${g_image} \
	    --blocksize ${g_blocksize}
	if [ $? -ne 0 ]; then
	   echo "tgtadm logicalunit new failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode logicalunit \
	    --op update \
	    --tid ${g_tid} \
	    --lun 1 \
	    --params vendor_id=NetEase,product_id=CurveVolume,product_rev=2.0
	if [ $? -ne 0 ]; then
	   echo "tgtadm logicalunit update failed"
	   exit 1
	fi

	tgtadm --lld iscsi \
	    --mode target \
	    --op bind \
	    --tid ${g_tid} \
	    -I ALL
	if [ $? -ne 0 ]; then
	   echo "tgtadm target bind failed"
	   exit 1
	fi
else
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

fi
`

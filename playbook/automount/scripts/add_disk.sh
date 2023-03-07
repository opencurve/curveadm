#!/usr/bin/env bash

#
#  Copyright (c) 2023 NetEase Inc.
#
#  Licensed under the Apache License, Version 2.0 (the "License");
#  you may not use this file except in compliance with the License.
#  You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.
#

# check parameters
if [ $# -ne 2 ]; then
    echo "Usage: $0 <device> <mountpoint>"
    exit 1
fi

g_device=$1
g_mountpoint=$2
g_systemd_name=""
g_disk_uuid=""
g_filesystem_type=""
g_options="defaults"

g_blkid_cmd="${SUDO_ALIAS} blkid -o value"
g_lsblk_cmd="${SUDO_ALIAS} lsblk"
g_mkdir_cmd="${SUDO_ALIAS} mkdir -p"
g_mount_cmd="${SUDO_ALIAS} mount"
g_umount_cmd="${SUDO_ALIAS} umount"
g_systemctl_cmd="${SUDO_ALIAS} systemctl"
g_cat_cmd="${SUDO_ALIAS} cat"
g_tee_cmd="${SUDO_ALIAS} tee"

function msg() {
    printf '%b' "$1" >&2
}

function success() {
    msg "\33[32m[✔]\33[0m ${1}${2}"
}

function die() {
    msg "\33[31m[✘]\33[0m ${1}${2}"
    exit 1
}

precheck() {
    # check block devices
    ${g_lsblk_cmd} ${g_device} >& /dev/null
    if [ $? -ne 0 ];then
        die "${g_device} is not a block devices!\n"
        exit 1
    fi
    # mkdir mountpoint
    if [ ! -d ${g_mountpoint} ];then
        die "${g_mountpoint} is not a directory!\n"
    fi
    
    # check uuid
    g_disk_uuid=`${g_blkid_cmd} -s UUID ${g_device}`
    if [ ! ${g_disk_uuid} ];then
        die "${g_device} has no uuid!\n"
    fi
    
    # check options by mount
    ${g_umount_cmd} ${g_mountpoint} >& /dev/null
    if [ "${OPTIONS}" != "" ];then
        g_options=${OPTIONS}
    fi
    out=`${g_mount_cmd} --options ${g_options} ${g_device} ${g_mountpoint} 2>&1`
    if [ $? -ne 0 ];then
        die "${out}!\n"
        exit 1
    fi
    ${g_umount_cmd} ${g_mountpoint} >& /dev/null
}

init() {
    src=${g_mountpoint//\//-}
    g_systemd_name=${src#-}
    g_filesystem_type=`${g_blkid_cmd} -s TYPE ${g_device}`
}

create_systemd() {
    # add mount
    echo "[Unit]
Description=Mount ${g_device} at ${g_mountpoint}

[Mount]
What=/dev/disk/by-uuid/${g_disk_uuid}
Where=${g_mountpoint}
Type=${g_filesystem_type}
Options=${g_options}

[Install]
WantedBy=multi-user.target
    " | ${g_tee_cmd} /etc/systemd/system/${g_systemd_name}.mount >& /dev/null
    
    # add automount
    echo "[Unit]
Description=Automount ${g_device} at ${g_mountpoint}

[Automount]
Where=${g_mountpoint}

[Install]
WantedBy=multi-user.target
    " | ${g_tee_cmd} /etc/systemd/system/${g_systemd_name}.automount >& /dev/null
    
    ${g_systemctl_cmd} daemon-reload >& /dev/null
    ${g_systemctl_cmd} enable ${g_systemd_name}.mount >& /dev/null
    ${g_systemctl_cmd} enable ${g_systemd_name}.automount >& /dev/null
    ${g_systemctl_cmd} start ${g_systemd_name}.automount >& /dev/null
    ${g_systemctl_cmd} start ${g_systemd_name}.mount >& /dev/null
    
    sleep 3
    
    status=`${g_systemctl_cmd} is-active ${g_systemd_name}.mount`
    auto_status=`${g_systemctl_cmd} is-active ${g_systemd_name}.automount`
    if [ "${status}" == "active" ] && [ "${auto_status}" == "active" ];then
        success "add automount ${g_device} to ${g_mountpoint} successfully!\n"
    else
        die "add automount ${g_device} to ${g_mountpoint} failed please check!\n"
        exit 1
    fi
}

precheck
init
create_systemd

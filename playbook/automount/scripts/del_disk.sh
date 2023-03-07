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
if [ $# -ne 1 ]; then
    echo "Usage: $0 <mountpoint>"
    exit 1
fi

g_mountpoint=$1
g_systemd_name=""
g_rm_cmd="${SUDO_ALIAS} rm"
g_umount_cmd="${SUDO_ALIAS} umount"

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
    # check systemd file is exist
    src=${g_mountpoint//\//-}
    g_systemd_name=${src#-}
    if [ ! -e /etc/systemd/system/${g_systemd_name}.automount ];then
        die "no /etc/systemd/system/${g_systemd_name}.automount file!\n"
        exit 1
    fi

    if [ ! -e /etc/systemd/system/${g_systemd_name}.mount ];then
        die "no /etc/systemd/system/${g_systemd_name}.mount file!\n"
        exit 1
    fi
}

del_systemd() {
    ${g_rm_cmd} /etc/systemd/system/${g_systemd_name}.automount
    ${g_rm_cmd} /etc/systemd/system/${g_systemd_name}.mount
    ${g_rm_cmd} /etc/systemd/system/multi-user.target.wants/${g_systemd_name}.automount
    ${g_rm_cmd} /etc/systemd/system/multi-user.target.wants/${g_systemd_name}.mount

    ${g_systemctl_cmd} daemon-reload >& /dev/null

    ${g_umount_cmd} ${g_mountpoint}

    success "rm automount ${g_mountpoint} successfully" 
}

precheck
init
del_systemd

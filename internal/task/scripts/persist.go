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
 * Created Date: 2022-05-05
 * Author: Ye Tao(YhhHaa)
 */

package scripts

var PERSIST_MOUNTPOINTS = `#!/usr/bin/env bash

# set path
fstab_log_dir=~/.curveadm/data/
fstab_log_path=~/.curveadm/data/fstab.log
fstab_path=/etc/fstab

# create path
if [[ ! -d "$fstab_log_dir" ]]; then mkdir -p $fstab_log_dir; fi
if [[ ! -f "$fstab_log_path" ]]; then touch $fstab_log_path; fi

# get devices and mount points from input parameters
declare -a device_arr
declare -a mount_point_arr
device_index=0
mount_point_index=0
status=0
for param in "$@"
do
    if [[ $param == "1" ]]; then status=1; continue; fi
    if [[ $param == "2" ]]; then status=2; continue; fi

    # get devices
    if [[ $status -eq 1 ]]; then
        device_arr[$device_index]=$param
        let device_index++
    fi

    # get mount point            
    if [[ $status -eq 2 ]]; then
        mount_point_arr[$mount_point_index]=$param
        let mount_point_index++
    fi
done

# read log into dict{uuid:mountpoint}
declare -A fstab_dict
while read line
do
    # change entry into arr format
    arr=($line)

    # save entry into dict
    if [[ -n $(echo $line | grep 'UUID' | grep 'ext4') ]]; then
        fstab_dict[${arr[0]}]=${arr[1]}
    fi
done < $fstab_log_path

# clear all past invalid entry in fstab with invalid fstab.log
for uuid in $(echo ${!fstab_dict[*]})
do
    if [[ -n ${uuid} ]]; then sed -i "/${uuid}/d" $fstab_path; fi
    if [[ -n ${fstab_dict[uuid]} ]]; then sed -i "\|${fstab_dict[uuid]}|d" $fstab_path; fi
done
cat /dev/null > $fstab_log_path


# delete invalid entry in fstab
for disk_device in ${device_arr[@]}
do
    disk_arr=($(lsblk -f | grep ${disk_device}))
    disk_uuid=${disk_arr[2]}

    if [[ -n ${disk_uuid} ]]; then 
        sed -i "/${disk_uuid}/d" $fstab_path
        sed -i "/${disk_uuid}/d" $fstab_log_path
    fi
done

for disk_mountpoint in ${mount_point_arr[@]}
do
    if [[ -n ${disk_mountpoint} ]]; then 
        sed -i "\|${disk_mountpoint}|d" $fstab_path
        sed -i "\|${disk_mountpoint}|d" $fstab_log_path
    fi
done

# write current device and mountpoint into fstab and fstab.log
for(( index=0; index<${#device_arr[@]}; index++ ))
do
    disk_arr=($(lsblk -f | grep ${device_arr[$index]}))
    disk_uuid=${disk_arr[2]}
    disk_mountpoint=${mount_point_arr[$index]}

    if [[ -n "${disk_uuid}" && -n "${disk_mountpoint}" ]]; then
        echo "UUID=${disk_uuid} ${disk_mountpoint} ext4 defaults,nofail 0 2" >> $fstab_path
        echo "UUID=${disk_uuid} ${disk_mountpoint} ext4 defaults,nofail 0 2" >> $fstab_log_path
    fi
done

`

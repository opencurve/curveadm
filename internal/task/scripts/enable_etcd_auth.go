/*
*  Copyright (c) 2023 NetEase Inc.
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
* Project: Curveadm
* Created Date: 2023-08-02
* Author: wanghai (SeanHai)
*/

package scripts

var ENABLE_ETCD_AUTH = `
#!/usr/bin/env bash

if [ $# -ne 3 ]; then
  echo "Usage: $0 endpoints username password"
  exit 1
fi

endpoints=$1
username=$2
password=$3
root_user=root

# create root user
etcdctl --endpoints=${endpoints} user add ${root_user}:${password} && \
etcdctl --endpoints=${endpoints} user grant-role ${root_user} root || exit 1

# create user if not root
if [ "${username}" != "${root_user}" ]; then
  etcdctl --endpoints=${endpoints} user add ${username}:${password} && \
  etcdctl --endpoints=${endpoints} user grant-role ${username} root || exit 1
fi

# enable auth
etcdctl --endpoints=${endpoints} auth enable --user=${root_user}:${password} || exit 1
`

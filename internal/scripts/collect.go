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
 * Created Date: 2021-11-26
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: collect PREFIX CONTAINER_ID DEST_DIR
 * Example: collect 8d9b0c0bdec5 /tmp/dest_dir
 */
var COLLECT = `
############################  GLOBAL VARIABLES
g_prefix="$1"
g_container_id="$2"
g_dest_dir="$3"
g_log_dir="$g_prefix/logs"
g_conf_dir="$g_prefix/conf"

############################ FUNCTIONS
function docker_cmd() {
    sudo docker exec $g_container_id /bin/bash -c "$1"
}

function docker_cp() {
    sudo docker cp $g_container_id:$1 $2
}

function copy_logs() {
    sudo docker logs $g_container_id > $g_dest_dir/logs/stdout
    sudo docker logs $g_container_id 2> $g_dest_dir/logs/stderr
    for file in $(docker_cmd "ls $g_log_dir | tail -n 5")
    do
        docker_cp $g_log_dir/$file $g_dest_dir/logs
    done
}

function copy_config() {
    docker_cp $g_conf_dir $g_dest_dir
}

function main() {
    mkdir -p $g_dest_dir/{logs,conf}
    copy_logs
    copy_config
}

############################  MAIN()
main "$@"
`

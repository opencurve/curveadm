/*
 *  Copyright (c) 2022 NetEase Inc.
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
 * Created Date: 2022-01-04
 * Author: chengyi01
 */

package scripts

/*
 * Usage: create fs before mount fs
 * Example: createfs /test
 */
var CREATEFS = `
g_curvefs_tool="curvefs_tool"
g_curvefs_tool_operator="create-fs"
g_rpc_timeout_ms="-rpcTimeoutMs=5000"
g_fsname="-fsName="
g_entrypoint="/entrypoint.sh"

function createfs() {
    g_fsname=$g_fsname$1

    $g_curvefs_tool $g_curvefs_tool_operator $g_fsname $g_rpc_timeout_ms 
}

createfs "$@"

if [ $? -eq 0 ]; then
    $g_entrypoint "$@"
fi
`

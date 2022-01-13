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
 * Created Date: 2022-01-13
 * Author: Jingli Chen (Wine93)
 */

package scripts

/*
 * Usage: recycle SOURCE DESTINATION SIZE
 * Example: recycle '/data/chunkserver0/copysets /data/chunkserver0/recycler' /data/chunkserver0/chunkfilepool 16781312
 */
var RECYCLE = `
g_source=$1
g_dest=$2
g_size=$3
chunkid=$(ls -vr ${g_dest} | head -n 1)
for file in $(find $g_source -type f -size ${g_size}c -printf '%p\n'); do
    chunkid=$((chunkid+1))
    mv $file $g_dest/$chunkid
done
`

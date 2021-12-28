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
 * Created Date: 2021-12-27
 * Author: Jingli Chen (Wine93)
 */

package scripts

var FORMAT = `
binary=$1
percent=$2
chunkfile_size=$3
chunkfile_pool_dir=$4
chunkfile_pool_meta_path=$5

mkdir -p $chunkfile_pool_dir
$binary \
  -allocatePercent=$percent \
  -fileSize=$chunkfile_size \
  -filePoolDir=$chunkfile_pool_dir \
  -filePoolMetaPath=$chunkfile_pool_meta_path \
  -fileSystemPath=$chunkfile_pool_dir
`

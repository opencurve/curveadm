#!/usr/bin/env bash

# Created Date: 2021-12-27
# Author: Jingli Chen (Wine93)


binary=$1
percent=$2
chunkfile_size=$3
chunkfile_pool_dir=$4
chunkfile_pool_meta_path=$5
chunkfile_block_size=$6

mkdir -p $chunkfile_pool_dir
$binary \
  -allocatePercent=$percent \
  -fileSize=$chunkfile_size \
  -filePoolDir=$chunkfile_pool_dir \
  -filePoolMetaPath=$chunkfile_pool_meta_path \
  -fileSystemPath=$chunkfile_pool_dir \
  -blockSize=$chunkfile_block_size \
  -undefok blockSize

#!/bin/bash

g_ls="${SUDO_ALIAS} ls"
g_ps="${SUDO_ALIAS} ps"
g_cat="${SUDO_ALIAS} cat"
g_tee="${SUDO_ALIAS} tee"

if [ ${PERF_TUNE} != "true" ]; then
    echo 'PERF_TUNE is not true, exit'
    exit
fi

set_value()
{
    local value=$1
    local path=$2
    echo ${value} | ${g_tee} ${path} &> /dev/null
}

for bcache in $(${g_ls} /sys/block | grep bcache)
do
    backing_dev=$(${g_cat} /sys/block/${bcache}/bcache/backing_dev_name)
    backing_sectors=$(${g_cat} /sys/block/${backing_dev}/queue/max_sectors_kb)
    backing_ahead=$(${g_cat} /sys/block/${backing_dev}/queue/read_ahead_kb)

    set_value ${backing_sectors} /sys/block/${bcache}/queue/max_sectors_kb 
    set_value ${backing_ahead} /sys/block/${bcache}/queue/read_ahead_kb
    set_value ${CACHE_MODE} /sys/block/${bcache}/bcache/cache_mode
    set_value 1 /sys/block/${bcache}/bcache/clear_stats
    set_value 0 /sys/block/${bcache}/bcache/readahead
    set_value 40 /sys/block/${bcache}/bcache/writeback_percent
    set_value 10 /sys/block/${bcache}/bcache/writeback_delay
    set_value 1 /sys/block/${bcache}/bcache/writeback_rate_minimum 
    set_value 0 /sys/block/${bcache}/bcache/cache/congested_read_threshold_us
    set_value 0 /sys/block/${bcache}/bcache/cache/congested_write_threshold_us
    set_value 0 /sys/block/${bcache}/bcache/sequential_cutoff
    set_value lru /sys/block/${bcache}/bcache/cache/cache0/cache_replacement_policy
    set_value 1 /sys/block/${bcache}/bcache/cache/internal/gc_after_writeback

done

echo "bcache perf tune success, cache mode is ${CACHE_MODE}"


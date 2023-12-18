#!/bin/bash

g_ls="${SUDO_ALIAS} ls"
g_cat="${SUDO_ALIAS} cat"
g_which="${SUDO_ALIAS} which"
g_readlink="${SUDO_ALIAS} readlink"

show_bcache()
{
    if [ -n "$(${g_which} bcache-status)" ]; then
        ${SUDO_ALIAS} bcache-status -s
    elif [ -n "$(${g_which} bcache)" ]; then
        ${SUDO_ALIAS} bcache show
    else
        for bcache in $(${g_ls} /sys/block | grep bcache)
        do
            echo "${bcache} info:"
            echo "----------------------------"
            echo "backing device: /dev/$(${g_cat} /sys/block/${bcache}/bcache/backing_dev_name)"
            echo "cache device: /dev/$(${g_readlink} /sys/block/${bcache}/bcache/cache/cache0 |awk -F'/' '{print $(NF-1)}')"
            echo "cache mode: $(${g_cat} /sys/block/${bcache}/bcache/cache_mode | grep -oP "(?<=\[)[^\]]*(?=\])")"
            echo "cache state: $(${g_cat} /sys/block/${bcache}/bcache/state)"
            echo
        done
    fi
}

show_bcache


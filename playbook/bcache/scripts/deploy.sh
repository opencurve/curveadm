#!/bin/bash

g_ls="${SUDO_ALIAS} ls"
g_lsmod="${SUDO_ALIAS} lsmod"
g_modinfo="${SUDO_ALIAS} modinfo"
g_which="${SUDO_ALIAS} which"
g_tee="${SUDO_ALIAS} tee"
g_make_bcache="${SUDO_ALIAS} make-bcache"
g_bcache_super_show="${SUDO_ALIAS} bcache-super-show"

defalut_cache_mode=none


set_value()
{
    local value=$1
    local path=$2
    echo ${value} | ${g_tee} ${path} &> /dev/null
}

pre_check()
{
    # check bcache-tools is installed
    if [ -z "$(${g_which} make-bcache)" ]; then
        echo "make-bcache could not be found"
        exit 1
    fi

    if [ -z "$(${g_which} bcache-super-show)" ]; then
        echo "bcache-super-show could not be found"
        exit 1
    fi

    # check bcache module is exist
    ${g_modinfo} bcache &> /dev/null
    if [ $? != 0 ]; then
        echo "bcache module not be found"
        exit 1
    fi

    # check bcache device is exist
    if [ -n "$(${g_ls} /sys/block | grep bcache)" ];then
        echo "bcache device is exist, clean it first"
        exit 1
    fi

    # check backend and cache device number
    if [ $(echo ${BACKING_DEV} |wc -l) != $(echo ${CACHE_DEV} |wc -l) ];then
        echo "only support one cache device with one backing device now!"
        exit 1 
    fi

    echo "pre_check success"
}

deploy_bcache()
{
    for hdd in ${BACKING_DEV}
    do
        ${g_make_bcache} -B --wipe-bcache ${hdd} &> /dev/null
        if [ $? = 0 ]; then
            set_value ${hdd} /sys/fs/bcache/register
        else
            echo "make bcache device ${hdd} failed"
            exit 1
        fi
    done

    for cache in ${CACHE_DEV}
    do
        ${g_make_bcache} -C --wipe-bcache -b 262144 ${cache} &> /dev/null
        if [ $? = 0 ]; then
            set_value ${cache} /sys/fs/bcache/register
        else
            echo "make bcache device ${cache} failed"
            exit 1
        fi
    done

    idx=0
    for cache in ${CACHE_DEV}
    do
        uuid=$(${g_bcache_super_show} ${cache} | grep cset.uuid | awk '{print $2}')
        set_value ${uuid} /sys/block/bcache${idx}/bcache/attach
        idx=$((idx+1))
    done

    echo "now set cache mode to ${defalut_cache_mode}"
    # using none mode before chunkfilepool formated
    for bcache in $(${g_ls} /sys/block | grep bcache)
    do
        set_value ${defalut_cache_mode} /sys/block/${bcache}/bcache/cache_mode
    done

    echo "bcache deploy success, please format chunkfilepool and walfilepool manually"
}

pre_check
deploy_bcache


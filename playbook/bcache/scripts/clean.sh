#!/bin/bash

g_ls="${SUDO_ALIAS} ls"
g_ps="${SUDO_ALIAS} ps"
g_cat="${SUDO_ALIAS} cat"
g_tee="${SUDO_ALIAS} tee"
g_umount="${SUDO_ALIAS} umount"
g_wipefs="${SUDO_ALIAS} wipefs"
g_bcache_super_show="${SUDO_ALIAS} bcache-super-show"


set_value()
{
    local value=$1
    local path=$2
    echo ${value} | ${g_tee} ${path} &> /dev/null
}

pre_check()
{
    #check chunkserver is running
    pid=$(${g_ps} -ef | grep chunkserver | grep -v grep | awk '{print $2}')
    if [ -n "${pid}" ]; then
        echo "chunkserver is running, please stop it first"
        exit 1
    fi

    #check bcache dirty data
    for bcache in $(${g_ls} /sys/block | grep bcache)
    do
        if [ "$(${g_cat} /sys/block/${bcache}/bcache/dirty_data)" != "0.0k" ]; then
            echo "${bcache} has dirty data, please stop chunkserver and wait it cleaned"
            exit 1
        fi
    done

    echo "pre_check success"
}


stop_bcache()
{
    ${g_umount} /data/chunkserver* &> /dev/null
    ${g_umount} /data/wal/chunkserver* &> /dev/null

    bcache_devs=$(${g_ls} /sys/block | grep bcache)
    for bcache in ${bcache_devs}
    do
        backdev=/dev/$(${g_cat} /sys/block/${bcache}/bcache/backing_dev_name)
        uuid=$(${g_bcache_super_show} ${backdev} |grep cset |awk '{print $NF}')

        set_value 1 /sys/block/${bcache}/bcache/detach
        set_value 1 /sys/fs/bcache/${uuid}/unregister
        set_value 1 /sys/block/${bcache}/bcache/stop
    done

    set_value 1 /sys/fs/bcache/pendings_cleanup

    sleep 1

    bcache_devs=$(${g_ls} /sys/block | grep bcache)
    cache_sets=$(${g_ls} /sys/fs/bcache | grep "-")
    if [ -n "${bcache_devs}" ] || [ -n "${cache_sets}" ]; then
        # need retry to wait bcache stop
        echo "stop bcache failed"
        exit 1
    fi
    echo "stop bcache success"
}

clean_bcache_data()
{
    if [ x"${CLEAN_DATA}" != x"true" ]; then
        echo "no need to clean data"
        exit 0
    fi
    
    for hdd in ${BACKING_DEV}
    do
        ${g_wipefs} -a --force ${hdd} &> /dev/null
        if [ $? != 0 ]; then
            echo "wipefs backing device ${hdd} failed"
            exit 1
        fi
    done

    for cache in ${CACHE_DEV}
    do
        ${g_wipefs} -a --force ${cache} &> /dev/null
        if [ $? != 0 ]; then
            echo "wipefs cache device ${cache} failed"
            exit 1
        fi
    done

    echo "clean backing and cache devices data success"
}


pre_check
stop_bcache
clean_bcache_data


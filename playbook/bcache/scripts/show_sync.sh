#!/bin/bash

g_ls="${SUDO_ALIAS} ls"
g_mdadm="${SUDO_ALIAS} mdadm"

show_sync()
{
    for cache in $(${g_ls} /dev | grep md)
    do
        echo "${cache} info:"
        echo "----------------------------"
        ${g_mdadm} --detail ${cache}
        echo
    done
}

show_sync
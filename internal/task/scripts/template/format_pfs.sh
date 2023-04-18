#!/usr/bin/env bash

# Created Date: 2023-04-18
# Author: Jingli Chen (Wine93)


############################  GLOBAL VARIABLES
g_pfs="{{.pfs}}"
g_setup_script="{{.setup_script}}"
g_device="{{.device}}"
g_curve_format="{{.curve_format}}"
g_percent="{{.percent}}"
g_controller=""
g_namespace=""
g_controller_saved="{{.controller_saved}}"

############################ FUNCTIONS
msg() {
    printf '%b' "$1" >&2
}

die() {
    msg "\33[31m[✘]\33[0m ${1}${2}"
    exit 1
}

# NVMe  0000:d8:00.0  144d  a80a  1  nvme  nvme3  nvme3n1
cleanup() {
    "${g_setup_script}" reset
    rm ${g_controller_saved}
    g_controller=$("${g_setup_script}" status | grep "${g_device}" | awk '{ print $2}')
    g_namespace="${g_controller}n1"
    if [ -z "${g_controller}" ]; then
        die "${g_device} controller not found\n"
    fi
}

main() {
    cleanup
    HUGE_EVEN_ALLOC=yes NRHUGE=51200 PCI_ALLOWED="${g_controller}" "${g_setup_script}"
    "${g_pfs}" -K "${g_controller}" -C spdk mkfs -f "${g_namespace}"
    "${g_curve_format}" \
        -allocatePercent="${g_percent}" \
        -filePoolDir=/"${g_namespace}"/chunkfilepool \
        -filePoolMetaPath=/"${g_namespace}"/chunkfilepool.meta \
        -fileSystemPath=/"${g_namespace}"/chunkfilepool \
        -fileSystemType=pfs \
        -pfs_pbd_name="${g_namespace}" \
        -spdk_nvme_controller="${g_controller}"
    echo "${g_controller} ${g_namespace}" > ${g_controller_saved}
}

############################  MAIN()
main "$@"

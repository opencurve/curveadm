#!/usr/bin/env bash

g_container_name="memcached-"${PORT}
g_exporter_container_name="memcached-exporter-"${EXPORTER_PORT}
g_docker_cmd="${SUDO_ALIAS} ${ENGINE}"
g_rm_cmd="${SUDO_ALIAS} rm -rf"

function msg() {
    printf '%b' "$1" >&2
}

function success() {
    msg "\33[32m[✔]\33[0m ${1}${2}"
}

function die() {
    msg "\33[31m[✘]\33[0m ${1}${2}"
    exit 1
}

precheck() {
    container_id=`${g_docker_cmd} ps --all --format "{{.ID}}" --filter name=${g_container_name} --all`
    if [ -z ${container_id} ]; then
        die "container [${g_container_name}] not exists!!!\n"
        exit 1
    fi
    if [ "${EXPORTER_PORT}" ];then
        container_id=`${g_docker_cmd} ps --all --format "{{.ID}}" --filter name=${g_exporter_container_name}`
        if [ -z ${container_id} ]; then
            die "container [${g_exporter_container_name}] not exists!!!\n"
            exit 1
        fi
    fi
}

stop_container() {
    msg=`${g_docker_cmd} rm ${g_container_name}`
    if [ $? -ne 0 ];then
        die "${msg}\n"
        exit 1
    fi
    success "rm container[${g_container_name}]\n"
    if [ "${EXPORTER_PORT}" ];then
        msg=`${g_docker_cmd} rm ${g_exporter_container_name}`
        if [ $? -ne 0 ];then
            die "${msg}\n"
            exit 1
        fi
        success "rm container[${g_exporter_container_name}]\n"
    fi
}

rm_cachefile() {
    cachefile_path=(${EXT_PATH//:/ })
    ${g_rm_cmd} ${cachefile_path}
}

precheck
stop_container
rm_cachefile

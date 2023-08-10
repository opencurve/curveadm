#!/usr/bin/env bash

g_container_name="memcached-"${PORT}
g_docker_cmd="${SUDO_ALIAS} ${ENGINE}"
g_rm_cmd="${SUDO_ALIAS} rm -rf"
g_mkdir_cmd="${SUDO_ALIAS} mkdir -p"
g_touch_cmd="${SUDO_ALIAS} touch"
g_status="running"

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
    # check ext path
    get_status_container
    if [ "${EXT_PATH}" ] && [ ${g_status} != "running" ]; then
        cachefile_path=(${EXT_PATH//:/ })
        ${g_rm_cmd} ${cachefile_path}
        ${g_mkdir_cmd} $(dirname ${cachefile_path})
        ${g_touch_cmd} ${cachefile_path}
    fi
}

start_container() {
    ${g_docker_cmd} start ${g_container_name} >& /dev/null
    success "start container[${g_container_name}]\n"
}

get_status_container() {
    g_status=`${g_docker_cmd} inspect --format='{{.State.Status}}' ${g_container_name}`
}

precheck
start_container

#!/usr/bin/env bash

g_container_name="memcached-"${PORT}
g_docker_cmd="${SUDO_ALIAS} ${ENGINE}"

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
    container_id=`${g_docker_cmd} ps --all --format "{{.ID}}" --filter name=${g_container_name}`
    if [ -z ${container_id} ]; then
        die "container [${g_container_name}] not exists!!!\n"
        exit 1
    fi
}

stop_container() {
    ${g_docker_cmd} stop ${g_container_name} >& /dev/null
    success "stop container[${g_container_name}]\n"
}

precheck
stop_container

#!/usr/bin/env bash

g_container_name="memcached-"${PORT}
g_expoter_container_name="memcached-exporter-"${EXPORTER_PORT}
g_start_args=""
g_docker_cmd="${SUDO_ALIAS} ${ENGINE}"
g_volume_bind=""
g_container_id=""

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
    g_container_id=`${g_docker_cmd} ps --all --format "{{.ID}}" --filter name=${g_container_name}`
    if [ -z ${g_container_id} ]; then
        success "container [${g_container_name}] not exists!!!"
        exit 1
    fi
    if [ "${EXPORTER_PORT}" ];then
        g_container_id=`${g_docker_cmd} ps --all --format "{{.ID}}" --filter name=${g_expoter_container_name}`
        if [ -z ${g_container_id} ]; then
            success "container [${g_expoter_container_name}] not exists!!!"
            exit 1
        fi
    fi
}

show_info_container() {
    ${g_docker_cmd} ps --all --filter "name=${g_container_name}" --format="table {{.ID}}\t{{.Names}}\t{{.Status}}"
    ${g_docker_cmd} ps --all --filter "name=${g_expoter_container_name}" --format="table {{.ID}}\t{{.Names}}\t{{.Status}}"
}

show_ip_port() {
    printf "memcached addr:\t%s:%d\n" ${LISTEN} ${PORT}
    printf "memcached-expoter addr:\t%s:%d\n" ${LISTEN} ${EXPORTER_PORT}
}

get_status_container() {
    status=`${g_docker_cmd} inspect --format='{{.State.Status}}' ${g_container_name}`
    if [ ${status} != "running" ]; then
            exit 1
    fi
    if [ "${EXPORTER_PORT}" ];then
        status=`${g_docker_cmd} inspect --format='{{.State.Status}}' ${g_expoter_container_name}`
        if [ ${status} != "running" ]; then
            exit 1
        fi
    fi
}

precheck
show_ip_port
show_info_container
get_status_container

#!/usr/bin/env bash

g_container_name="memcached-"${PORT}
g_start_args=""
g_docker_cmd="${SUDO_ALIAS} ${ENGINE}"
g_lsof_cmd="${SUDO_ALIAS} lsof"
g_rm_cmd="${SUDO_ALIAS} rm -rf"
g_mkdir_cmd="${SUDO_ALIAS} mkdir -p"
g_touch_cmd="${SUDO_ALIAS} touch"
g_volume_bind=""
g_status=""
g_user=""

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
    container_id=`${g_docker_cmd} ps --format "{{.ID}}" --filter name=${g_container_name} --all`
    if [ "${container_id}" ]; then
        success "container [${g_container_name}] already exists, skip\n"
        exit 0
    fi

    # check port
    ${g_lsof_cmd} -i:${PORT} >& /dev/null
    if [ $? -eq 0 ];then
        die "port[${PORT}] is in use!\n"
        exit 1
    fi

    # check ext path
    if [ "${EXT_PATH}" ]; then
        cachefile_path=(${EXT_PATH//:/ })
        ${g_rm_cmd} ${cachefile_path}
        ${g_mkdir_cmd} $(dirname ${cachefile_path})
        ${g_touch_cmd} ${cachefile_path}
    fi
}

init() {
    ${g_docker_cmd} pull ${IMAGE} >& /dev/null
    if [ "${LISTEN}" ]; then
        g_start_args="${g_start_args} --listen=${LISTEN}"
    fi
    if [ "${PORT}" ]; then
        g_start_args="${g_start_args} --port=${PORT}"
    fi
    if [ "${USER}" ]; then
        g_start_args="${g_start_args} --user=${USER}"
        g_user="--user ${USER}"
    fi
    if [ "${MEMORY_LIMIT}" ]; then
        g_start_args="${g_start_args} --memory-limit=${MEMORY_LIMIT}"
    fi
    if [ "${MAX_ITEM_SIZE}" ]; then
        g_start_args="${g_start_args} --max-item-size=${MAX_ITEM_SIZE}"
    fi
    if [ "${EXT_PATH}" ]; then
        g_start_args="${g_start_args} --extended ext_path=/memcached/data${EXT_PATH}"
        volume_path=(${EXT_PATH//:/ })
        g_volume_bind="--volume ${volume_path}:/memcached/data${volume_path}"
    fi
    if [ "${EXT_WBUF_SIZE}" ]; then
        g_start_args="${g_start_args} --extended ext_wbuf_size=${EXT_WBUF_SIZE}"
    fi
    if [ "${EXT_ITEM_AGE}" ]; then
        g_start_args="${g_start_args} --extended ext_item_age=${EXT_ITEM_AGE}"
    fi
    if [ "${VERBOSE}" ];then
        g_start_args="${g_start_args} -${VERBOSE}"
    fi
}

create_container() {
    success "create container [${g_container_name}]\n"
    ${g_docker_cmd} create --name ${g_container_name} ${g_user} --network host ${g_volume_bind} ${IMAGE} memcached ${g_start_args} >& /dev/null

    success "start container [${g_container_name}]\n"
    ${g_docker_cmd} start ${g_container_name} >& /dev/null

    success "wait 3 seconds, check container status...\n"
    sleep 3
    get_status_container
    show_info_container
    if [ ${g_status} != "running" ]; then
        exit 1
    fi
}

get_status_container() {
    g_status=`${g_docker_cmd} inspect --format='{{.State.Status}}' ${g_container_name}`
}

show_info_container() {
    ${g_docker_cmd} ps --all --filter "name=${g_container_name}" --format="table {{.ID}}\t{{.Names}}\t{{.Status}}"
}

precheck
init
create_container

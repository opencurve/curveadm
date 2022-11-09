#!/usr/bin/env bash

set -e
set -o pipefail

g_kind="$1"
g_prefix="/${g_kind}/playground"
g_roles=("etcd" "mds" "chunkserver")
[ "${g_kind}" = "curvefs" ] && { g_roles=("etcd" "mds" "metaserver"); }
g_user="playground"
g_volume="/playground"
g_topology="/curvebs/tools/conf/topology.json"

function start_service() {
    local role=$1
    local sequence=$2
    local prefix="${g_prefix}"/"${role}${sequence}"
    local conf_path="${prefix}"/conf/"${role}".conf
    local log_dir="${prefix}"/logs
    local data_dir="${prefix}"/data
    chmod 700 "${prefix}"/data

    case $role in
        etcd)
            mkdir -p "$g_prefix/data/wal"
            /curvebs/etcd/sbin/etcd \
                --config-file "$conf_path" \
                > "${log_dir}"/etcd.log 2>&1 &
            ;;
        mds)
            /curvebs/mds/sbin/curvebs-mds \
                --confPath "$conf_path" &
            ;;
        chunkserver)
            /curvebs/chunkserver/sbin/curvebs-chunkserver \
                -conf="$conf_path" \
                -enableExternalServer=false \
                -copySetUri=local://"${data_dir}"/copysets \
                -raftLogUri=curve://"${data_dir}"/copysets \
                -raftSnapshotUri=curve://"${data_dir}"/copysets \
                -raft_sync_segments=true \
                -raft_max_install_snapshot_tasks_num=1 \
                -chunkServerIp=127.0.0.1 \
                -chunkFilePoolDir="${data_dir}" \
                -walFilePoolDir="${data_dir}" \
                -raft_sync=true \
                -raft_max_segment_size=8388608 \
                -raft_use_fsync_rather_than_fdatasync=false \
                -chunkFilePoolMetaPath="${data_dir}"/chunkfilepool.meta \
                -chunkServerStoreUri=local://"${data_dir}" \
                -chunkServerMetaUri=local://"${data_dir}"/chunkserver.dat \
                -bthread_concurrency=18 \
                -raft_sync_meta=true \
                -chunkServerExternalIp=127.0.0.1 \
                -chunkServerPort=820${sequence} \
                -walFilePoolMetaPath="${data_dir}"/walfilepool.meta \
                -recycleUri=local://"${data_dir}"/recycler \
                -graceful_quit_on_sigterm=true &
            ;;
        *)
            echo "unknown role: $role"
            exit 1
            ;;
    esac

}

function start_etcd() {
    for ((i=0;i<3;i++)) do
        start_service etcd "$i"
    done
}

function start_mds() {
    for ((i=0;i<3;i++)) do
        start_service mds "$i"
    done
}

function start_chunkserver() {
    for ((i=0;i<3;i++)) do
        start_service chunkserver "$i"
    done
}

function create_physicalpool() {
    /curvebs/tools/sbin/curvebs-tool -op=create_physicalpool -cluster_map="${g_topology}"
}

function create_logicalpool() {
    /curvebs/tools/sbin/curvebs-tool -op=create_logicalpool -cluster_map="${g_topology}"
}

function start_nebd() {
    /curvebs/nebd/sbin/nebd-server \
        -confPath=/etc/nebd/nebd-server.conf \
        -log_dir=/curvebs/nebd/logs &
}

function create_volume() {
    curve_ops_tool create \
        -userName="${g_user}" \
        -fileName="${g_volume}" \
        -fileLength=10
}

function map_volume() {
    mkdir -p /curvebs/nebd/data/lock
    touch /etc/curve/curvetab
    curve-nbd map --nbds_max=16 cbd:pool/${g_volume}_${g_user}_
    wait -n
}

function start_curvebs() {
    start_etcd
    sleep 3
    start_mds
    sleep 3
    create_physicalpool
    start_chunkserver
    sleep 25
    create_logicalpool
    sleep 25
    create_volume
    start_nebd
    map_volume
}

function main() {
    if [ "${1}" = "curvebs" ]; then
        start_curvebs
    else
        echo "unsupport kind: ${1}"
        exit 1
    fi
}

main "$@"
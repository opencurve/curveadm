#!/usr/bin/env bash

# Usage: bash rotate.sh /mnt/logs/curvefs/client 86400

############################  GLOBAL VARIABLES
g_log_dir="$1"
g_color_yellow=$(printf '\033[33m')
g_color_red=$(printf '\033[31m')
g_color_normal=$(printf '\033[0m')

############################  BASIC FUNCTIONS
msg() {
    printf '%b' "${1}" >&2
}

success() {
    msg "${g_color_yellow}[✔]${g_color_normal} ${1}${2}"
}

die() {
    msg "${g_color_red}[✘]${g_color_normal} ${1}${2}"
    exit 1
}

############################ FUNCTIONS
precheck() {
    local log_dir="$1"
    if [ ! -d "${log_dir}" ]; then
        die "Log directory ${log_dir} does not exist.\n"
    fi

    if [ ! -w "${log_dir}" ]; then
        die "Log directory ${log_dir} is not writable.\n"
    fi
}

expired() {
    local path="$1"
    local timeout="$2"
    local mtime=$(stat -c %Y "${path}")
    local now=$(date +%s)
    if (( now - mtime > timeout )); then
        return 0
    fi
    return 1
}

delete() {
    local path="$1"
    rm "${path}"
    success "Delete ${path}\n"
}

rotate() {
    local log_dir="$1"
    local timeout="$2"
    for file in $(ls "${log_dir}" | grep -E '(access|curve-fuse|aws)(.+)(\.log)?'); do
        local path="${log_dir}/${file}"
        if expired "${path}" "${timeout}"; then
            delete "${path}"
        fi
    done
}

main() {
    local log_dir="$1"
    local timeout="$2"  # seconds
    precheck "${log_dir}"
    rotate "${log_dir}" "${timeout}"
}

############################  MAIN()
main "$@"

#!/usr/bin/env bash

############################  GLOBAL VARIABLES
g_curveadm_home="$HOME/.curveadm"
g_bin_dir="$g_curveadm_home/bin"
g_profile="${HOME}/.profile"
g_download_url="https://curveadm.nos-eastchina1.126.net/release/curveadm-latest.tar.gz"
g_color_yellow=`printf '\033[33m'`
g_color_red=`printf '\033[31m'`
g_color_normal=`printf '\033[0m'`

############################  BASIC FUNCTIONS
function msg() {
    printf '%b' "$1" >&2
}

function success() {
    msg "$g_color_yellow[✔]$g_color_normal ${1}${2}"
}

function die() {
    msg "$g_color_red[✘]$g_color_normal ${1}${2}"
    exit 1
}

function program_must_exist() {
    local ret='0'
    command -v $1 >/dev/null 2>&1 || { local ret='1'; }

    if [ "$ret" -ne 0 ]; then
        die "You must have '$1' installed to continue.\n"
    fi
}

############################ FUNCTIONS
function setup() {
    mkdir -p $g_curveadm_home/{bin,data,logs,temp}

    local confpath="$g_curveadm_home/curveadm.cfg"
    if [ ! -f $confpath ]; then
        cat << __EOF__ > $confpath
[defaults]
log_level = error

[ssh_connection]
retries = 3
timeout = 10
__EOF__
    fi
}

function install_binray() {
    local ret=1
    local tempfile="/tmp/curveadm-$(date +%s%6N).tar.gz"
    curl $g_download_url -sLo $tempfile
    if [ $? -eq 0 ]; then
        tar -zxvf $tempfile -C $g_bin_dir 1>/dev/null
        ret=$?
    fi

    rm  $tempfile
    if [ $ret -eq 0 ]; then
        chmod 755 "$g_bin_dir/curveadm"
    else
        die "Download curveadm binray failed\n"
    fi
}

function set_profile() {
    shell=`echo $SHELL | awk 'BEGIN {FS="/";} { print $NF }'`
    if [ -f "${HOME}/.${shell}_profile" ]; then
        g_profile="${HOME}/.${shell}_profile"
    elif [ -f "${HOME}/.${shell}_login" ]; then
        g_profile="${HOME}/.${shell}_login"
    elif [ -f "${HOME}/.${shell}rc" ]; then
        g_profile="${HOME}/.${shell}rc"
    fi

    case :$PATH: in
        *:$g_bin_dir:*) ;;
        *) echo "export PATH=$g_bin_dir:\$PATH" >> $g_profile ;;
    esac
}

function print_success() {
    success "Install curveadm success, please run 'source $g_profile'\n"
}

function main() {
    setup
    install_binray
    set_profile
    print_success
}

############################  MAIN()
main "$@"
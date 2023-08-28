#!/usr/bin/env bash

############################  GLOBAL VARIABLES
g_color_yellow=`printf '\033[33m'`
g_color_red=`printf '\033[31m'`
g_color_normal=`printf '\033[0m'`
g_curveadm_home="$HOME/.curveadm"
g_bin_dir="$g_curveadm_home/bin"
g_http_bin_dir="$g_curveadm_home/http"
g_profile="${HOME}/.profile"
g_root_url="https://curveadm.nos-eastchina1.126.net/release"
g_latest_url="${g_root_url}/__version"
g_latest_version=$(curl -Is $g_latest_url | awk 'BEGIN {FS=": "}; /^x-nos-meta-curveadm-latest-version/{print $2}')
g_latest_version=${g_latest_version//[$'\t\r\n ']}
g_upgrade="$CURVEADM_UPGRADE"
g_version="${CURVEADM_VERSION:=$g_latest_version}"
g_download_url="${g_root_url}/curveadm-${g_version}.tar.gz"
g_plugin="$CURVEADM_PLUGIN"
g_plugin_dir="$g_curveadm_home/plugins/$g_plugin"
g_plugin_url="https://curveadm.nos-eastchina1.126.net/plugins/${g_plugin}-$(uname -m).tar.gz"

############################  BASIC FUNCTIONS
msg() {
    printf '%b' "$1" >&2
}

success() {
    msg "$g_color_yellow[✔]$g_color_normal ${1}${2}"
}

die() {
    msg "$g_color_red[✘]$g_color_normal ${1}${2}"
    exit 1
}

program_must_exist() {
    local ret='0'
    command -v $1 >/dev/null 2>&1 || { local ret='1'; }

    if [ "$ret" -ne 0 ]; then
        die "You must have '$1' installed to continue.\n"
    fi
}

############################ FUNCTIONS
backup() {
    if [ -d "$g_curveadm_home" ]; then
        mv $g_curveadm_home "${g_curveadm_home}-$(date +%s).backup"
    fi
}

setup() {
    mkdir -p $g_curveadm_home/{bin,data,plugins,logs,http/logs,http/conf,temp}

    # generate config file
    local confpath="$g_curveadm_home/curveadm.cfg"
    if [ ! -f $confpath ]; then
        cat << __EOF__ > $confpath
[defaults]
log_level = error
sudo_alias = "sudo"
timeout = 300
auto_upgrade = true
debug = false

[ssh_connections]
retries = 3
timeout = 10
__EOF__
    fi

    # generate http service config file
    local httpConfpath="$g_curveadm_home/http/conf/pigeon.yaml"
    if [ ! -f $httpConfpath ]; then
        cat << __EOF__ > $httpConfpath
servers:
  - name: curveadm
    log_level: info
    listen: :11000
__EOF__
    fi
}

install_binray() {
    local ret=1
    local tempfile="/tmp/curveadm-$(date +%s%6N).tar.gz"
    curl $g_download_url -skLo $tempfile
    if [ $? -eq 0 ]; then
        tar -zxvf $tempfile -C $g_curveadm_home --strip-components=1 1>/dev/null
        ret=$?
    fi

    rm  $tempfile
    if [ $ret -eq 0 ]; then
        chmod 755 "$g_bin_dir/curveadm" "$g_http_bin_dir/pigeon"
    else
        die "Download curveadm failed\n"
    fi
}

install_plugin() {
    local ret=1
    mkdir -p $g_plugin_dir
    local tempfile="/tmp/curveadm-plugin-$g_plugin-$(date +%s%6N).tar.gz"
    curl $g_plugin_url -sLo $tempfile
    if [ $? -eq 0 ]; then
        tar -zxvf $tempfile -C $g_plugin_dir --strip-components=1 1>/dev/null
        ret=$?
    fi

    rm  $tempfile
    if [ $ret -eq 0 ]; then
        success "Plugin '$g_plugin' installed\n"
    else
        die "Download plugin '$g_plugin' failed\n"
    fi
}

set_profile() {
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

print_install_success() {
    success "Install curveadm $g_version success, please run 'source $g_profile'\n"
}

print_upgrade_success() {
    if [ -f "$g_curveadm_home/CHANGELOG" ]; then
        cat "$g_curveadm_home/CHANGELOG"
    fi
    success "Upgrade curveadm to $g_version success\n"
}

install() {
    backup
    setup
    install_binray
    set_profile
    print_install_success
}

upgrade() {
    install_binray
    print_upgrade_success
}

main() {
    if [ ! -z $g_plugin ]; then
        install_plugin
    elif [ "$g_upgrade" == "true" ]; then
        upgrade
    else
        install
    fi
}

############################  MAIN()
main "$@"

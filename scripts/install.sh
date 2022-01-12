#!/usr/bin/env bash

############################  GLOBAL VARIABLES
g_upgrade="$CURVEADM_UPGRADE"
g_version="${CURVEADM_VERSION:=latest}"
g_curveadm_home="$HOME/.curveadm"
g_bin_dir="$g_curveadm_home/bin"
g_profile="${HOME}/.profile"
g_download_url="https://curveadm.nos-eastchina1.126.net/release/curveadm-${g_version}.tar.gz"
g_color_yellow=`printf '\033[33m'`
g_color_red=`printf '\033[31m'`
g_color_normal=`printf '\033[0m'`

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
setup() {
    mkdir -p $g_curveadm_home/{bin,data,logs,temp}

    local confpath="$g_curveadm_home/curveadm.cfg"
    if [ ! -f $confpath ]; then
        cat << __EOF__ > $confpath
[defaults]
log_level = error
sudo_alias = "sudo"

[ssh_connection]
retries = 3
timeout = 10
__EOF__
    fi
}

install_binray() {
    local ret=1
    local tempfile="/tmp/curveadm-$(date +%s%6N).tar.gz"
    curl $g_download_url -sLo $tempfile
    if [ $? -eq 0 ]; then
        tar -zxvf $tempfile -C $g_curveadm_home --strip-components=1 1>/dev/null
        ret=$?
    fi

    rm  $tempfile
    if [ $ret -eq 0 ]; then
        chmod 755 "$g_bin_dir/curveadm"
    else
        die "Download curveadm failed\n"
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
    success "Install curveadm success, please run 'source $g_profile'\n"
}

print_upgrade_success() {
    echo ""
    cat "$g_curveadm_home/CHANGELOG"
    echo ""
    success "Upgrade curveadm success\n"
}

install() {
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
    if [ $g_upgrade ] && [ $g_upgrade = "true" ]; then
        upgrade
    else
        install
    fi
}

############################  MAIN()
main "$@"

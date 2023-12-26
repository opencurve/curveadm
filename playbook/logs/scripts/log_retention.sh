#!/bin/bash

############################  GLOBAL VARIABLES
g_color_yellow=$(printf '\033[33m')
g_color_red=$(printf '\033[31m')
g_color_normal=$(printf '\033[0m')

############################  BASIC FUNCTIONS
msg() {
    printf '%b' "${1}" >&2
}

success() {
    msg "${g_color_yellow}[✔]${g_color_normal} ${1}"
}

die() {
    msg "${g_color_red}[✘]${g_color_normal} ${1}"
    exit 1
}
############################ FUNCTIONS
pre_check() {
    if [ -z "$(which logrotate)" ]; then
        die "logrotate could not be found"
    fi

    if [ -z "$(which jq)" ]; then
        die "jq could not be found"
    fi
}

load_config() {
    # Specify JSON file path
    json_file="config.json"

    # Check if the file exists
    if [ ! -f "$json_file" ]; then
        die "json file ${json_file} does not exist.\n"
    fi

    log_folder_path=$(jq -r '.log_folder_path' "$json_file")
    clean=$(jq -r '.clean' "$json_file")
    rotate=$(jq -r '.rotate' "$json_file")
    period=$(jq -r '.period' "$json_file")
    compress=$(jq -r '.compress' "$json_file")
    missingok=$(jq -r '.missingok' "$json_file")
    dateext=$(jq -r '.dateext' "$json_file")
    notifempty=$(jq -r '.notifempty' "$json_file")
    create=$(jq -r '.create' "$json_file")
    size=$(jq -r '.size' "$json_file")
}

delete_all_gz_files() {
    local folder="$1"
    find "$folder" -name "*.gz" -type f -exec rm {} \;
    if [ $? -eq 0 ]; then
        success "Clean all .gz files in $folder\n"
    fi
}

retention_log() {
    match_string=".log"
    find_files=$(find "$log_folder_path" -type f -name "*$match_string*" -not -name "*.gz")
    
    if [ -z "$find_files" ]; then
        die "Log file not found, please check whether the log file path is correct.\n"
    fi

    for file in $find_files; do
        dir_path=$(dirname -- "$file")
        dir_path_name=$(basename -- "$dir_path")

        if [ "$clean" == "true" ]; then
            delete_all_gz_files "$dir_path"
            continue
        fi

        if [ -e "/etc/logrotate.d/curve-$dir_path_name" ];then
            continue
        fi

        # Generate logrotate configuration
        logrotate_config="${dir_path}/*.log*[!gz] ${dir_path}/*.log {\n"
        logrotate_config+="    rotate ${rotate}\n"
        logrotate_config+="    ${period}\n"

        if [ "$compress" == "true" ]; then
            logrotate_config+="    compress\n"
        fi

        if [ "$missingok" == "true" ]; then
            logrotate_config+="    missingok\n"
        fi

        if [ "$notifempty" == "true" ]; then
            logrotate_config+="    notifempty\n"
        fi

        if [ "$dateext" == "true" ]; then
            logrotate_config+="    dateext\n"
        fi

        if [ "$create" != ""  ]; then
            logrotate_config+="    create ${create}\n"
        fi

        if [ "$size" != "" ]; then
            logrotate_config+="    size ${size}\n"
        fi

        logrotate_config+="}\n"

        # Write logrotate configuration to file
        echo -e "$logrotate_config" > "/etc/logrotate.d/curve-$dir_path_name"

        logrotate -d /etc/logrotate.conf 2>&1 | grep -i 'error:'
        if [ $? -eq 0 ]; then
            die "Logrotate configuration file error.\n"
        else
            success "Logrotate configuration generated and written to '/etc/logrotate.d/curve-$dir_path_name' file.\n"
        fi  
    done 

    logrotate /etc/logrotate.conf
    if [ $? -eq 0 ]; then
        success "Logrotate started successfully.\n"
    fi
}

main() {
    pre_check
    load_config
    retention_log
}

############################  MAIN()
main

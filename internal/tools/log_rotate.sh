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
    msg "${g_color_yellow}[✔]${g_color_normal} ${1}${2}"
}

die() {
    msg "${g_color_red}[✘]${g_color_normal} ${1}${2}"
    exit 1
}

############################ FUNCTIONS
load_config() {
    # Specify JSON file path
    json_file="config.json"

    # Check if the file exists
    if [ ! -f "$json_file" ]; then
        die "json file ${json_file} does not exist.\n"
    fi

    # Read variables in JSON file
    interval=$(jq -r '.interval' "$json_file")
    log_directory=$(jq -r '.logDirectory' "$json_file")
    compress_or_clean=$(jq -r '.compressOrClean' "$json_file")
    max_tar_file=$(jq -r '.maxTarFile' "$json_file")

}

precheck() {
    # Check whether interval is greater than or equal to 0
    if [[ ! ($interval =~ ^[0-9]+$) ]]; then
        die "The interval parameter is invalid.\n"
    elif [[ ! ($interval -ge 0) ]]; then
        die "The interval parameter is invalid.\n"
    fi

    # Check if log_directory is an absolute path
    absolute_path_regex="^\/([a-zA-Z0-9-]+\/)*[a-zA-Z0-9-]+\/?$"
    if ! [[ $log_directory =~ $absolute_path_regex ]]; then
        die "The log_directory parameter is invalid.\n"
    fi


     # Check if compress_or_clean is equal to 0 or equal to 1
    if [[ ! ($compress_or_clean =~ ^[0-9]+$) ]]; then
        die "The compress_or_clean parameter is invalid.\n"
    elif [[ ! ($compress_or_clean -eq 0 || $compress_or_clean -eq 1) ]]; then
        die "The compress_or_clean parameter is invalid.\n"
    fi

    # Check if max_tar_file is greater than or equal to 1
    if [[ ! ($max_tar_file =~ ^[0-9]+$) ]]; then
        die "The max_tar_file parameter is invalid.\n"
    elif [[ ! ($max_tar_file -ge 1) ]]; then
        die "The max_tar_file parameter is invalid.\n"
    fi
}

check_and_delete_tar_files() {
    local folder="$1"

    while true; do
        # Get the count of .tar.gz files
        local file_count=$(find "$folder" -maxdepth 1 -type f -name "*.tar.gz" | wc -l)

        # Check if file count exceeds the limit
        if [ "$file_count" -gt "$max_tar_file" ]; then
            # Delete the oldest file
            oldest_file=$(ls -t "$folder"/*.tar.gz | tail -n 1)
            rm "$oldest_file"
            msg "Deleted oldest tar file: $oldest_file\n"
        else
            msg "Tar file count within limit: $file_count\n"
            break
        fi
    done
}

delete_expired_tar_files() {
    local folder="$1"
    find "$folder" -name "*.tar.gz" -type f -mtime +$interval -exec rm {} \;
    if [ $? -eq 0 ]; then
        msg "Clean all expired tar files in $folder\n"
    fi
}

precompress_log() {
    # Determine whether the file exists and has content
    if [ -s "$file" ]; then
        # Get the modification time of a file
        modification_time=$(stat -c %Y "$file")

        # Get the timestamp of a specified date
        specified_timestamp=$(date -d "$day_before" +%s)

        # Determine whether the file modification time is before the specified date
        if [ "$modification_time" -lt "$specified_timestamp" ]; then
            cat "$file" > "$dir_path/log_before_$day_before.txt"
            > "$file"
            tar -czvf "$dir_path/log_before_$day_before.tar.gz" -C "$dir_path" "log_before_$day_before.txt"
            rm "$dir_path/log_before_$day_before.txt"
            check_and_delete_tar_files "$dir_path"
            precompress=true
        fi
    else
        nodata=true
    fi
}

retention_log() {
    match_string=".log"
    current_date=$(date +%s)
    day_before=$(date -d "$interval days ago" +%Y-%m-%d)
    find_files=$(find "$log_directory" -type f -name "*$match_string*")

    file_count=$(echo "$find_files" | wc -l)
    if [ $file_count -lt 1 ]; then
        die "Log file not found, please check whether the log file path is correct.\n"
    fi

    for file in $find_files; do
        # Get the file name and determine whether it is the target log
        filename=$(basename -- "$file")
        dir_path=$(dirname -- "$file")
        precompress=false
        nodata=false
        precompress_log

        if [ "$nodata" = true ]; then
            msg "No data in $file, skipping...\n\n"
            continue
        fi

        regex="("
        regex+=$day_before
        regex+=")"
        matched_line=$(grep -m 1 -P -n "$regex" "$file" | cut -d ':' -f 1)

        if [ "$compress_or_clean" -eq 1 ]; then
            # Perform compression operation
            if [[ ! -z "$matched_line" ]]; then
                # Extract the data before the matching line, use the tar command to package 
                # and compress the data, and delete the temporary files
                head -n "$((matched_line - 1))" "$file" > "$dir_path/log_before_$day_before.txt"
                tar -czvf "$dir_path/log_before_$day_before.tar.gz" -C "$dir_path" "log_before_$day_before.txt"
                rm "$dir_path/log_before_$day_before.txt"

                # Extract matching lines and subsequent data to filtered.txt, delete the original file, 
                # and rename filtered.txt to the original file name
                sed -n "$matched_line,\$p" "$file" > $dir_path/filtered.txt
                rm "$file"
                mv $dir_path/filtered.txt "$file"

                check_and_delete_tar_files "$dir_path"
                success "Compress log in $file\n\n"
            elif [ "$precompress" = true ]; then
                success "Compress log in $file\n\n"
            else
                msg "No data was matched in $file. The following are the first 10 lines of the log file. Please change the date and try deleting again:\n" 
                head "$file"
                msg "Compress log failed. Try setting the interval parameter to 0.\n\n"
            fi
        else
            # Perform cleanup operations
            delete_expired_tar_files "$dir_path"
            if [[ ! -z "$matched_line" ]]; then
                sed -n "$matched_line,\$p" "$file" > $dir_path/filtered.txt
                rm "$file"
                mv $dir_path/filtered.txt "$file"
                msg "Clean expired log in $file\n\n"
            elif [ "$precompress" = true ]; then
                success "Clean expired log in $file\n\n"
            else
                msg "No data was matched in $file. The following are the first 10 lines of the log file. Please change the date and try deleting again:\n" 
                head "$file"
                msg "Compress log failed. Try setting the interval parameter to 0.\n\n"

            fi
        fi 

    done    
}

main() {
    load_config
    precheck
    retention_log
}

############################  MAIN()
main


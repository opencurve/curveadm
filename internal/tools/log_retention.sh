#!/bin/bash

check_date=true
has_h_option=false
default_directory="/tmp/curvefs/logs" # Absolute path to the default log folder
match_string=".log"

while getopts ":d:hp:s:" opt; do
  case $opt in
    d) 
      d=$OPTARG
      date_regex="^[0-9]{4}-[0-9]{2}-[0-9]{2}$"
      if [[ $d =~ $date_regex ]]; then
        echo "Logs after $OPTARG will be retained" >&2
      else
        echo "The data parameter is invalid, please use -h to query the correct format." >&2
        exit 1
      fi
      ;;
    h) 
      has_h_option=true
      ;;
    p) 
      default_directory=$OPTARG
      absolute_path_regex="^\/([a-zA-Z0-9]+\/)*[a-zA-Z0-9]+\/?$"
      if [[ $default_directory =~ $absolute_path_regex ]]; then
        echo "The log file absolute path is set to $OPTARG." >&2
      else
        echo "The path parameter is invalid, please use -h to query the correct format." >&2
      fi
      ;;
    s) 
      s=$OPTARG
      number_regex="^[0-9]+$"
      if [[ $s =~ $number_regex ]]; then
        echo "The loop will retain logs within $OPTARG days and ignore the data parameter." >&2
        check_date=false
      else
        echo "The sleep parameter is invalid, please use -h to query the correct format." >&2
        exit 1
      fi
      ;;
    \?)
      echo "Invalid parameter：$OPTARG" >&2
      exit 1
      ;;
  esac
done

if $has_h_option; then
  echo "help information："
  echo "Use -d <2023-11-11> in the command line argument to specify the date."
  echo "Use -h to get help information."
  echo "Use -p </c/Users/czy/Desktop/curve-master/files> to specify the log file path."
  echo "Use -s <number of days> to specify the cycle time and ignore the data parameter."
  exit 0
fi

# If there is no input parameter s
if $check_date; then
  if [ -z "$d" ]; then
    current_date=$(date +'%Y-%m-%d')
    two_days_before=$(date -d '2 days ago' +'%Y-%m-%d')
    d=$two_days_before
    echo "The data parameter is not provided, and the logs 2 days ($two_days_before) before the current date ($current_date) are saved by default." >&2
  fi

  current_timestamp=$(date -d "$d" +%s)
  current_timestamp_now=$(date +%s)

  # The input date must be less than or equal to the current date
  if [[ "$current_timestamp" -le "$current_timestamp_now" ]]; then  
    find_files=$(find "$default_directory" -type f)
    # Record the number of logs processed
    count=0
    for file in $find_files; do
      # Get the file name and determine whether it is the target log
      filename=$(basename -- "$file")
      if [[ $filename == *"$match_string"* ]]; then
        echo "$file"
        count=$((count + 1))

        # Delete all logs before the first match to the target date
        regex="("
        regex+=$d
        regex+=")"
        matched_line=$(grep -m 1 -P -n "$regex" "$file" | cut -d ':' -f 1)
        if [[ ! -z "$matched_line" ]]; then
          sed -n "$matched_line,\$p" "$file" > filtered.txt
          rm "$file"
          mv filtered.txt "$file"
        else
          echo "No data was matched. The following are the first 10 lines of the log file. Please change the date and try deleting again:" >&2
          head "$file"
        fi
      
      fi
    done

    if [ $count -lt 1 ]; then
      echo "Log file not found, please check whether the log file path is correct." >&2
      exit 1
    fi

  else
    echo "$d later than current time."
    exit 1
  fi

fi

# If you enter parameter s
if ! $check_date; then
  # Set sleep time (days)
  s_in_seconds=$((s * 24 * 60 * 60))
  # Define termination signal handling function
  function on_terminate {
    echo "A termination signal is received and the script exits."
    exit 0
  }

  # Set the termination signal processing function
  trap on_terminate SIGINT SIGTERM

  # Loop infinitely until a termination signal is received
  while true; do
    d=$(date -d "$s days ago" +%Y-%m-%d)
    echo "Logs after $d will be retained."

    find_files=$(find "$default_directory" -type f)
    count=0
    for file in $find_files; do
      # Get the file name and determine whether it is the target log
      filename=$(basename -- "$file")
      if [[ $filename == *"$match_string"* ]]; then
        echo "$file"
        count=$((count + 1))

        # Delete all logs before the first match to the target date
        regex="("
        regex+=$d
        regex+=")"
        matched_line=$(grep -m 1 -P -n "$regex" "$file" | cut -d ':' -f 1)
        if [[ ! -z "$matched_line" ]]; then
          sed -n "$matched_line,\$p" "$file" > filtered.txt
          rm "$file"
          mv filtered.txt "$file"
        fi

      fi    
    done
     
    if [ $count -lt 1 ]; then
      echo "Log file not found, please check whether the log file path is correct." >&2
      exit 1
    fi

    sleep $s_in_seconds
  done
fi
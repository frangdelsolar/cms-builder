#!/bin/bash

total_log_count=0

# Function to print a log with color based on log level
print_log() {
  local file="$1"
  local line_number="$2"
  local line="$3"
  local log_level="$4"

  case "$log_level" in
    "Debug")
      color="\033[0;32m"  # Green
      ;;
    "Error")
      color="\033[0;31m"  # Red
      ;;
    "Trace")
      color="\033[0;34m"  # Blue
      ;;
    "Info")
      color="\033[0;36m"  # Cyan
      ;;
    "Warn")
      color="\033[0;33m"  # Yellow
      ;;
    *)
      color="\033[0m"  # Default
      ;;
  esac

  echo "${color}$file:$line_number: $line\033[0m"
}


# Function to find logs in a file and return the count
find_logs() {
  local file="$1"
  local log_pattern="log\.(Debug|Error|Trace|Info|Warn)\(\)"  
  local line_number=0
  local log_count=0

  while read line; do
    line_number=$((line_number + 1))
    if [[ $line =~ $log_pattern ]]; then
      log_level="${BASH_REMATCH[1]}"
      # print_log "$file" "$line_number" "$line" "$log_level"
      echo "$file:$line_number: $log_level"
      log_count=$((log_count + 1))
    fi
  done < "$file"
}

# Find all files in the repository
files=$(find ../builder -type f)

# Counter for logs found
# Iterate through each file and find logs
for file in $files; do
  # echo "\033[0;17mChecking file: $file\033[0m"
  echo "Checking file: $file"
  find_logs "$file"
done
# Print the total number of logs found
echo "Total logs found: $total_log_count"

# Warn the user if there are too many logs
if [[ $total_log_count -gt 100 ]]; then
  echo "WARNING: Too many logs found. Consider reducing the log level or cleaning up old logs."
fi
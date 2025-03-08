#!/bin/bash

# Check if the script received exactly one argument (the file path)
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path-to-file>"
    exit 1
fi

# Assign the argument to a variable
file_path="$1"

# Check if the file exists and is readable
if [ ! -f "$file_path" ]; then
    echo "Error: File '$file_path' not found!"
    exit 1
fi

# Read the contents of the file
file_contents=$(cat "$file_path")

# Make the POST request using curl
curl 'http://localhost:8090/verify' \
    -X POST \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
    --data-urlencode "version=1.0" \
    --data-urlencode "withVet=true" \
    --data-urlencode "body=$file_contents"

#!/bin/bash

PORT=${PORT:-8090}
ENDPOINT="http://localhost:$PORT/verify"

if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <path-to-file>"
    exit 1
fi

file_path="$1"

if [ ! -f "$file_path" ]; then
    echo "Error: File '$file_path' not found!"
    exit 1
fi

file_contents=$(cat "$file_path")

curl "$ENDPOINT" \
    -X POST \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/x-www-form-urlencoded; charset=UTF-8' \
    --data-urlencode "version=1.0" \
    --data-urlencode "body=$file_contents"

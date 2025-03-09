#!/usr/bin/env sh

eval export "$(cat .env)"
docker run --name gobra-playground -p 8090:"$PORT" -d \
    -e "PORT=$PORT" \
    -e "JAVA_PATH=$JAVA_PATH" \
    -e "GOBRA_PATH=$GOBRA_PATH" \
    gobra-playground

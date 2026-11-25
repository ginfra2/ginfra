#!/usr/bin/env bash

# This file will not be updated, so any additions or changes are safe.

if [ $# -lt 1 ]; then
    echo "Not enough arguments."
    echo "  ARG1 - SERVICE ROOT PATH"
    exit 1
fi

#bin/start.sh $1

# Run it
build/ginfra_service.exe ${1}


#!/usr/bin/env bash

# This file will not be updated, so any additions or changes are safe.

if [ $# -lt 1 ]; then
    echo "Not enough arguments."
    echo "  ARG1 - SERVICE ROOT PATH"
    exit 1
fi

# Run it
build/ginfra_service.exe ${1} ${2} ${3} ${4} ${5} ${6} ${7} ${8} ${9}


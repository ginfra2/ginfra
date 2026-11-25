#!/usr/bin/env bash

if [ $# -lt 1 ]; then
    echo "Not enough arguments."
    echo "  ARG1 - SERVICE ROOT PATH"
    exit 1
fi

# Run it
ginfra_stub local ${1} ${2} ${3} ${4} ${5} ${6} ${7} ${8} ${9}


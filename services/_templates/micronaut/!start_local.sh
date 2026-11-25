#!/usr/bin/env bash

# This file will not be updated, so any additions or changes are safe.

if [ $# -lt 1 ]; then
    echo "Not enough arguments."
    echo "  ARG1 - SERVICE ROOT PATH"
    exit 1
fi

# Run prep
java -jar assets/ginfraapp-1.0.0.jar config props

# Set environment.
if [ -f ./ENV ] ; then
  set -a
  . ./ENV
  set +a
fi

# Run it
java -javaagent:assets/opentelemetry-javaagent.jar -jar build/libs/<<<ServiceName>>>-0.1-all.jar ${1} ${2} ${3} ${4} ${5} ${6} ${7} ${8} ${9}


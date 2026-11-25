#!/usr/bin/env bash

# This file will not be updated and can be changed.

# This should be called from the repo root.

if [ ! -f service_config.json ] ; then
    cp build/service_config.json.NEW ./service_config.json
fi

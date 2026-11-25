#!/usr/bin/env bash

# This file will not be updated and can be changed.

# This should be called from the repo root.

if [ ! -f postgres_url ] ; then
    echo "postgres://postgres:YOURPASSWORD@192.168.1.220:5432" >  postgres_url
fi

if [ ! -f service_config.json ] ; then
    cp build/service_config.json.NEW ./service_config.json
fi

#!/usr/bin/env bash

DEPLOYMENT="default"
NAME="<<<ServiceName>>>"
PORT="8080"

if [ $# -lt 1 ]; then
    echo "Not enough arguments."
    echo "  ARG1 - SERVICE ROOT PATH"
    exit 1
fi

# HAX Get the config.
DEPLOYMENT=$(grep -Po '"deployment":.*?[^\\]",' service_config.json | cut -d ":" -f 2- | tr -d '",')
NAME=$(grep -Po '"name":.*?[^\\]",' service_config.json | cut -d ":" -f 2- | tr -d '",')
# Add port to the specifics or something.
CURI=$(grep -Po '"config_host_uri":.*?[^\\]",' service_config.json | cut -d ":" -f 2- | tr -d '",')
CTOKEN=$(grep -Po '"config_access_token":.*?[^\\]",' service_config.json | cut -d ":" -f 2- | tr -d '",')

# Setup .env
echo "VITE_GINFRA_CONFIG_URI=${CURI}" > .env
echo "VITE_GINFRA_ACCESS_TOKEN=${CTOKEN}" >> .env
echo "VITE_GINFRA_DEPLOYMENT=${DEPLOYMENT}" >> .env
echo "VITE_GINFRA_NAME=${NAME}" >> .env
echo "VITE_GINFRA_PORT=${PORT}" >> .env

# Run it
npm run dev


#!/usr/bin/env bash

apt-get update && \
  apt-get install jq nginx -y

mkdir -p /service/log


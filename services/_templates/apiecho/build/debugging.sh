#!/usr/bin/env bash

########################################################################################
# Convenience script to install packages that help with debugging.
# This file will not be updated, so any additions or changes are safe.

apt update -y
apt install -y procps iproute2 wget less vim lsof net-tools

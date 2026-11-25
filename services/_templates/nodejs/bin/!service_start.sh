#!/usr/bin/env bash

##########################################################################
# Server start
# This file will not be updated so any changes are safe.

node server.js >> log/server.log 2>&1 &

#!/usr/bin/env bash

##########################################################################
# Server restart
# This file will not be updated so any changes are safe.

killall node
sleep 1
bin/service_start.sh $1

#!/bin/bash

source ${CATTLE_HOME:-/var/lib/cattle}/common/scripts.sh

cd $(dirname $0)

chmod +x bin/telemetry

mkdir -p content-home
mv bin content-home

stage_files

# Make sure that when node start is doesn't think it holds the config.sh lock
unset CATTLE_CONFIG_FLOCKER

if /etc/init.d/telemetry status; then
    /etc/init.d/telemetry restart
else
    /etc/init.d/telemetry start
fi

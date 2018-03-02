#!/bin/bash

set -eou pipefail

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function finish() {
    echo -e -n $COLOR_RESET
}
trap finish EXIT

echo -e -n $COLOR_GRAY

DEBUG=${DEBUG:-no}
if [ "yes" == "$DEBUG" ]; then
    set -x
fi

function log() {
    set +x
    echo -e "$COLOR_BLUE[$(date +"%F %T")] _dev-reload-server $COLOR_RED|$COLOR_RESET" $@$COLOR_GRAY
    if [ "yes" == "$DEBUG" ] ; then
        set -x
    fi
}

function kill_aptomi() {
    pkill -f "aptomi server" || true
}

if make install ; then
    kill_aptomi
    log "Reloaded aptomi server"
    aptomi server &
else
    log "Failed to build aptomi, killed aptomi server"
    kill_aptomi
fi

#!/bin/bash

set -eou pipefail

echo "APTOMI_DB=$APTOMI_DB"

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function finish() {
# todo source _dev... and use kill_aptomi from it
#    kill_aptomi
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
    echo -e "$COLOR_BLUE[$(date +"%F %T")] dev-watch-server $COLOR_RED|$COLOR_RESET" $@$COLOR_GRAY
    if [ "yes" == "$DEBUG" ] ; then
        set -x
    fi
}

log "Watching repo and reloading aptomi server after any changes outside of webui folder"

if ! hash fswatch 2>/dev/null; then
    log "App fswatch isn't installed but required"
    log "\tOn macOS just run: brew install fswatch"
    exit 1
fi

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"
pushd $workdir 1>/dev/null

log "Workdir: $workdir"

tools/dev-reload-server.sh
fswatch -o -l 1 ./cmd ./pkg Makefile ./tools | xargs -n1 -I "{}" tools/dev-reload-server.sh

popd



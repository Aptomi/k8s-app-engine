#!/bin/bash

set -eou pipefail

echo "APTOMI_DB=$APTOMI_DB"

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function kill_aptomi() {
    pkill -f "aptomi server" || true
}

function finish() {
    kill_aptomi
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

watched_path="cmd pkg Makefile tools vendor"

log "Watching repo and reloading aptomi server after any changes in: $watched_path"

if ! hash fswatch 2>/dev/null; then
    log "App fswatch isn't installed but required"
    log "\tOn macOS just run: brew install fswatch"
    exit 1
fi

workdir="$( cd "$( dirname "${BASH_SOURCE[0]}" )/../" && pwd )"
pushd $workdir 1>/dev/null

log "Workdir: $workdir"

tools/dev-reload-server.sh
fswatch -o -l 1 $watched_path | xargs -n1 -I "{}" tools/dev-reload-server.sh

popd



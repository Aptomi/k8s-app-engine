#!/bin/bash

set -eou pipefail

#APTOMI_DB

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
    echo -e "$COLOR_BLUE[$(date +"%F %T")] demo-watch-apply $COLOR_RED|$COLOR_RESET" $@$COLOR_GRAY
    if [ "yes" == "$DEBUG" ] ; then
        set -x
    fi
}

repo="git@github.com:Frostman/aptomi-demo.git"
interval=1

#git log -n 1 --pretty=format:"%H"

tmp=$(mktemp -d)

pushd $tmp 1>/dev/null

    log "Working dir: $tmp"

    git clone $repo . 2>/dev/null

    log "Polling repo $repo each $interval seconds and auto-applying policy"
    log ""
    log "Legend:\t_ - printed each $interval seconds if no changes in repo"
    log ""

    current_head=""

    while sleep $interval ; do
        git fetch origin

        remote_head=$(git log -n 1 --pretty=format:"%H" origin/master)

        if [ "$current_head" != "$remote_head" ]; then
            log "Remote repo changed"

            git reset --hard origin/master

            mkdir -p $APTOMI_DB/policy
            cp -r ./* $APTOMI_DB/policy

            log "Running aptomi policy apply --noop to ensure that policy isn't broken"
            if aptomi policy apply --noop ; then
                log "Running aptomi policy apply"
                if ! aptomi policy apply ; then
                    log "aptomi policy apply failed, will re-try next time"
                else
                    log "aptomi policy apply success, switched to new HEAD"
                    current_head=$remote_head
                fi
            else
                log "aptomi policy apply --noop failed, will re-try next time"
            fi

            echo ""
        else
            echo -n "_"
        fi

    done

popd

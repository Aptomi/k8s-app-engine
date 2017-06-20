#!/bin/bash

set -e

#APTOMI_DB

repo="git@github.com:Frostman/aptomi-demo.git"
interval=1

#git log -n 1 --pretty=format:"%H"

tmp=$(mktemp -d)

pushd $tmp 1>/dev/null

    echo "Working dir: $tmp"

    git clone $repo . 2>/dev/null

    echo "Polling repo $repo each $interval seconds and auto-applying policy"
    echo -e "\nLegend:\n\t_ - printed each $interval seconds if no changes in repo\n"

    current_head=""

    while sleep $interval ; do
        git fetch origin

        remote_head=$(git log -n 1 --pretty=format:"%H" origin/master)

        if [ "$current_head" != "$remote_head" ]; then
            echo "Remote repo changed"

            git reset --hard origin/master

            mkdir -p $APTOMI_DB/policy
            cp -r ./* $APTOMI_DB/policy

            if aptomi policy apply --noop ; then
                if ! aptomi policy apply ; then
                    echo "aptomi policy apply failed, will re-try next time"
                else
                    echo "aptomi policy apply success, switched to new HEAD"
                    current_head=$remote_head
                fi
            else
                echo "aptomi policy apply --noop failed, will re-try next time"
            fi

        else
            echo -n "_"
        fi

    done

popd

#!/bin/bash

set -ex

tmp=$(mktemp -d)

cp -r ./policy/* $tmp/
pushd $tmp
    git init
    git add -A
    git add -f _external/charts/*.tgz
    git commit -a -m "Initial demo state $(date)"
    git remote add origin git@github.com:Aptomi/aptomi-demo.git
    git push -f origin master
popd

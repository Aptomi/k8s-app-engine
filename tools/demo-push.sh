#!/bin/bash

set -ex

tmp=$(mktemp -d)

cp -r demo/* $tmp/
pushd $tmp
    git init
    git add -A
    git commit -a -m "Initial demo state $(date)"
    git remote add origin git@github.com:Frostman/aptomi-demo.git
    git push -f origin master
popd

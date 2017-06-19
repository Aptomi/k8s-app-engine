#!/bin/bash

set -ex

# Build & install the latest aptomi binary
make install

# Reset aptomi policy
aptomi policy reset --force

# Sync demo policy
git clone git@github.com:Frostman/aptomi-demo.git $APTOMI_DB/aptomi-demo

# Run aptomi policy apply in noop mode
aptomi policy apply --noop

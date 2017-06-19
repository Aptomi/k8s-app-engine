#!/bin/bash

set -ex

# Build & install the latest aptomi binary
make install

# Reset aptomi policy
aptomi policy reset --force

# Sync demo policy
git clone https://github.com/Frostman/aptomi-demo $APTOMI_DB/aptomi-demo

# Run aptomi policy apply in noop mode
aptomi policy apply --noop

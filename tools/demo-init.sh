#!/bin/bash

set -ex

# Build & install the latest aptomi binary
make install

# Reset aptomi policy
aptomi policy reset --force

# Sync demo policy
./tools/demo-sync.sh

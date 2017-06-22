#!/bin/bash

set -exou pipefail

echo "APTOMI_DB=$APTOMI_DB"

# Build & install the latest aptomi binary
make install

# Reset aptomi policy
aptomi policy reset --force

# Sync demo policy
cp -r ./demo/ $APTOMI_DB/policy

# Run aptomi policy apply in noop mode
aptomi policy apply --noop

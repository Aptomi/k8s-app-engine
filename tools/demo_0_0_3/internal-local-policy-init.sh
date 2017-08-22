#!/bin/bash

set -exou pipefail

echo "APTOMI_DB=$APTOMI_DB"

# Reset aptomi policy
./aptomi policy reset --force

# Sync demo policy
cp -r ./policy/ $APTOMI_DB/policy

# Run aptomi policy apply in noop mode
./aptomi policy apply --noop

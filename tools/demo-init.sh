#!/bin/bash

set -exou pipefail

echo "APTOMI_DB=$APTOMI_DB"

# Init local database
./tools/local-policy-init.sh

# Push demo policy to remote github repo
./tools/demo-push.sh

# Start watcher/puller from that remote github repo
./tools/demo-watch-apply.sh

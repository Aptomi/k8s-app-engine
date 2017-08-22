#!/bin/bash

set -exou pipefail

echo "APTOMI_DB=$APTOMI_DB"

# Change directory to the directory of the script
cd "$(dirname "$0")"

# Init local database with demo policy
./internal-local-policy-init.sh

# Push demo policy to remote github repo
./internal-policy-push.sh

# Start watcher/puller from that remote github repo
./internal-watch-apply.sh

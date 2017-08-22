#!/bin/bash

set -exou pipefail

echo "APTOMI_DB=$APTOMI_DB"

# Change directory to the directory of the script
cd "$(dirname "$0")"

./aptomi server

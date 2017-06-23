#!/bin/bash

set -eou pipefail

echo "APTOMI_DB=$APTOMI_DB"

grep -rl '^enabled: false' $APTOMI_DB/policy/ | xargs sed -i '' 's/^enabled: false/enabled: true/g'
grep -rl '^  enabled: false' $APTOMI_DB/policy/ | xargs sed -i '' 's/^  enabled: false/  enabled: true/g'

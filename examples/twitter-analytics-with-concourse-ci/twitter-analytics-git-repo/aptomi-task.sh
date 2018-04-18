#!/bin/sh

set -eou pipefail

echo "Running aptomi"
echo "Action: ${ACTION}"
echo "File path: ${FILEPATH}"

export PATH=$PATH:/aptomi
APTOMI_CLIENT_CONFIG_DIR="$HOME/.aptomi"

# Create config
mkdir ${APTOMI_CLIENT_CONFIG_DIR}
cat >${APTOMI_CLIENT_CONFIG_DIR}/config.yaml <<EOL
debug: true

api:
  host: host.docker.internal
  port: 27866
EOL

# Apply policy
aptomictl version
aptomictl login -u admin -p admin
aptomictl policy ${ACTION} --wait -f ${FILEPATH}

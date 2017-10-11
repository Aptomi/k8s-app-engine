#!/bin/bash

set -ex

# Clean left over pid file
pidFile=${APACHEDS_INSTANCE_DIRECTORY}/run/apacheds-${APACHEDS_INSTANCE}.pid
[[ -e $pidFile ]] && rm $pidFile

# Execute the server in console mode and not as a daemon.
exec ./apacheds console ${APACHEDS_INSTANCE}

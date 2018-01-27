#!/bin/bash

# This script completely removes Aptomi from the host, including:
# - Aptomi server & its config
# - Aptomi client & its config
# - Aptomi database
#
# YOU WILL LOSE ALL DATA IN APTOMI AFTER RUNNING THIS SCRIPT

SCRIPT_NAME=`basename "$0"`

COLOR_GRAY='\033[0;37m'
COLOR_BLUE='\033[0;34m'
COLOR_RED='\033[0;31m'
COLOR_RESET='\033[0m'

function log() {
    echo -e "$COLOR_BLUE[$(date +"%F %T")] $SCRIPT_NAME $COLOR_RED|" $@$COLOR_GRAY
}

function run_as_root() {
  local CMD="$*"

  if [ $EUID -ne 0 ]; then
    CMD="sudo $CMD"
  fi

  $CMD
}

run_as_root killall aptomi
run_as_root rm -f /usr/local/bin/aptomi
run_as_root rm -f /usr/local/bin/aptomictl
run_as_root rm -rf /etc/aptomi
run_as_root rm -rf ~/.aptomi
run_as_root rm -rf /var/lib/aptomi
log "Aptomi binaries deleted and all data erased"
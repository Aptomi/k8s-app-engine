#!/usr/bin/env bash

set -eou pipefail

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source "${DIR}/util.sh"

function cleanup() {
    set +x
    etcd::stop
}

trap cleanup EXIT

etcd::start

export APTOMI_DB_ENDPOINTS=127.0.0.1:${ETCD_PORT}
export APTOMI_TEST_DB_ENDPOINTS=${APTOMI_DB_ENDPOINTS}

set -x
$@
set +x

#!/bin/bash

if ! [ -x "$(command -v telnet)" ]; then
  echo 'telnet is not installed' >&2
  exit 1
fi

if ! [ -x "$(command -v jq)" ]; then
  echo 'jq is not installed' >&2
  exit 1
fi

set -eou pipefail

export PATH=${GOPATH}/bin:$PATH
DEBUG=${DEBUG:-no}
DEBUG_MODE=false
if [ "yes" == "$DEBUG" ]; then
    DEBUG_MODE=true
    set -x
fi

FAILED=1
CONF_DIR=$(mktemp -d)
POLICY_DIR=$(mktemp -d)
POLICY_DIR_TMP=$(mktemp -d)

# copy policy over, create clusters from templates
cp -R examples/twitter-analytics/* $POLICY_DIR
cp ${POLICY_DIR}/policy/clusters/clusters.{yaml.template,yaml}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
source ${DIR}/util.sh

function cleanup() {
    stop_server
    rm -rf ${CONF_DIR}
    etcd::stop
}

trap cleanup EXIT

function stop_server() {
    echo "Stopping server..."
    kill ${SERVER_PID} &>/dev/null || true

    # there should only be several errors in the server log (one related to alice not having permissions, and another one related to carol not having rights to instantiate services)
    errors=$(grep "error" "${CONF_DIR}/server.log" | grep -v "doesn't have ACL permissions to manage object" | grep -v "do not allow claim" || true)
    if [ "$errors" != "" ]; then
        echo "Found unexpected errors"
        echo "$errors"
        echo "Smoke tests failed"
        exit 1
    fi

    echo "No errors found in server logs"
    if [ "yes" == "$DEBUG" ]; then
        # print the entire server log in debug mode
        [[ -e "${CONF_DIR}/server.log" ]] && awk '{print "[[SERVER]] " $0}' ${CONF_DIR}/server.log || echo "No server log found."
    else
        # if smoke tests failed, then print last 100 lines of the server log
        if [ $FAILED -eq 1 ]; then
            [[ -e "${CONF_DIR}/server.log" ]] && tail -n 100 ${CONF_DIR}/server.log | awk '{print "[[SERVER]] " $0}' || echo "No server log found."
            echo "See last 100 lines of the server log above"
        fi
    fi

    if [ $FAILED -eq 1 ]; then
        echo "Smoke tests failed"
    else
        echo "Smoke tests passed successfully"
    fi
}

etcd::start

APTOMI_PORT=$(util::free_port)

cat >${CONF_DIR}/config.yaml <<EOL
debug: ${DEBUG_MODE}

api:
  host: 127.0.0.1
  port: ${APTOMI_PORT}

db:
  prefix: smoke
  endpoints:
    - localhost:${ETCD_PORT}

enforcer:
  disabled: false
  noop: true

updater:
  disabled: false
  noop: true

domainAdminOverrides:
  Sam: true

users:
  ldap:
    - host: localhost
      port: 10389
      basedn: "o=aptomiOrg"
      filter: "(&(objectClass=organizationalPerson))"
      filterbyname: "(&(objectClass=organizationalPerson)(cn=%s))"
      labeltoattributes:
        name: cn
        description: description
        global_ops: isglobalops
        is_operator: isoperator
        mail: mail
        team: team
        org: o
        short-description: role
        deactivated: deactivated
EOL

aptomi server --config ${CONF_DIR} &>${CONF_DIR}/server.log &
SERVER_PID=$!

echo "Server PID: ${SERVER_PID}"

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server failed to start"
    exit 1
fi

WAIT_FLAGS="--wait --wait-time 10s"

function login() {
    aptomictl --config ${CONF_DIR} login --username $1 --password $1
}

function check_policy() {
    expected="$1"
    query="$2"

    login sam
    actual="$(aptomictl --config ${CONF_DIR} policy show -o json | jq "$2")"

    if [ "$actual" -eq "$expected" ]; then
        echo "Found value is equal to expected $actual for query $query"
        return 0
    fi

    echo "Expected value $expected but found $actual for $query"
    return 1
}

function check_policy_version() {
    check_policy $1 .Metadata.Generation
}

function change_policy() {
    cmd="$1"
    files="$2"
    expectedVersion="$3"

    # run in noop mode
    aptomictl --config ${CONF_DIR} policy ${cmd} --noop ${WAIT_FLAGS} ${files}

    # run in normal mode
    aptomictl --config ${CONF_DIR} policy ${cmd} ${WAIT_FLAGS} ${files}
    check_policy_version ${expectedVersion}
}

function check_claims() {
    files="$1"
    aptomictl --config ${CONF_DIR} claim status --wait ${files}
}

login alice
if aptomictl --config ${CONF_DIR} policy apply -f ${POLICY_DIR}/policy &>/dev/null ; then
    echo "Alice shouldn't be able to upload full policy"
    exit 1
fi

check_policy_version 1

login sam
change_policy apply "-f ${POLICY_DIR}/policy/rules -f ${POLICY_DIR}/policy/clusters" 2

# adding the same clusters should not result in an error
change_policy apply "-f ${POLICY_DIR}/policy/clusters" 2

login frank
change_policy apply "-f ${POLICY_DIR}/policy/analytics_pipeline" 3

login john
change_policy apply "-f ${POLICY_DIR}/policy/twitter_stats" 4

login john
change_policy apply "-f ${POLICY_DIR}/policy/john-prod-ts.yaml" 5
check_claims "-f ${POLICY_DIR}/policy/john-prod-ts.yaml"

login alice
change_policy apply "-f ${POLICY_DIR}/policy/alice-dev-ts.yaml" 6
check_claims "-f ${POLICY_DIR}/policy/alice-dev-ts.yaml"

login bob
change_policy apply "-f ${POLICY_DIR}/policy/bob-dev-ts.yaml" 7
check_claims "-f ${POLICY_DIR}/policy/bob-dev-ts.yaml"

check_policy 3 ".Objects.social.claim | length"
check_claims "-f ${POLICY_DIR}/policy/john-prod-ts.yaml -f ${POLICY_DIR}/policy/alice-dev-ts.yaml -f ${POLICY_DIR}/policy/bob-dev-ts.yaml"

# delete Alice's claim
login alice
change_policy delete "-f ${POLICY_DIR}/policy/alice-dev-ts.yaml" 8
check_policy 2 ".Objects.social.claim | length"

# upgrade prod claim
sed -e 's/demo11/demo12/g' ${POLICY_DIR}/policy/john-prod-ts.yaml > ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
login john
change_policy apply "-f ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml" 9

# apply Carol's claim
login carol
change_policy apply "-f ${POLICY_DIR}/policy/carol-dev-ts.yaml" 10
check_policy 3 ".Objects.social.claim | length"

# delete all claims
login sam
change_policy delete "-f \"${POLICY_DIR}/policy/*-ts.yaml\"" 11
check_policy 0 ".Objects.social.claim | length"

# delete the rest of the objects
login sam
change_policy delete "-f ${POLICY_DIR}/policy" 12
check_policy 0 ".Objects.platform.service | length"
check_policy 0 ".Objects.social.service | length"
check_policy 0 ".Objects.platform.claim | length"
check_policy 0 ".Objects.social.claim | length"
check_policy 0 ".Objects.platform.rule | length"
check_policy 0 ".Objects.social.rule | length"
check_policy 0 ".Objects.platform.bundle | length"
check_policy 0 ".Objects.social.bundle | length"
check_policy 0 ".Objects.system.aclrule | length"
check_policy 0 ".Objects.system.cluster | length"

# import all objects at once
login sam
change_policy apply "-f ${POLICY_DIR}/policy" 13

# check object counts
check_policy 4 ".Objects.social.claim | length"
check_policy 5 ".Objects.platform.service | length"
check_policy 0 ".Objects.platform.claim | length"
check_policy 5 ".Objects.platform.bundle | length"
check_policy 1 ".Objects.social.service | length"
check_policy 1 ".Objects.social.bundle | length"
check_policy 1 ".Objects.platform.rule | length"
check_policy 1 ".Objects.system.rule | length"
check_policy 2 ".Objects.social.rule | length"
check_policy 3 ".Objects.system.aclrule | length"
check_policy 1 ".Objects.system.cluster | length"

# check claims
check_claims "-f ${POLICY_DIR}/policy/john-prod-ts.yaml -f ${POLICY_DIR}/policy/alice-dev-ts.yaml -f ${POLICY_DIR}/policy/bob-dev-ts.yaml"

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server not running after all tests"
    exit 1
fi

FAILED=0

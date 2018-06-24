#!/bin/bash

if ! [ -x "$(command -v telnet)" ]; then
  echo 'telnet is not installed' >&2
  exit 1
fi

set -eou pipefail

export PATH=${GOPATH}/bin:$PATH
DEBUG=${DEBUG:-no}
DEBUG_MODE=false

# Handle debug mode, if specified in environment variables
if [ "yes" == "$DEBUG" ]; then
    DEBUG_MODE=true
    set -x
fi

# Handle profiling mode, if specified
CPU_PROFILE=${CPU_PROFILE:-""}

# Handle tracing mode, if specified
TRACE_PROFILE=${TRACE_PROFILE:-""}

CONF_DIR=$(mktemp -d)

# Use existing bolt.db as a starting point, if specified in environment variables
DB=${DB:-""}
if [ -f "$DB" ]; then
    cp "$DB" "${CONF_DIR}/db.bolt"
fi

POLICY_DIR=$(mktemp -d)
POLICY_DIR_TMP=$(mktemp -d)

# copy policy over, create clusters from templates
cp -R examples/twitter-analytics/* $POLICY_DIR
cp ${POLICY_DIR}/policy/clusters/clusters.{yaml.template,yaml}

function cleanup() {
    stop_server
    rm -rf ${CONF_DIR}
}

trap cleanup EXIT

function free_port() {
    for port in $(seq 10000 11000); do
        echo -ne "\035" | telnet 127.0.0.1 $port > /dev/null 2>&1;
        [ $? -eq 1 ] && echo "$port" && break;
    done
}

function stop_server() {
    echo "Stopping server..."
    kill ${SERVER_PID} &>/dev/null || true

    if [ "yes" == "$DEBUG" ]; then
        [[ -e "${CONF_DIR}/server.log" ]] && awk '{print "[[SERVER]] " $0}' ${CONF_DIR}/server.log || echo "No server log found."
    fi
}

APTOMI_PORT=$(free_port)

cat >${CONF_DIR}/config.yaml <<EOL
debug: ${DEBUG_MODE}

api:
  host: 127.0.0.1
  port: ${APTOMI_PORT}

db:
  connection: ${CONF_DIR}/db.bolt

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

SERVER_OPTIONS=""
if [ ! -z "$CPU_PROFILE" ]; then
    SERVER_OPTIONS+=" --cpuprofile ${CPU_PROFILE}"
fi
if [ ! -z "$TRACE_PROFILE" ]; then
    SERVER_OPTIONS+=" --traceprofile ${TRACE_PROFILE}"
fi

aptomi server ${SERVER_OPTIONS} --config ${CONF_DIR} &>${CONF_DIR}/server.log &
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

function change_policy() {
    cmd="$1"
    files="$2"

    # run in normal mode
    aptomictl --config ${CONF_DIR} policy ${cmd} ${WAIT_FLAGS} ${files}
}

login sam

change_policy apply "-f ${POLICY_DIR}/policy/rules -f ${POLICY_DIR}/policy/clusters"
change_policy apply "-f ${POLICY_DIR}/policy/analytics_pipeline"
change_policy apply "-f ${POLICY_DIR}/policy/twitter_stats"

while :
do
    change_policy apply "-f ${POLICY_DIR}/policy/john-prod-ts.yaml"
    change_policy apply "-f ${POLICY_DIR}/policy/alice-dev-ts.yaml"
    change_policy apply "-f ${POLICY_DIR}/policy/bob-dev-ts.yaml"
    change_policy apply "-f ${POLICY_DIR}/policy/carol-dev-ts.yaml"

    # upgrade prod claim
    sed -e 's/demo11/demo12/g' ${POLICY_DIR}/policy/john-prod-ts.yaml > ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
    change_policy apply "-f ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml"

    # delete all claims
    change_policy delete "-f ${POLICY_DIR}/policy/john-prod-ts.yaml"
    change_policy delete "-f ${POLICY_DIR}/policy/alice-dev-ts.yaml"
    change_policy delete "-f ${POLICY_DIR}/policy/bob-dev-ts.yaml"
    change_policy delete "-f ${POLICY_DIR}/policy/carol-dev-ts.yaml"
done

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server not running after all tests"
    exit 1
fi

echo "Stress tests successfully passed"

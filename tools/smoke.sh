#!/bin/bash

if ! [ -x "$(command -v telnet)" ]; then
  echo 'telnet is not installed' >&2
  exit 1
fi

if ! [ -x "$(command -v jq)" ]; then
  echo 'jq is not installed' >&2
  exit 1
fi

set -exou pipefail

CONF_DIR=$(mktemp -d)
POLICY_DIR=$(mktemp -d)
POLICY_DIR_TMP=$(mktemp -d)

# copy policy over, create secrets and clusters from templates
cp -R examples/03-twitter-analytics/* $POLICY_DIR
cp ${POLICY_DIR}/_external/secrets/secrets.yaml.template ${CONF_DIR}/secrets.yaml
cp ${POLICY_DIR}/policy/Sam/clusters.{yaml.template,yaml}

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
    [[ -e "${CONF_DIR}/server.log" ]] && echo "Server log location: ${CONF_DIR}/server.log" || echo "No server log found."
}

APTOMI_PORT=$(free_port)

cat >${CONF_DIR}/config.yaml <<EOL
debug: true

api:
  host: 127.0.0.1
  port: ${APTOMI_PORT}

db:
  connection: ${CONF_DIR}/db.bolt

enforcer:
  noop: true
  interval: 0.1s

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

secretsDir: ${CONF_DIR}
EOL

aptomi server --config ${CONF_DIR} &>${CONF_DIR}/server.log &
SERVER_PID=$!

echo "Server PID: ${SERVER_PID}"

sleep 3

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server failed to start"
    exit 1
fi

if aptomictl policy --username Alice --config ${CONF_DIR} apply -f ${POLICY_DIR}/policy &>/dev/null ; then
    echo "Alice shouldn't be able to upload policy"
    exit 1
fi

function check_policy_version() {
    expected="$1"

    actual="$(aptomictl policy show --username Sam --config ${CONF_DIR} -o json | jq .Metadata.Generation)"

    if [ "$actual" -eq "$expected" ]; then
        echo "Found policy version is equal to expected $actual"
        return 0
    fi

    echo "Expected policy version $expected but found $actual"
    return 1
}

WAIT_FLAGS="--wait --wait-interval 0.1s --wait-attempts 10"

# apply full policy (w/o Carol)
check_policy_version 1
aptomictl policy apply ${WAIT_FLAGS} --username Sam --config ${CONF_DIR} -f ${POLICY_DIR}/policy/Sam
check_policy_version 2
aptomictl policy apply ${WAIT_FLAGS} --username Frank --config ${CONF_DIR} -f ${POLICY_DIR}/policy/Frank
check_policy_version 3
aptomictl policy apply ${WAIT_FLAGS} --username John --config ${CONF_DIR} -f ${POLICY_DIR}/policy/John
check_policy_version 4
aptomictl policy apply ${WAIT_FLAGS} --username John --config ${CONF_DIR} -f ${POLICY_DIR}/policy/john-prod-ts.yaml
check_policy_version 5
aptomictl policy apply ${WAIT_FLAGS} --username Alice --config ${CONF_DIR} -f ${POLICY_DIR}/policy/alice-stage-ts.yaml
check_policy_version 6
aptomictl policy apply ${WAIT_FLAGS} --username Bob --config ${CONF_DIR} -f ${POLICY_DIR}/policy/bob-stage-ts.yaml
check_policy_version 7

# delete Alice's dependency
aptomictl policy delete ${WAIT_FLAGS} --username Alice --config ${CONF_DIR} -f ${POLICY_DIR}/policy/alice-stage-ts.yaml
check_policy_version 8

# upgrade prod dependency
sed -e 's/demo11/demo12/g' ${POLICY_DIR}/policy/john-prod-ts.yaml > ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
aptomictl policy apply ${WAIT_FLAGS} --username John --config ${CONF_DIR} -f ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
check_policy_version 9

# apply Carol's dependency
aptomictl policy apply ${WAIT_FLAGS} --username Carol --config ${CONF_DIR} -f ${POLICY_DIR}/policy/carol-stage-ts.yaml
check_policy_version 10

# delete all dependencies
aptomictl policy delete ${WAIT_FLAGS} --username Sam --config ${CONF_DIR} -f "${POLICY_DIR}/policy/*-ts.yaml"
check_policy_version 11

# delete all definitions
aptomictl policy delete ${WAIT_FLAGS} --username Sam --config ${CONF_DIR} -f ${POLICY_DIR}/policy
check_policy_version 12

aptomictl policy show --username Sam --config ${CONF_DIR} -o json | jq .

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server not running after all tests"
    exit 1
fi

echo "Smoke tests successfully passed"

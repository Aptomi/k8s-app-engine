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

APTOMI="${GOPATH}/bin/aptomi"
APTOMICTL="${GOPATH}/bin/aptomictl"
DEBUG=${DEBUG:-no}
DEBUG_MODE=false
if [ "yes" == "$DEBUG" ]; then
    DEBUG_MODE=true
    set -x
fi

CONF_DIR=$(mktemp -d)
POLICY_DIR=$(mktemp -d)
POLICY_DIR_TMP=$(mktemp -d)

# copy policy over, create clusters from templates
cp -R examples/twitter-analytics/* $POLICY_DIR
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

${APTOMI} server --config ${CONF_DIR} &>${CONF_DIR}/server.log &
SERVER_PID=$!

echo "Server PID: ${SERVER_PID}"

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server failed to start"
    exit 1
fi

function login() {
    ${APTOMICTL} --config ${CONF_DIR} login --username $1 --password $1
}

login alice
if ${APTOMICTL} --config ${CONF_DIR} policy apply -f ${POLICY_DIR}/policy &>/dev/null ; then
    echo "Alice shouldn't be able to upload full policy"
    exit 1
fi

function check_policy() {
    expected="$1"
    query="$2"

    login sam
    actual="$(${APTOMICTL} --config ${CONF_DIR} policy show -o json | jq "$2")"

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

WAIT_FLAGS="--wait --wait-attempts 20"

# apply full policy (w/o Carol)
check_policy_version 1

login sam
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/Sam
check_policy_version 2

login frank
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/Frank
check_policy_version 3

login john
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/John
check_policy_version 4

login john
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/john-prod-ts.yaml
check_policy_version 5

login alice
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/alice-stage-ts.yaml
check_policy_version 6

login bob
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/bob-stage-ts.yaml
check_policy_version 7

check_policy 3 ".Objects.social.dependency | length"

# delete Alice's dependency
login alice
${APTOMICTL} --config ${CONF_DIR} policy delete ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/alice-stage-ts.yaml
check_policy_version 8
check_policy 2 ".Objects.social.dependency | length"

# upgrade prod dependency
sed -e 's/demo11/demo12/g' ${POLICY_DIR}/policy/john-prod-ts.yaml > ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
login john
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR_TMP}/john-prod-ts-changed.yaml
check_policy_version 9

# apply Carol's dependency
login carol
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/carol-stage-ts.yaml
check_policy_version 10
check_policy 3 ".Objects.social.dependency | length"

# delete all dependencies
login sam
${APTOMICTL} --config ${CONF_DIR} policy delete ${WAIT_FLAGS} -f "${POLICY_DIR}/policy/*-ts.yaml"
check_policy_version 11
check_policy 0 ".Objects.social.dependency | length"

# delete all definitions
login sam
${APTOMICTL} --config ${CONF_DIR} policy delete ${WAIT_FLAGS} -f ${POLICY_DIR}/policy
check_policy_version 12
check_policy 0 ".Objects.platform.contract | length"
check_policy 0 ".Objects.social.contract | length"
check_policy 0 ".Objects.platform.dependency | length"
check_policy 0 ".Objects.social.dependency | length"
check_policy 0 ".Objects.platform.rule | length"
check_policy 0 ".Objects.social.rule | length"
check_policy 0 ".Objects.platform.service | length"
check_policy 0 ".Objects.social.service | length"
check_policy 0 ".Objects.system.aclrule | length"
check_policy 0 ".Objects.system.cluster | length"

login sam
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/Sam
check_policy_version 13

login frank
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/Frank
check_policy_version 14

login john
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/John
check_policy_version 15

login john
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/john-prod-ts.yaml
check_policy_version 16

login alice
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/alice-stage-ts.yaml
check_policy_version 17

login bob
${APTOMICTL} --config ${CONF_DIR} policy apply ${WAIT_FLAGS} -f ${POLICY_DIR}/policy/bob-stage-ts.yaml
check_policy_version 18

check_policy 5 ".Objects.platform.contract | length"
check_policy 0 ".Objects.platform.dependency | length"
check_policy 5 ".Objects.platform.service | length"
check_policy 1 ".Objects.social.contract | length"
check_policy 3 ".Objects.social.dependency | length"
check_policy 1 ".Objects.social.service | length"
check_policy 3 ".Objects.system.rule | length"
check_policy 3 ".Objects.system.aclrule | length"
check_policy 2 ".Objects.system.cluster | length"

sleep 1

SERVER_RUNNING=`ps | grep aptomi | grep "${SERVER_PID}" || true`
if [ -z "$SERVER_RUNNING" ]; then
    echo "Server not running after all tests"
    exit 1
fi

echo "Smoke tests successfully passed"

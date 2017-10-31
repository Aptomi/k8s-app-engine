#!/bin/bash

if ! [ -x "$(command -v telnet)" ]; then
  echo 'telnet is not installed' >&2
  exit 1
fi

set -eou pipefail

CONF_DIR=$(mktemp -d)
POLICY_DIR="examples/03-twitter-analytics"

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
    kill ${SERVER_PID} || true
    [[ -e "${CONF_DIR}/server.log" ]] && awk '{print "[[SERVER]] " $0}' ${CONF_DIR}/server.log || echo "No server log found."
    [[ -e "${CONF_DIR}/client.log" ]] && awk '{print "[[CLIENT]] " $0}' ${CONF_DIR}/client.log || echo "No client log found."
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
  disabled: true

domainAdminOverrides:
  Sam: true

users:
  ldap:
    - host: localhost
      port: 10389
      basedn: o=aptomiOrg
      filter: (&(objectClass=organizationalPerson))
      labeltoattributes:
        id: dn
        name: cn
        description: description
        global_ops: isglobalops
        is_operator: isoperator
        mail: mail
        team: team
        org: o
        short-description: role
        deactivated: deactivated

secretsDir: ${POLICY_DIR}
EOL

aptomi server --config ${CONF_DIR} &>${CONF_DIR}/server.log &
SERVER_PID=$!

echo "Server PID: ${SERVER_PID}"

sleep 3

if aptomictl policy --username Alice --config ${CONF_DIR} apply -f ${POLICY_DIR}/policy &>/dev/null ; then
    echo "Alice shouldn't be able to upload policy"
    exit 1
fi

aptomictl policy --username Sam --config ${CONF_DIR} apply -f ${POLICY_DIR}/policy &>${CONF_DIR}/client.log

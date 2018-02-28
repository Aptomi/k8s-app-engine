#!/usr/bin/env bash

if ! [ -x "$(command -v telnet)" ]; then
  echo 'telnet is not installed' >&2
  exit 1
fi

if ! [ -x "$(command -v jq)" ]; then
  echo 'jq is not installed' >&2
  exit 1
fi

function aptomi::login() {
#    aptomictl --config ${CONF_DIR} login --username $1 --password $1
    aptomictl login --username $1 --password $1
}

function assert_equal() {
    local expected="$1"
    local actual="$2"
    if [ "$expected" -eq "$actual" ]; then
        echo "EQUAL"
    else
        echo "assert_equal failed:\n expected: $expected\n actual: $actual"
        return 1
    fi
}

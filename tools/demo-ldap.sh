#!/bin/bash
set -ex

docker rm -f aptomi-ldap-demo || true
docker run --name aptomi-ldap-demo -d -p 10389:10389 aptomi/ldap-demo:latest

#!/bin/bash
docker rm -f aptomi-ldap-demo
docker run --name aptomi-ldap-demo -d -p 10389:10389 aptomi-ldap-demo:latest

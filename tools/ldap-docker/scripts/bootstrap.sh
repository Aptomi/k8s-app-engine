#!/bin/bash

set -ex

wait_for_ldap ()
{
  echo "Waiting for LDAP server to start"
  until ldapsearch -h localhost -p 10389 -D uid=admin,ou=system -w secret >/dev/null
  do
    sleep 1
  done
}

stop_ldap ()
{
  ./apacheds stop default
}

start_ldap ()
{
  ./apacheds start default
  wait_for_ldap
}

restart_ldap ()
{
  stop_ldap
  start_ldap
}

# When a fresh data folder is detected then bootstrap the instance configuration.
if [ ! -d ${APACHEDS_INSTANCE_DIRECTORY} ]; then
    mkdir ${APACHEDS_INSTANCE_DIRECTORY}
    cp -rv ${APACHEDS_BOOTSTRAP}/* ${APACHEDS_INSTANCE_DIRECTORY}
    chown -v -R ${APACHEDS_USER}:${APACHEDS_GROUP} ${APACHEDS_INSTANCE_DIRECTORY}
fi

# Start LDAP server
start_ldap

# Create aptomi partition & schema
ldapmodify -h localhost -p 10389 -D uid=admin,ou=system -w secret -c -a -f /aptomi-org.ldif
ldapmodify -h localhost -p 10389 -D uid=admin,ou=system -w secret -c -a -f /aptomi-schema.ldif

# Restart LDAP server
restart_ldap

# Create aptomi users
ldapmodify -c -h localhost -p 10389 -D uid=admin,ou=system -w secret -f /aptomi-users.ldif

# Stop LDAP server
stop_ldap

#!/bin/bash

VERDIR='2.0.0-M24'
VER='apacheds-2.0.0-M24'

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
  ./$VER/bin/apacheds.sh stop
}

start_ldap ()
{
  ./$VER/bin/apacheds.sh start
  wait_for_ldap
}

restart_ldap ()
{
  stop_ldap
  start_ldap
}

# Kill previous versions if they are still running
pkill -f apacheds

# Download LDAP server
if [ ! -f $VER.tar.gz ]; then
  wget http://apache.mirrors.hoobly.com//directory/apacheds/dist/$VERDIR/$VER.tar.gz
fi

# Remove previous version and unpack
rm -rf ./$VER/
tar -zxf $VER.tar.gz

# Start LDAP server
start_ldap

# Create aptomi partition
ldapmodify -h localhost -p 10389 -D uid=admin,ou=system -w secret -c -a -f aptomi-org.ldif

# Restart LDAP server
restart_ldap

# Create aptomi schema
ldapmodify -h localhost -p 10389 -D uid=admin,ou=system -w secret -c -a -f aptomi-schema.ldif

# Restart LDAP server
restart_ldap

# Create aptomi users
ldapmodify -c -h localhost -p 10389 -D uid=admin,ou=system -w secret -f aptomi-users.ldif

# What is it
Aptomi demo LDAP server containing pre-baked user data imported from /data

## Build
    docker build . -t aptomi-ldap-demo:latest

## Run
    docker run --name aptomi-ldap-demo -d -p 10389:10389 aptomi-ldap-demo:latest

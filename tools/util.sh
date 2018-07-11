#!/usr/bin/env bash

function util::free_port() {
    for port in $(seq 10000 11000); do
        echo -ne "\035" | telnet 127.0.0.1 $port > /dev/null 2>&1;
        [ $? -eq 1 ] && echo "$port" && break;
    done
}

function etcd::start() {
    etcd::stop
    export ETCD_PORT=$(util::free_port)
    docker run --name aptomi-etcd-smoke -d -p ${ETCD_PORT}:2379 quay.io/coreos/etcd /usr/local/bin/etcd --advertise-client-urls http://0.0.0.0:${ETCD_PORT} --listen-client-urls http://0.0.0.0:2379
    sleep 1
}

function etcd::stop() {
    docker rm -f aptomi-etcd-smoke || true
}

#!/bin/bash
set -exou pipefail

rm -rf ${PWD}/.aptomi-install-cache

function test_install {
    os="$1"
    cmd="$2"

    docker run --rm \
                    -e DEBUG=yes \
                    -p 27866:27866 \
                    -v ${PWD}/.aptomi-install-cache:/root/.aptomi-install-cache \
                    -v ${PWD}/scripts:/scripts \
                    -w /scripts \
                    aptomi/aptomi-test-install:${os} \
                    sh -c "${cmd}"
}

test_install xenial "./aptomi_install.sh && ./aptomi_uninstall_and_clean.sh"
test_install xenial "./aptomi_install.sh --with-example && ./aptomi_uninstall_and_clean.sh"

test_install centos7 "./aptomi_install.sh && ./aptomi_uninstall_and_clean.sh"
test_install centos7 "./aptomi_install.sh --with-example && ./aptomi_uninstall_and_clean.sh"

echo "All install scripts successfully verified"

#!/bin/bash
set -exou pipefail

sudo rm -rf /usr/local/bin/*aptomi* /etc/aptomi /var/lib/aptomi ~/.aptomi

hostname

ls -la

export PATH=$PATH:"$WORKSPACE"/bin
export GOPATH="$WORKSPACE"
export GOROOT=/usr/local/go
export PATH=$GOPATH/bin:$GOROOT/bin:$PATH
mkdir -p "$GOPATH/bin"

pushd src/github.com/Aptomi/aptomi

make vendor

tools/demo-ldap.sh

make lint

DEBUG=yes make smoke

source /jenkins/aptomi-coveralls.io
make coverage-full coverage-publish

tools/test-install.sh

sudo rm -rf "${PWD}/.aptomi-install-cache"

tools/publish-docker.sh

docker rmi $(docker images --filter "dangling=true" -q --no-trunc) || true
docker rmi $(docker images | grep "none" | awk '/ / { print $3 }') || true

tools/publish-charts.sh

popd

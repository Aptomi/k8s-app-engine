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

mkdir "$WORKSPACE/_log"
make vendor &>"$WORKSPACE/_log/vendor.txt"

tools/demo-ldap.sh

make lint

DEBUG=yes make smoke

source /jenkins/aptomi-coveralls.io
#make coverage-full coverage-publish

tools/test-install.sh

sudo rm -rf "${PWD}/.aptomi-install-cache"

GIT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
if [[ "${GIT_BRANCH}" == "master" ]]; then
    tools/publish-docker.sh
    tools/publish-charts.sh
fi

docker rmi $(docker images --filter "dangling=true" -q --no-trunc) || true
docker rmi $(docker images | grep "none" | awk '/ / { print $3 }') || true

popd

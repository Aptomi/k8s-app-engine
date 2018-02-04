#!/bin/bash
set -eou pipefail

VERSION=$(git describe --tags --long)

docker build -t aptomi/aptomi:master -f Dockerfile .
docker build -t aptomi/aptomictl:master -f Dockerfile.client .

docker push aptomi/aptomi:master
docker push aptomi/aptomictl:master

docker tag aptomi/aptomi:master aptomi/aptomi:${VERSION}
docker tag aptomi/aptomictl:master aptomi/aptomictl:${VERSION}

docker push aptomi/aptomi:${VERSION}
docker push aptomi/aptomictl:${VERSION}

#!/bin/bash
set -eou pipefail

docker build -t aptomi/aptomi:master -f Dockerfile .
docker build -t aptomi/aptomictl:master -f Dockerfile.client .

docker push aptomi/aptomi:master
docker push aptomi/aptomictl:master

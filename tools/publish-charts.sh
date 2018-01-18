#!/usr/bin/env bash

set -eux

CHARTS=${PWD}/charts
DIST=${PWD}/charts-dist

#helm repo add aptomi http://aptomi.io/charts
rm -rf ${DIST} || true
mkdir ${DIST}
#git clone git@github.com:Aptomi/charts.git ${DIST}

for dir in ${CHARTS}/*/; do
    pushd ${dir}
        echo "Building ${dir}"
        rm -rf ./charts
        helm dep up
    popd
done

pushd ${DIST}
    for dir in ${CHARTS}/*/; do
        helm package ${dir}
    done

    helm repo index --url http://aptomi.io/charts .
    git add -f *tgz index.yaml
    git commit -a -m "Charts updated at $(date)"
    git push
popd

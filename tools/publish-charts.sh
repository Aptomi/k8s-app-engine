#!/usr/bin/env bash

if [ $# -eq 1 ]; then
    org="$1"
fi

set -eux

CHARTS=${PWD}/charts
DIST=${CHARTS}/dist


for repo in examples; do
    helm repo add aptomi-${repo} http://aptomi.io/charts/${repo}
    rm -rf ${DIST} || true
    git clone git@github.com:Aptomi/charts.git ${DIST}
    for dir in charts/${repo}/*/; do
        pushd ${dir}
            echo "Building ${dir}"
            rm -rf ./charts
            helm dep up
        popd
    done
    pushd ${DIST}/${repo}
        for dir in ${CHARTS}/${repo}/*/; do
            helm package ${dir}
        done

        helm repo index --url http://aptomi.io/charts/examples --merge index.yaml .
        git add -f *tgz index.yaml
        git commit -a -m "Charts '${repo}' updated at $(date)"
        git push
    popd
done

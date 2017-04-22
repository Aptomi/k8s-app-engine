#!/bin/bash
set -e

if ! hash kubectl 2>/dev/null; then
    echo "kubectl is not installed"
    exit 1
fi

if ! hash helm 2>/dev/null; then
    echo "helm is not installed"
    exit 1
fi

NAMESPACE=$1
if [[ -z $NAMESPACE ]]; then
    echo ""
    echo "Namespace is not specified, usage: deploy.sh <namespace>"
    echo ""
    exit 1
fi

workdir=$(dirname $0)

helm repo add aptomi-charts https://f001.backblazeb2.com/file/aptomi-charts
helm repo update

helm="helm upgrade --install --namespace $NAMESPACE"

$helm kafka-1 aptomi-charts/kafka -f $workdir/configs/kafka.yaml
$helm spark-1 aptomi-charts/spark -f $workdir/configs/spark.yaml
$helm hdfs-1 aptomi-charts/hdfs

kubectl -n $NAMESPACE get svc

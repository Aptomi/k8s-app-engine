#!/bin/bash

set -ex

helm delete --purge $(helm list --all -q)
kubectl delete ns aptomi

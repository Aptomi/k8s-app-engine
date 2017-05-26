#!/bin/bash

set -ex

helm delete --purge $(helm list --all -q) || true
kubectl delete ns demo

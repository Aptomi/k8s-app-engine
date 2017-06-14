#!/bin/bash

set -ex

# Build and install aptomi binary
make install

# Reset aptomi policy
aptomi policy reset --force

# Add global objects
aptomi policy add cluster demo/clusters
aptomi policy add rules demo/rules.global.yaml
aptomi policy add users demo/users.yaml
aptomi policy add secrets demo/secrets.yaml
aptomi policy add chart demo/charts

# Add istio service (system-level)
aptomi policy add service demo/services/istio
aptomi policy add context demo/services/istio

# Add analytics pipeline service
aptomi policy add service demo/services/analytics-pipeline
aptomi policy add context demo/services/analytics-pipeline

# Add twitter stats service
aptomi policy add service demo/services/twitter-stats
aptomi policy add context demo/services/twitter-stats

# Add dependencies
aptomi policy add dependencies demo/dependencies/dependencies.alice-prod.yaml

# Show
aptomi policy apply --noop

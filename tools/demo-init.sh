#!/bin/bash

set -ex

# Reset aptomi policy
aptomi policy reset --force

# Add global objects
aptomi policy add cluster demo/clusters
aptomi policy add rules demo/rules
aptomi policy add users demo/users
aptomi policy add secrets demo/secrets
aptomi policy add chart demo/charts

# Add istio service
aptomi policy add service demo/services/istio
aptomi policy add context demo/services/istio

# Add analytics pipeline service
aptomi policy add service demo/services/analytics_pipeline
aptomi policy add context demo/services/analytics_pipeline

# Add twitter stats service
aptomi policy add service demo/services/twitter_stats
aptomi policy add context demo/services/twitter_stats

# Add dependencies
aptomi policy add dependencies demo/dependencies/dependencies.frank-prod-ts.yaml

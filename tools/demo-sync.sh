#!/bin/bash

set -ex

# Add global objects
aptomi policy add cluster demo/clusters
aptomi policy add rules demo/rules
aptomi policy add users demo/users
aptomi policy add secrets demo/secrets
aptomi policy add chart demo/charts

# Add service mesh istio service (system-level)
aptomi policy add service demo/services/istio
aptomi policy add context demo/services/istio

# John defines and operates analytics_pipeline service (defines all contexts)
aptomi policy add service demo/services/analytics_pipeline
aptomi policy add context demo/services/analytics_pipeline

# Frank defines and operates twitter_stats service (defines all contexts)
aptomi policy add service demo/services/twitter_stats
aptomi policy add context demo/services/twitter_stats

# Add initial dependency (Frank runs production instance of twitter_stats)
aptomi policy add dependencies demo/dependencies/dependencies.frank-prod-ts.yaml

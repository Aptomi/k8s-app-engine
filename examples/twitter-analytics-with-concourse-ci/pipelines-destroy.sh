#!/bin/bash

fly -t tutorial login
fly -t tutorial sync
fly -t tutorial destroy-pipeline -n -p apply-rules
fly -t tutorial destroy-pipeline -n -p dev-test-changes
fly -t tutorial destroy-pipeline -n -p update-prod

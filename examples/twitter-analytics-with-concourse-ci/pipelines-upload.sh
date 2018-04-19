#!/bin/bash

REPO_URI=$1

if [[ -z $REPO_URI ]];
then
    echo "Missing mandatory argument"
    echo "     Usage:  ./pipelines-upload.sh [git-repo-URI]"
    echo "   Example:  ./pipelines-upload.sh git@github.com:username/repo.git"
    exit 1
fi

set -x

fly -t tutorial login
fly -t tutorial sync

fly -t tutorial set-pipeline -n -c twitter-analytics-git-repo/pipeline-dev-test-changes.yml -l twitter-analytics-git-repo/pipeline-params.yml -v repo=$REPO_URI -p dev-test-changes
fly -t tutorial unpause-pipeline -p dev-test-changes

fly -t tutorial set-pipeline -n -c twitter-analytics-git-repo/pipeline-apply-rules.yml -l twitter-analytics-git-repo/pipeline-params.yml -v repo=$REPO_URI -p apply-rules
fly -t tutorial unpause-pipeline -p apply-rules

fly -t tutorial set-pipeline -n -c twitter-analytics-git-repo/pipeline-update-prod.yml -l twitter-analytics-git-repo/pipeline-params.yml -v repo=$REPO_URI -p update-prod
fly -t tutorial unpause-pipeline -p update-prod

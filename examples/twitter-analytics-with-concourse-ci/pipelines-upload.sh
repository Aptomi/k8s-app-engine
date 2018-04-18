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

fly -t tutorial set-pipeline -n -c twitter-analytics-git-repo/pipeline-dev-to-prod.yml -l twitter-analytics-git-repo/pipeline-params.yml -v repo=$REPO_URI -p dev-to-prod
fly -t tutorial unpause-pipeline -p dev-to-prod

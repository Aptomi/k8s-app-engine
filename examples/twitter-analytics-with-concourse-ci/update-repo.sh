#!/bin/bash

REPO_URI=$1

if [[ -z $REPO_URI ]];
then
    echo "Missing mandatory argument"
    echo "     Usage:  ./update-repo.sh [git-repo-URI]"
    echo "   Example:  ./update-repo.sh git@github.com:username/repo.git"
    exit 1
fi

set -x

rm -rf ./tmp-repo
mkdir tmp-repo
cp -R twitter-analytics-git-repo/* tmp-repo/
cp -R ../twitter-analytics/policy tmp-repo/

cd tmp-repo
git init
git remote add origin $REPO_URI
git add .
git commit -m "initial commit"
git push --force --mirror

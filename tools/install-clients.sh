#!/bin/bash

set -exou pipefail

brew uninstall kubernetes-cli || true
brew uninstall kubernetes-helm || true

sudo cp tools/clients/kubectl /usr/local/bin/
sudo cp tools/clients/helm /usr/local/bin/
sudo cp tools/clients/istioctl /usr/local/bin/

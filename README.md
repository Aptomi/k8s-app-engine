# Aptomi

[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi - make test)](https://ci.aptomi.io/job/aptomi%20-%20make%20test/)

## What can Aptomi do for you
...

## How to build and run Aptomi
Bulding Aptomi from source and running integration tests is a very straightforward process. All you need is Go (ideally 1.9.1) and a couple of external packages:
* graphviz - so that Aptomi can generate diagrams via [GraphViz](http://www.graphviz.org/Download..php)
* telnet - for the script which runs smoke tests
* docker - to run provided LDAP server with sample user data

If you are on macOS, install brew, install [Docker For Mac](https://docs.docker.com/docker-for-mac/install/) and run: 
```
brew install graphviz telnet docker
```

Install Helm, Istio, Kubectl clients:
```
./tools/install-clients.sh`
```

All Go dependencies are managed using [Glide](https://glide.sh/). The following command will fetch all dependencies and put them into "vendor" folder:
```
make vendor 
```

To build the binary:
```
make 
```

## User Guide

### How to start provided LDAP Server with sample data
1. In order to run Aptomi examples and smoke tests, there is an LDAP Server with sample users in a docker container. It's very easy to build and run it, just follow the instructions in [ldap-docker](tools/ldap-docker/README.md)
2. It's also recommended to download and install [Apache Directory Studio](http://directory.apache.org/studio/) for browsing LDAP. Once installed, follow these [step-by-step instructions](http://directory.apache.org/apacheds/basic-ug/1.4.2-changing-admin-password.html) to connect

### Running examples
...

## Dev Guide

### Running tests

Command    | Target          | LDAP Required
-----------|-----------------|--------------
```make test```    | Unit tests | No
```make alltest``` | Integration + Unit tests | Yes
```make smoke```   | Smoke tests + Integration + Unit tests | Yes

### How to develop
Command     | Target          | Description
------------|-----------------|--------------
```make fmt```  | Format code | Re-formats all code according to Go standards
```make lint``` | Examine code | Run linters to examine Go source code and reports suspicious constructs

## Provided scripts
* `./tools/demo-gke.sh` - to set up 2 k8s clusters on GKE for demo. supports `up`, `down`, or `status`

## How to set up demo environment on Google Cloud
1. ```brew install kubernetes-cli kubernetes-helm```
  * Do we still need this, in addition to `./tools/install-clients.sh`?
1. ```curl https://sdk.cloud.google.com | bash```
1. Create new project in https://console.cloud.google.com/
1. ```gcloud auth login```
1. ```gcloud config set project <YOUR_PROJECT_ID>```
1. https://console.cloud.google.com/ -> API Manager -> Enable API
  1. Google Container Engine API
  1. Google Compute Engine API
1. ```./tools/gke-demo.sh up```
1. ```./tools/gke-demo.sh status```
1. Run demo (see README_DEMO.md for instructions)
1. ```./tools/gke-demo.sh down``` - don't forget to destroy your clusters, so you don't continue to get billed for them

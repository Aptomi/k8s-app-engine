![aptomi Logo](aptomi-logo.png)

[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi - make test)](https://ci.aptomi.io/job/aptomi%20-%20make%20test/)

TODO: add slack, email? web site?

Aptomi is a project that simplifies roll-out, operation and control of microservice-based applications. Policy-based service exchange enables individual teams to put their microservices on auto pilot and control service consumption, while providing insights and contextual visibility.

Aptomi is particularly useful in environments with multiple teams, clouds, environments and data centers, where policy-based management plays an important role in operating large decentralized infrastructure.

Aptomi’s primary focus is Docker and Kubernetes, but it can work on any environment regardless of container runtime and container orchestration technologies.

TODO: Picture, where it's in the stack

## Table of contents
TODO: use https://github.com/thlorenz/doctoc

## Features
- **Easy way to deploy and manage complex applications**
  - TODO: link to a diagram of a complex application
- **Run on top of any container platform**
  - k8s, OpenShift, AWS ECS, GKE, Docker Datacenter, Mesos
- **Friendly for Dev and Ops teams**
  - Keep using your existing application templates (Helm, k8s YAMLs, ksonnet, Ansible, etc)
  - Speak services, not containers. Collaborate between organizations and rely on services published by other teams
  - Easy changes to the running code -- seconds to propagate updated parameters to the underlying container infrastructure
- **Lazy allocation of resources**
  - Containers are running only when needed (i.e. when someone declared an intent to consume the corresponding service)
- **Continuous state enforcement**
  - Desired state of all services is rendered as a system and continuously validated/enforced 
- **Flexible rule engine. *See examples of user-defined rules:***
  - *Production Instances* get deployed to *us-west*, *Staging Instances* get deployed to *us-west*
  - *Web* and *Mobile* teams always share the same *small* flavor of *Analytics* service in *Staging*, while 
    *Healthcare* team gets a dedicated *high-performance* instance of the same service
  - *Development* teams can never deploy to *Production*
  - *Personal* development instances of *MyApp* can only be running *from 7am to 11pm* and should be terminated overnight 
    for all developers
- **Insights & Contextual visibility**
  - Understand what services are running, why they were instantiated, visualize dependencies and impact of changes. No
    need to deal with thousands of individual containers 

## User Guide

### Installation
TODO: "go get", script, docker container? it need to have clients, graphviz, etc

### Configuring LDAP
Aptomi needs to be configured with user data source in order to retrieve their labels/properties. It's recommended to
start with LDAP, which is also required by Aptomi examples and smoke tests.
1. LDAP Server with sample users is provided in a docker container. It's very easy to build and run it, just follow the instructions in [ldap-docker](tools/ldap-docker)
2. It's also recommended to download and install [Apache Directory Studio](http://directory.apache.org/studio/) for browsing LDAP. Once installed, follow these [step-by-step instructions](http://directory.apache.org/apacheds/basic-ug/1.4.2-changing-admin-password.html) to connect

### Getting Started
Once Aptomi is installed, you can get started by running the following examples:

Example    | Description 
-----------|-------------
[Example 1](examples/01/) | Description of Example 1 
[Example 2](examples/02/) | Description of Example 2
[Example 3](examples/03/) | Description of Example 3

### How It Works
TODO: architecture diagrams
TODO: how policy enforcement works

### Learning Aptomi language
TODO: policy documentation 

## Dev Guide

### Building From Source
Bulding Aptomi from source and running integration tests is a very straightforward process. All you need is Go (ideally 1.9.1) and a couple of external packages:
* graphviz - so that Aptomi can generate diagrams via [GraphViz](http://www.graphviz.org/Download..php)
* telnet - for the script which runs smoke tests
* docker - to run provided LDAP server with sample user data

Check out Aptomi source code from the repo:
```
mkdir $GOPATH/src/github.com/Aptomi
cd $GOPATH/src/github.com/Aptomi
git clone git@github.com:Aptomi/aptomi.git
```

If you are on macOS, install brew, install [Docker For Mac](https://docs.docker.com/docker-for-mac/install/) and run: 
```
brew install graphviz telnet docker
```

Install Helm, Istio, Kubectl clients:
```
./tools/install-clients.sh
```

All Go dependencies are managed using [Glide](https://glide.sh/). The following command will fetch all dependencies defined in `glide.lock` and put them into "vendor" folder:
```
make vendor 
```

To build the binary:
```
make 
```

### Tests & Linters

Command    | Target          | LDAP Required
-----------|-----------------|--------------
```make test```    | Unit tests | No
```make alltest``` | Integration + Unit tests | Yes
```make smoke```   | Smoke tests + Integration + Unit tests | Yes

Command     | Target          | Description
------------|-----------------|--------------
```make fmt```  | Format code | Re-formats all code according to Go standards
```make lint``` | Examine code | Run linters to examine Go source code and reports suspicious constructs

### How to contribute
TODO: write something about it

List of areas where we could use help:
- Adding support for additional cloud providers (AWS ECS, GKE, Docker Datacenter, Mesos)
- ... ... ...
- See [Feature Backlog](https://github.com/Aptomi/aptomi/milestone/11)

### Provided scripts
* `./tools/demo-gke.sh` - to set up 2 k8s clusters on GKE for demo. supports `up`, `down`, or `status`

### How to set up demo environment on Google Cloud
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

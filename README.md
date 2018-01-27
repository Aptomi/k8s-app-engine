![Aptomi Logo](images/aptomi-logo-new.png)

[![Release](https://img.shields.io/github/release/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/releases/latest)
[![License](https://img.shields.io/github/license/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aptomi/aptomi)](https://goreportcard.com/report/github.com/Aptomi/aptomi)
[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi%20-%20tests)](https://ci.aptomi.io/job/aptomi%20-%20tests/)
[![Coverage Status](https://coveralls.io/repos/github/Aptomi/aptomi/badge.svg)](https://coveralls.io/github/Aptomi/aptomi)
[![Godoc](https://godoc.org/github.com/Aptomi/aptomi?status.svg)](https://godoc.org/github.com/Aptomi/aptomi)
[![GitHub last commit](https://img.shields.io/github/last-commit/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/commits/master)
[![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io)

[Aptomi](http://aptomi.io) simplifies roll-out, operation and control of container-based applications on k8s. It introduces a
service-centric abstraction that allows Dev and Ops to collaborate asynchronously.  It enables teams to create and operate services,
share them across the organization, and fully control their lifecycle. Changes and updates are executed with a goal of minimizing
disruptive impact on depending services.

It is particularly useful in environments with multiple teams, clouds and data centers, where intent-based management
plays an important role in running large application infrastructure. Aptomi’s current focus is Kubernetes, but it's
designed to work with any container runtime and container orchestration technologies.

![What is Aptomi](images/aptomi-what-is.png)

## Demo

### Asciinema
[![asciicast](https://asciinema.org/a/k8ZpQTazoSaDV24fiLbG7DfT9.png)](https://asciinema.org/a/k8ZpQTazoSaDV24fiLbG7DfT9?speed=2)

### Youtube (more detailed)
[![youtube](http://img.youtube.com/vi/HL4RwoBnuTc/0.jpg)](http://www.youtube.com/watch?v=HL4RwoBnuTc)

## Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Features & Benefits](#features--benefits)
- [Where Aptomi is located in the stack](#where-aptomi-is-located-in-the-stack)
- [Quickstart](#quickstart)
  - [Installation](#installation)
  - [Preparing k8s Clusters](#preparing-k8s-clusters)
  - [Running Examples](#running-examples)
- [Architecture & How It Works](#architecture--how-it-works)
  - [Components](#components)
  - [State Enforcement](#state-enforcement)
  - [Language](#language)
- [Dev Guide](#dev-guide)
  - [Building From Source](#building-from-source)
  - [Tests & Code Validation](#tests--code-validation)
  - [Web UI](#web-ui)
  - [How to contribute](#how-to-contribute)
  - [How to release](#how-to-release)
  - [Roadmap](#roadmap)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Features & Benefits
- **Easy way to deploy and manage applications**
  - See [example](examples/twitter-analytics/diagram.png) of a multi-service application for k8s
- **Designed to run on top of any container platform**
  - k8s, OpenShift (with support coming for AWS ECS, GKE, Docker Datacenter, Mesos)
- **Friendly for Dev and Ops**
  - Think applications and services, not infrastructure primitives
  - Keep using your existing application templates (Helm, Ksonnet, k8s YAMLs, etc)
  - Make real-time changes to the running code (change parameters, relocate the whole application w/ dependencies to another cluster, restrict access, etc)
  - Minimizes disruption impact of change on depending services
- **Continuous state enforcement**
  - Desired state of all services is rendered as a system and continuously validated/enforced 
  - Built-in service discovery ensures all dependencies always are up to date    
- **Lazy allocation of resources**
  - Containers are running only when needed (i.e. when someone declared an intent to consume the corresponding service)
- **Flexible rule engine. *See examples of user-defined rules:***
  - *Production Instances* get deployed to *us-west*, *Staging Instances* get deployed to *us-west*
  - *Web* and *Mobile* teams always share the same *small* flavor of *Analytics* service in *Staging*, while 
    *Healthcare* team gets a dedicated *high-performance* instance of the same service
  - *Development* teams can never deploy to *Production*
  - *Personal* development instances of *MyApp* can only be running *from 7am to 11pm* and should be terminated overnight 
    for all developers
- **Insights & Contextual visibility**
  - UI to understand what services are running, why they were instantiated, visualize dependencies and impact of changes. No
    need to deal with thousands of individual containers
    ![Aptomi UI](images/aptomi-ui.png) 
        
## Where Aptomi is located in the stack
Aptomi sits in between CI/CD and container orchestration. Being in deployment path for applications, it can apply
higher-level policy rules (see examples above) and configure the underlying infrastructure components accordingly. 

![Aptomi Stack](images/aptomi-stack.png) 

## Quickstart

### Installation
There are several ways to install Aptomi. You may pick one that suits your needs the best:

Installation Mode     | Aptomi / UI | App Deployment | Description
----------------------|--------------------|----------------|-------------
[Concepts](docs/install_concepts.md) | *Yes (local)* | *No* | This is **NOT** a fully functional installation. Use this only if you want get familiar with Aptomi concepts, API and UI. Aptomi binaries will be installed on a local machine, it will be pre-uploaded with an example, while the actual engine for deploying apps to k8s will be disabled.
[Compact](docs/install_compact.md) | *Yes (local)* | *Yes* | Aptomi binaries will be installed on a local machine. Apps can be deployed via Aptomi to any local or remote k8s (minikube, docker for mac, GKE, etc)
[Kubernetes](docs/install_kubernetes.md) | *Yes (in k8s)* | *Yes* | Aptomi itself will be deployed on k8s in a container. Apps can be deployed via Aptomi to any local or remote k8s (minikube, docker for mac, GKE, etc)

### Preparing k8s Clusters

TODO: ...

You need to have access to k8s cluster in order to deploy services from the provided examples. Two k8s clusters will enable you to
take full advantage of Aptomi policy engine and use cluster-based rules.
1. If you don't have k8s clusters set up, follow [these instructions](examples/README.md) and run the provided script to create them in Google Cloud.
    ```
    ./tools/demo-gke.sh up
    ```

### Running Examples

TODO: ...

Once Aptomi is up and running and k8s clusters are set up, you can get started by running the following examples:

Example    | Description  | Diagram
-----------|--------------|--------------
[twitter-analytics](examples/twitter-analytics) | Twitter Analytics Application, multiple components, 2 k8s clusters | [Diagram](examples/twitter-analytics/diagram.png)

More examples are coming.

## Architecture & How It Works

### Components
![Aptomi Components](images/aptomi-components.png) 

### State Enforcement
![Aptomi Enforcement](images/aptomi-enforcement.png)

### Language
![Aptomi Language](images/aptomi-language.png)

See [language documentation](docs/language.md)

## Dev Guide

### Building From Source
In order to build Aptomi from source you will need Go (the latest 1.9.x) and a couple of external dependencies:
* glide - all Go dependencies for Aptomi are managed via [Glide](https://glide.sh/)
* docker - to run Aptomi in container, as well as to run sample LDAP server with user data
* kubernetes-cli and kubernetes-helm for using Kubernetes with Helm
* npm - to build UI, as well as automatically generate table of contents in README.md 
* telnet, jq - for the script which runs smoke tests

If you are on macOS, install [Homebrew](https://brew.sh/) and [Docker For Mac](https://docs.docker.com/docker-for-mac/install/), then run: 
```
brew install go glide docker kubernetes-cli kubernetes-helm npm telnet jq
```

Check out Aptomi source code from the repo:
```
mkdir -p $GOPATH/src/github.com/Aptomi
cd $GOPATH/src/github.com/Aptomi
git clone https://github.com/Aptomi/aptomi.git
```

In order to build Aptomi, you must first tell Glide to fetch all of its dependencies. It will read the list of
dependencies defined in `glide.lock` and fetch them into a local "vendor" folder. After that, you must run Go to
build and install the binaries. There are convenient Makefile targets for both, run them:
```
make vendor 
make install
```

### Tests & Code Validation

Command    | Action          | LDAP Required
-----------|-----------------|--------------
```make test```    | Unit tests | No
```make alltest``` | Integration + Unit tests | Yes
```make smoke```   | Smoke tests + Integration + Unit tests | Yes
```make profile-engine```   | Profile engine for CPU usage | No
```make coverage```   | Calculate code coverage by unit tests | No
```make coverage-full```   | Calculate code coverage by unit & integration tests | Yes

Command     | Action          | Description
------------|-----------------|--------------
```make fmt```  | Re-format code | Re-formats all code according to Go standards
```make lint``` | Examine code | Run linters to examine Go source code and reports suspicious constructs

### Web UI
Source code is available in [webui](webui)

Make sure you have latest `node` and `npm`. We have tested with node v8.9.1 and npm 5.5.1 and it's
known to work with these.

Command     | Action
------------|----------
```npm install```  | Install dependencies
```npm run dev``` | Serve with hot reload at localhost:8080
```npm run build``` | Build for production with minification
```npm run build --report``` | Build for production and view the bundle analyzer report
```npm run unit``` | Run unit tests: *coming soon*
```npm run e2e``` | Run e2e tests: *coming soon*
```npm run test``` | Run all tests: *coming soon*

### How to contribute
Report a bug. Send us a pull request.

List of areas where we could use help:
- Feedback from Dev & Ops teams on service & rule definitions
- Adding support for additional cloud providers (AWS ECS, GKE, Docker Datacenter, Mesos)
- Also, see [Feature Backlog](https://github.com/Aptomi/aptomi/milestone/11)

### How to release
Use `git tag` and `make release` for creating new release.

1. Create annotated git tag and push it to github repo. Use commit message like `Aptomi v0.1.2`.

```
git tag -a v0.1.2
git push origin v0.1.2
```

1. Create GitHub API token with the `repo` scope selected to upload artifacts to GitHub release page. You can create
one [here](https://github.com/settings/tokens/new). This token should be added to the environment variables as `GITHUB_TOKEN`.

1. Run `make release`. It'll create everything needed and upload all artifacts to github.

1. Go to https://github.com/Aptomi/aptomi/releases/tag/v0.1.2 and fix changelog / description if needed.

### Roadmap
We will soon publish the list of items for Q4 2017 and Q1 2018. In the meantime,
[Feature Backlog](https://github.com/Aptomi/aptomi/milestone/11) is a good place to look at the roadmap items
which are being considered.

If you have any questions, please contact us on [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io).

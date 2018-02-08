![Aptomi Logo](images/aptomi-logo-new.png)

[![Release](https://img.shields.io/github/release/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/releases/latest)
[![License](https://img.shields.io/github/license/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aptomi/aptomi)](https://goreportcard.com/report/github.com/Aptomi/aptomi)
[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi%20-%20tests)](https://ci.aptomi.io/job/aptomi%20-%20tests/)
[![Coverage Status](https://coveralls.io/repos/github/Aptomi/aptomi/badge.svg)](https://coveralls.io/github/Aptomi/aptomi)
[![Godoc](https://godoc.org/github.com/Aptomi/aptomi?status.svg)](https://godoc.org/github.com/Aptomi/aptomi)
[![GitHub last commit](https://img.shields.io/github/last-commit/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/commits/master)
[![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io)

[Aptomi](http://aptomi.io) simplifies roll-out, operation and control of container-based applications on **Kubernetes**. It introduces a
**service-centric abstraction** that allows Dev and Ops to collaborate asynchronously. It enables teams to create and operate services,
share them across the organization, fully control their lifecycle while enforcing Ops/Governance policies. Changes and updates are executed
with a goal of minimizing disruptive impact on depending services.

It is particularly useful in environments with multiple teams, clouds and data centers, where intent-based management
plays an important role in running large application infrastructure. Aptomi’s current focus is Kubernetes, but it's
designed to work with any container runtime and container orchestration technologies.

![What is Aptomi](images/aptomi-what-is.png)

## Demo

### Short demo (5 minutes, Asciinema)
[![asciicast](images/aptomi-asciinema.png)](https://asciinema.org/a/k8ZpQTazoSaDV24fiLbG7DfT9?speed=2)

### Detailed demo (13 minutes, Youtube)
[![youtube](images/aptomi-demo-youtube.png)](http://www.youtube.com/watch?v=HL4RwoBnuTc)

## Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Features & Benefits](#features--benefits)
- [Quickstart](#quickstart)
  - [Step #1: Installation](#step-1-installation)
  - [Step #2: Setting up k8s Cluster](#step-2-setting-up-k8s-cluster)
  - [Step #3: Running Examples](#step-3-running-examples)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [State Enforcement](#state-enforcement)
  - [Language](#language)
- [How to contribute](#how-to-contribute)
- [Roadmap](#roadmap)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Features & Benefits
- **Deploy and manage container-based applications with ease**
  - Dev and Ops think applications and services, not infrastructure primitives and thousands of containers
  - Built-in service discovery ensures all dependencies always are up to date
  - No need to change existing application templates (Helm, Ksonnet, k8s YAMLs, etc)
  - Run on k8s, OpenShift (support for AWS ECS, GKE, Docker Datacenter, Mesos is pluggable)
- **Lazy allocation of resources**
  - Containers are only running when the corresponding service has consumers
- **Continuous state enforcement**
  - Desired state of all services is rendered as a DAG system and continuously validated/enforced
  - Changes/rules can be enforced at any time (change service parameters, relocate the whole application w/ dependencies to another cluster, restrict access, etc)
  - Disruption impact of change on depending services is minimized
- **Flexible rule engine**. Examples:
  - *Production Instances* get deployed to *us-west*, *Staging Instances* get deployed to *us-west*
  - *Web* and *Mobile* teams always share the same *small* flavor of *Analytics* service in *Staging*, while 
    *Healthcare* team gets a dedicated *high-performance* instance of the same service
  - *Development* teams can never deploy to *Production*
  - *Personal* development instances of *MyApp* can only be running *from 7am to 11pm* and should be terminated overnight 
    for all developers
- **Insights & Contextual visibility**
  - UI to explore services instances, see why services were instantiated, visualize dependencies and impact of changes

## Quickstart

### Step #1: Installation
There are several ways to install Aptomi. The simplest one is **Compact**, but you may pick one that suits your needs:

Installation Mode     | Complexity | Aptomi        | App Deployment | Description
----------------------|------------|---------------|----------------|------------
[Compact](docs/install_compact.md) | Easy | Local machine | *Yes* | Aptomi will be installed on a local machine (binaries or in a single Docker container)
[Kubernetes](docs/install_kubernetes.md) | Medium | Container on k8s  | *Yes* | Aptomi will be deployed on k8s via Helm chart

You can also install it in a stripped-down mode, mostly to explore concepts and look at API/UI. It will use a fake executor and thus will **NOT** be able to perform any app deployments to k8s:

Installation Mode     | Aptomi / UI | App Deployment | Description
----------------------|--------------------|----------------|-------------
[Concepts](docs/install_concepts.md) | Local machine | *No* | Use this only if you want get familiar with Aptomi concepts, API and UI. k8s is not required

### Step #2: Setting up k8s Cluster

You need to have access a k8s cluster to deploy apps from the provided examples:

Kubernetes Cluster | When to use     | How to run
------------|-----------------|-----------
Your own    | If you already have k8s cluster set up | [Configure Aptomi to use an existing k8s cluster](docs/k8s_own.md)
Google Kubernetes Engine | Useful if you have a new Google account and free credits | [Configure Aptomi to use GKE](docs/k8s_gke.md)
k8s / Minikube | Single-node, local machine with 16GB+ RAM | [Configure Aptomi to use Minikube](docs/k8s_minikube.md)
k8s / Docker For Mac | Single-node, local machine with 16GB+ RAM | [Configure Aptomi to use Docker For Mac](docs/k8s_docker_for_mac.md)

Having a powerful k8s cluster with good internet connection will definitely provide *better experience* compared to a single-node k8s local cluster. GKE would be one of the best options.

### Step #3: Running Examples

Once Aptomi server is up and k8s cluster is ready, you can get started by running the following examples:

Example    | Description
-----------|------------
[twitter-analytics](examples/twitter-analytics) | Twitter Analytics Application, multiple services, multi-cloud

## How It Works

### Architecture
![Aptomi Components](images/aptomi-components.png) 

### State Enforcement
![Aptomi Enforcement](images/aptomi-enforcement.png)

### Language
![Aptomi Language](images/aptomi-language.png)

See [language documentation](docs/language.md)

## How to contribute
The very least you can do is to [report a bug](https://github.com/Aptomi/aptomi/issues)!

If you want to make a pull request for a bug fix or contribute a feature, see our [Development Guide](docs/dev_guide.md) for how to develop, run and test your code.

In general, we are always looking for feedback on:
- Aptomi object model - definitions of services, contracts, rules, clusters
- Pluggability - support for additional label sources (in addition to LDAP), app engines (in addition to Helm), cloud providers

Contact us on [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io).

## Roadmap
[Feature Backlog](https://github.com/Aptomi/aptomi/milestone/11), as well as weekly project milestones, are good places to look at the roadmap items.

If you have any questions, please contact us on [![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io).

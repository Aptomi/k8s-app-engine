![Aptomi Logo](images/aptomi-logo-new.png)

[![Release](https://img.shields.io/github/release/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/releases/latest)
[![License](https://img.shields.io/github/license/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/LICENSE.md)
[![Go Report Card](https://goreportcard.com/badge/github.com/Aptomi/aptomi)](https://goreportcard.com/report/github.com/Aptomi/aptomi)
[![Build Status](https://ci.aptomi.io/buildStatus/icon?job=aptomi%20-%20tests)](https://ci.aptomi.io/job/aptomi%20-%20tests/)
[![Coverage Status](https://coveralls.io/repos/github/Aptomi/aptomi/badge.svg)](https://coveralls.io/github/Aptomi/aptomi)
[![Godoc](https://godoc.org/github.com/Aptomi/aptomi?status.svg)](https://godoc.org/github.com/Aptomi/aptomi)
[![GitHub last commit](https://img.shields.io/github/last-commit/Aptomi/aptomi.svg)](https://github.com/Aptomi/aptomi/commits/master)
[![Slack Status](https://img.shields.io/badge/slack-join_channel-ff69b4.svg)](http://slack.aptomi.io)

[Aptomi](http://aptomi.io) is a platform for development teams that simplifies roll-out and operation of container-based applications on **Kubernetes**. It introduces a service-centric abstraction, which allows to compose applications from multiple components connected together. It supports components packaged using Helm, ksonnet, k8s YAMLs or any other Kubernetes-friendly way.

Aptomi's approach to **application delivery** becomes especially powerful in a **multi-team** setup, where components owned by different teams must be put together into a service. With ownership boundaries, Dev teams can specify **multi-cluster** and **multi-env** (e.g. dev, stage, prod) service behavior, as well as control lifecycle and updates of their respective services.

It also provides contextual **visibility** into teams and services, allowing to visualize dependencies and impact of changes. 

![What is Aptomi](images/aptomi-what-is.png)

## Demo

### Short demo (5 minutes, Asciinema)
[![asciicast](images/aptomi-asciinema.png)](https://asciinema.org/a/k8ZpQTazoSaDV24fiLbG7DfT9?speed=2)

### Detailed demo (13 minutes, Youtube)
[![youtube](images/aptomi-demo-youtube.png)](http://www.youtube.com/watch?v=HL4RwoBnuTc)

## Table of contents
<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->


- [Why Aptomi](#why-aptomi)
- [Quickstart](#quickstart)
  - [Step #1: Installation](#step-1-installation)
  - [Step #2: Setting up k8s Cluster](#step-2-setting-up-k8s-cluster)
  - [Step #3: Running Examples](#step-3-running-examples)
- [How It Works](#how-it-works)
  - [Architecture](#architecture)
  - [Language](#language)
- [How to contribute](#how-to-contribute)
- [Roadmap](#roadmap)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## Why Aptomi

* Why would I want to use Aptomi [as a Developer](docs/benefits.md#for-developers)
* Why would I want to use Aptomi [as an Operator](docs/benefits.md#for-operators)
* How does Aptomi fit into [CI/CD model with Spinnaker/Jenkins](docs/benefits.md#how-aptomi-fits-into-cicd-jenkins-spinnaker)
* Why would I want to use Aptomi [if I'm already implementing Kubernetes/OpenShift](docs/benefits.md#why-would-i-want-to-use-aptomi-if-im-already-implementing-kubernetesopenshift)

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

See [artchitecture documentation](docs/architecture.md)

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

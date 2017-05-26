# Aptomi

## Dependencies

All Go dependencies are managed using [Glide](https://glide.sh/).
* Install dependencies (vendor dir) with versions from ```glide.lock``` file:
  ```glide install```
* Update dependencies versions (in ```glide.lock``` file): ```glide update```

Currently there is only one external dependency -
[GraphViz](http://www.graphviz.org/Download..php) and it could be installed on
macOS using ```brew install graphviz```.

# How to test


To run tests on a project:

```shell
make test
```

Or to run tests on concrete package:

```shell
go test -v ./pkg/slinga
```

## How to build and run

Directory 'testdata' is excluded from processing by 'go' tool:
https://golang.org/cmd/go/#hdr-Description_of_package_lists

To build a binary (named ```aptomi```):

```shell
make build
```

Must define environment variables:

* ```APTOMI_DB``` = <path to the directory where usage/allocation state will be serialized"
* ```APTOMI_POLICY``` = <path to the directory where policy will be taken from"

To run ```aptomi``` without compilation:

```shell
go run main.go show config
go run main.go policy apply
go run main.go policy apply --noop
go run main.go show graph
```

## Tools

* ```make fmt``` runs fmt that will ensure code style and print changed files
* ```make vet``` runs vet that examines Go source code and reports suspicious
  constructs, such as Printf calls whose arguments do not align with the format
  string. Vet uses heuristics that do not guarantee all reports are genuine
  problems, but it can find errors not caught by the compilers
* ```make lint``` runs linter for Go source code

## Issues observed:
1. With the current definition of services and contexts, there is NO way to provide different
   implementation of the same "service interface"
      - e.g. SQL service -> MySQL or MariaDB
2. Duplication of data in context definitions


## Questions:
1. How service developer workflow would change with aptomi? How to roll out a change to a service?
   Code change -> container rebuild -> push a change to production
   We need to make emphasis on "as code" (!)
2. Service, context - use IDs instead of names?


Another event to update component in addition to existing 4 events?



Add property "aliases" to top of the context, it'll mean for which service requests this context will be found.

Example context "context.test.kafka.yaml":

name: test
service: kafka
aliases:
  - mq
  - MessageQueue


if we implement aliases in services, then what happens if we request "db"?
- we need to match a context first
- then the context points us to a specific service


Processing flow:

1. user defines dependency on service "kafka", labels+=user.labels
2. search contexts with service or alias "kafka"
3. process criteria for each found context and select first matching, labels+=context.service.labels
4. find corresponding allocation
5. for each component:
5.1. labels+=component.labels
5.2. labels+=context.labels
5.3. labels+=allocation.labels

## What needs to be done:

* [Done - RA] add labels to dependencies
* [Done - RA] recursively process folders when loading policy
* [Done - RA] dry run (via 'trace' attribute)
* [Done - RA/SL] service discovery
* [Done - SL] use temp files to path params to Helm charts instead of CLI
* [Done - SL] save Helm output to temp file and print its name for future debug
* [Done - SL] wrap spark job for twitter stats with Helm chart
* [SL] update demo policy to have demo-like topology
* [Demo - SL] make demo policy works with real charts & tools
* [SL] build & push fresh aptomi images & helm charts
* [SL] add check for !compromised user
* [SL or RA] Make sure that if something failed during apply, new state will not be saved, so, we can just re-run apply
* [SL] add Last updated x seconds ago to tweeviz and custom info field (User name + stage / prod)
* [SL] aptomi show endpoints (print services endpoints)
* [SL] code change demo - two versions of tweepub with different visaulizations available as 2 docker tags
* [SL] impl multiple k8s support - add some file with list of k8s clusters (and tiller addresses) and specify cluster name in metadata for helm chart
* [Done - RA] more compact visualization
* [Done - RA] "no changes" shit is incorrect, doesn't take component updates into account
* [RA] criteria -> accept and reject

* We can consider adding "tests for DevOps" (smart tracing)

## Demo scenario:

1. Show policy
   - Data Analytics Pipeline
   - Twitter Real-Time Stats

2. Allocate instances for users
   - Shared DPP and separate "prod" TRSs for two users

3. U1 makes a code change and deploys to staging
   - allocate "staging" TRS with a code change (background color + h1)

4. U1 propagates a change to production
   - "prod" TRS updated

5. U2 gets marked as "untrusted"
   - loses access to his "prod"

## Bad stuff:

1. Can arrive to the same instance with different sets of labels. Unclear what to do in this case

2. Two users -> same service instance -> relies on different component instances. E.g. two users, single twitter-stats, two kafkas. Invalid case?


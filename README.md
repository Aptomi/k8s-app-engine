# Aptomi

## Dependencies

All Go dependencies are managed using [Glide](https://glide.sh/).
* Install dependencies (vendor dir) with versions from ```glide.lock``` file:
  ```glide install```
* Update dependencies versions (in ```glide.lock``` file): ```glide update```

Currently there is only one external dependency -
[GraphViz](http://www.graphviz.org/Download..php) and it could be installed on
macOS using ```brew install graphviz```.

# If you see issue with gcloud token

```shell
kubectl --context cluster-us-east get pods
kubectl --context cluster-us-west get pods
helm --kube-context cluster-us-west list --all
helm --kube-context cluster-us-east list --all
```

# How to test

To run smoke tests (it will run unit tests and apply policy in noop mode):

```shell
make smoke
```

To run just all unit tests:

```shell
make test
```

To run tests on specific package:

```shell
go test -v ./pkg/slinga
```

## How to install & configure LDAP Server
1. Download & configure LDAP server, including creation of all the test users
```shell
cd tools/ldap-server
ldap_server_init.sh
```
2. Download and install Apache Directory Studio. Check that connection can be established
  - http://directory.apache.org/studio/
  - http://directory.apache.org/apacheds/basic-ug/1.4.2-changing-admin-password.html

## How to build and run

Directory 'testdata' is excluded from processing by 'go' tool:
https://golang.org/cmd/go/#hdr-Description_of_package_lists

To build a binary (named ```aptomi```):

```shell
make build
```

Must define environment variables:

* ```APTOMI_DB``` = <path to aptomi working directory, where aptomi will store all its files (policy, logs, etc)>

To run ```aptomi``` without compilation:

```shell
go run main.go show config
go run main.go policy apply
go run main.go policy apply --noop
go run main.go policy apply --noop --show
go run main.go policy apply --trace
```

## Tools

* ```make validate``` runs code re-format and error/warning checks on code (everything listed below). In you want to run individual things:
  * ```make fmt``` runs fmt that will ensure code style and print changed files
  * ```make vet``` runs vet that examines Go source code and reports suspicious
    constructs, such as Printf calls whose arguments do not align with the format
    string. Vet uses heuristics that do not guarantee all reports are genuine
    problems, but it can find errors not caught by the compilers
  * ```make lint``` runs linter for Go source code

## How to set up demo environment on Google Cloud

1. ```brew install kubernetes-cli kubernetes-helm```
1. ```curl https://sdk.cloud.google.com | bash```
1. ```gcloud auth login```
1. Create new project in https://console.cloud.google.com/
1. ```gcloud config set project <YOUR_PROJECT_ID>```
1. https://console.cloud.google.com/ -> API Manager -> Enable API
  1. Google Container Engine API
  1. Google Compute Engine API
1. ```./tools/gke-demo.sh up```
1. ```./tools/gke-demo.sh status```
1. Run demo (see README_DEMO.md for instructions)
1. ```./tools/gke-demo.sh down``` - don't forget to destroy your clusters, so you don't continue to get billed for them


## Fundamental design questions (for production code, not PoC):
1. Policy & Language
    1. Can arrive to the same instance with different sets of labels. Unclear what to do in this case
    1. Two users -> use same service instance -> it relies on different component instances. E.g. two users, single twitter-stats, then it branches into two kafkas due to labels. If we refer to ".instance" of kafka from twitter-stats, it won't work (same problem as above)
    1. Inheritance of contexts to avoid data duplication
    1. How to implement service aliases (mysql vs. mariadb, etc). Do we match contexts first or services? With the current definition of services and contexts, there is NO way to provide different implementation of the same "service interface (e.g. SQL service -> MySQL or MariaDB)
    1. Add "aptomi test" and special language for Ops to write and run basic tests after the policy is defined (talk to Roman about this)

1. Access control
    1. Introduce a notion of user roles into Aptomi
    1. Introduce a notion of "policy namespace" into Aptomi
      - allow ops to create policy namespaces and specify access rules (who can make what changes in which namespace)
      - e.g. Dev can have full access to their own playgrounds (create services, define instantiation rules, change dependencies) to test and deploy to their clusters
      - other pieces will have access for Ops only

1. Integrations
    1. Figure out a final solution for service discovery
    1. Integration with CI/CD, at least demo "Code change -> container rebuild -> push a change to production" without having explicit tags in Aptomi policy

1. Code/Implementation
  1. Error handling
  1. Finish breaking it down into packages
  1. Every object should have a kind (type) and ID. Use ID (unique) instead of names (non-unique)
  1. Aptomi DB
    1. Move away from file-based storage (db.yaml)
    1. Schema changes between versions. If we change a format of parameter (e.g. 'criteria'), how to handle it correctly?
  1. Better handling of Aptomi revisions. Compare policy, instances, users to detect difference
  1. Right now there is no API (only partial API for UI). For processing, we run CLI in the loop. CLI does actions directly, not via API
  1. Add unit tests for corner cases. E.g. when user gets deleted and disappears, circular service dependency, circular component dependency
  1. When something fails in the middle of applying policy. E.g. some components got deployed and some didn't. How to handle it?
  1. LDAP - can we subscribe to events? I.e. so we can get notified when user labels change

1. Testing
  1. Store history of aptomi revisions and continuously regression test against old stored runs. To emulate production use cases and Aptomi updates

## Resolved issues
  1. Handle "partial matchings" correctly. E.g. access to kafka is allowed, but kafka depends on zookeeper and access to zookeeper is not allowed. The whole thing should be "rolled back"
  1. Store all revisions. Every time we apply policy version would increase
  1. If calculation logic changes between runs, how can be force these changes to be applied? It thinks that there are no changes. --newrevision


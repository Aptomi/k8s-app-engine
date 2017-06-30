# Aptomi

## Dependencies

All Go dependencies are managed using [Glide](https://glide.sh/).
* Install dependencies (vendor dir) with versions from `glide.lock` file:
  `glide install`
* Update dependencies versions (in `glide.lock` file): `glide update`

External dependencies can be installed with:
* [GraphViz](http://www.graphviz.org/Download..php), on Mac OS via `brew install graphviz`
* Helm, Istio, Kubectl, via `./tools/install-clients.sh`

# How to test

Unit test:
```shell
make test
```

Integration + Unit tests (LDAP must be up and running):
```shell
make alltest
```

Smoke tests (it will 'alltest' and apply policy in noop mode):

```shell
make smoke
```

## Tools

* ```make validate``` runs code re-format and error/warning checks on code (everything listed below). In you want to run individual stages:
  * ```make fmt``` runs fmt that will re-format code and output the list of changed files
  * ```make vet``` runs vet that examines Go source code and reports suspicious
    constructs, such as Printf calls whose arguments do not align with the format
    string. Vet uses heuristics that do not guarantee all reports are genuine
    problems, but it can find errors not caught by the compilers
  * ```make lint``` runs linter for Go source code

## How to build

To build `aptomi` binary:

```shell
make build
```

To install `aptomi` binary:

```shell
make install
```

## How to run Aptomi

Must define environment variables:
* ```APTOMI_DB``` = <path to aptomi working directory, where aptomi will store all its files (policy, logs, etc)>

Running compiled version:
* `aptomi policy reset --force` to delete all files from Aptomi database and start from scratch with an empty state
* `aptomi policy apply --noop` to run policy resolution, print what would be done, but don't do actual deployment
* `aptomi policy apply --emulate` to run policy resoluton, mark objects as deployed, but don't actually deploy (emulate deployment). Useful if you want to test Aptomi engine, but don't want to wait while everything will be deployed to k8
* `aptomi policy apply --full` to re-create missing instances and update existing instances. Useful if deployed instances somehow went missing from the underlying cloud (e.g. got manually deleted)
* `aptomi policy apply --newrevision` generate new revision of the configuration, even if no changes were detected. Can be useful if you changed the logic of how parameters are calculated and want to re-generate aptomi DB

Running without compilation, e.g.:
```shell
go run main.go policy apply --noop
```


## Provided scripts

* `./tools/demo-gke.sh` - to set up 2 k8s clusters on GKE for demo. supports `up`, `down`, or `status`
* `./tools/demo-init.sh` - starts the demo (except LDAP)
  * `./tools/demo-local-policy-init.sh` - init local database with demo policy
  * `./tools/demo-push.sh` - pushes demo policy to https://github.com/Frostman/aptomi-demo/
  * `./tools/demo-watch-apply.sh` - watches remote github repo. Once new commit is detected, it updates the local copy and runs Aptomi
* `tools/dev-watch-server.sh` - starts Aptomi UI in Dev mode. If .go files get changed, it will recompile and re-launch the server

## How to install & configure local LDAP Server
1. Download & configure LDAP server, including creation of all the test users for the demo
```shell
./tools/ldap-server/ldap_server_init.sh
```
2. Download and install Apache Directory Studio. Check that connection can be established
  - http://directory.apache.org/studio/
  - See http://directory.apache.org/apacheds/basic-ug/1.4.2-changing-admin-password.html on how to connect to LDAP

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


## Outstanding Issues
1. **Policy & Language**
    1. Can arrive to the same instance with different sets of labels. Unclear what to do in this case
    1. Two users -> use same service instance -> it relies on different component instances. E.g. two users, single twitter-stats, then it branches into two kafkas due to labels. If we refer to ".instance" of kafka from twitter-stats, it won't work (same problem as above)
    1. Inheritance of contexts to avoid data duplication
    1. How to implement service aliases (mysql vs. mariadb, etc). Do we match contexts first or services? With the current definition of services and contexts, there is NO way to provide different implementation of the same "service interface (e.g. SQL service -> MySQL or MariaDB)
    1. Add "aptomi test" and special language for Ops to write and run basic tests after the policy is defined (talk to Roman about this)

1. **Access control**
    1. Introduce a notion of user roles into Aptomi
    1. Introduce a notion of "policy namespace" into Aptomi
        1. allow ops to create policy namespaces and specify access rules (who can make what changes in which namespace)
        1. e.g. Dev can have full access to their own playgrounds (create services, define instantiation rules, change dependencies) to test and deploy to their clusters. Other pieces will have access for Ops only

1. **Integrations**
    1. Figure out a final solution for service discovery
    1. Integration with CI/CD, at least demo "Code change -> container rebuild -> push a change to production" without having explicit tags in Aptomi policy

1. **Code/Implementation**
    1. Error handling
    1. Finish breaking it down into packages
    1. Every object should have a kind (type) and ID. Use ID (unique) instead of names (non-unique)
    1. System of plugins for the engine, so you can easily insert additional integrations (e.g. istio)
    1. Aptomi DB
        1. Move away from file-based storage (db.yaml)
        1. Schema changes between versions. If we change a format of parameter (e.g. 'criteria'), how to handle it correctly?
    1. Better handling of Aptomi revisions. Compare policy, instances, users to detect difference
    1. Right now there is no API (only partial API for UI). For processing, we run CLI in the loop. CLI does actions directly, not via API
    1. Add unit tests for corner cases. E.g. when user gets deleted and disappears, circular service dependency, circular component dependency
    1. When something fails in the middle of applying policy. E.g. some components got deployed and some didn't. How to handle it?
    1. Re-think CLI flags and come up with better names
    1. LDAP - can we subscribe to events? I.e. so we can get notified when user labels change
    1. Clean up Istio tech debt

1. **Testing**
    1. Store history of aptomi revisions and continuously regression test against old stored runs. To emulate production use cases and Aptomi updates

## Resolved Issues
1. Handle "partial matchings" correctly. E.g. access to kafka is allowed, but kafka depends on zookeeper and access to zookeeper is not allowed. The whole thing should be "rolled back"
1. Store all revisions. Every time we apply policy version would increase
1. If calculation logic changes between runs, how can be force these changes to be applied? It thinks that there are no changes. --newrevision


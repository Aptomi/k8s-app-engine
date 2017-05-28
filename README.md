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

## Fundamental design questions (for production code, not PoC):
1. Model & Engine
  1. Can arrive to the same instance with different sets of labels. Unclear what to do in this case
  1. Two users -> same service instance -> relies on different component instances. E.g. two users, single twitter-stats, two kafkas. Invalid case?
  1. Inheritance of contexts to avoid data duplication
  1. How to implement service aliases (mysql vs. mariadb, etc). Do we match contexts first or services? With the current definition of services and contexts, there is NO way to provide different implementation of the same "service interface (e.g. SQL service -> MySQL or MariaDB)
  1. Service, context - shall we use IDs (unique) instead of names (non-unique)?
  1. When something failed in the middle of applying policy. How to handle it?
  1. Handle "partial matchings" correctly. E.g. access to kafka is allowed, but kafka depends on zookeeper and access to zookeeper is not allowed
  1. Detect circular dependencies (global cycle between services, not only cycle within one service between its components)

1. CI/CD
  1. How service developer workflow would change with aptomi? How to roll out a change to a service? Code change -> container rebuild -> push a change to production

1. Error handling

1. DB
  1. How to handle schema change. If we change a format of parameter (e.g. 'criteria'), how do we handle it correctly?. Prev can be one version. Next can be another version
  1. Do we need to store all versions? Every time we apply policy version would increase

1. Structure of Go project
  1. Right now everything is in one package. Not very good
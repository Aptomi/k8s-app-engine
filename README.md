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

Issues observed:
1. With the current definition of services and contexts, there is NO way to provide different
   implementation of the same "service interface"
      - e.g. SQL service -> MySQL or MariaDB
2. Duplication of data in context definitions


Questions:
1. How service developer workflow would change with aptomi? How to roll out a change to a service?
   Code change -> container rebuild -> push a change to production
   We need to make emphasis on "as code" (!)
2. Service, context - use IDs instead of names?

To install dependencies:
# go get gopkg.in/yaml.v2
# go get github.com/stretchr/testify
# go get github.com/Knetic/govaluate
# go get github.com/awalterschulze/gographviz
# go get github.com/golang/glog
# there will be also dependency on cobra for CLI

Also:
http://www.graphviz.org/Download..php
(mac os you can just # brew install graphviz)

To build a package:
# go build aptomi/slinga

To run tests on a package:
# go test aptomi/slinga

To run all tests:
# go test -v ./...

To build a CLI named "aptomi" (instead of aptomi-cli):
# go build -o aptomi   (relative path, if in aptomi-cli directory)
# go build -o aptomi aptomi/aptomi-cli   (absolute path)

Directory 'testdata' is excluded from processing by 'go' tool:
https://golang.org/cmd/go/#hdr-Description_of_package_lists

Must define environment variables:
APTOMI_DB = <path to the directory where usage/allocation state will be serialized"
APTOMI_POLICY = <path to the directory where policy will be taken from"

To run Aptomi CLI without compilation:
# cd aptomi-cli
# go run main.go show config
# go run main.go policy apply
# go run main.go policy apply --noop
# go run main.go show graph


Issues observed:
1) With the current definition of services and contexts, there is NO way to provide different
   implementation of the same "service interface"
      - e.g. SQL service -> MySQL or MariaDB
2) Duplication of data in context definitions


Questions:
1) How service developer workflow would change with aptomi? How to roll out a change to a service?
   Code change -> container rebuild -> push a change to production
   We need to make emphasis on "as code" (!)
2) Service, context - use IDs instead of names?
To install dependencies:
# go get gopkg.in/yaml.v2
# go get github.com/stretchr/testify
# go get github.com/Knetic/govaluate
# go get github.com/awalterschulze/gographviz
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

Directory 'testdata' is excluded from processing by 'go' tool:
https://golang.org/cmd/go/#hdr-Description_of_package_lists


Issues observed:

1) With the current definition of services and contexts, there is NO way to provide different
   implementation of the same "service interface"
      - e.g. SQL service -> MySQL or MariaDB

2) Duplication of data in context definitions
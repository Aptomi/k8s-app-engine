To install dependencies:
# go get gopkg.in/yaml.v2
# go get github.com/stretchr/testify
# go get github.com/Knetic/govaluate
# go get github.com/awalterschulze/gographviz

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
.PHONY: default
default: clean build test

.PHONY: vendor
vendor:
	glide install --strip-vendor

.PHONY: test
test:
	go test -v ./cmd/...
	go test -v ./pkg/...

.PHONY: build
build:
	go build -o aptomi

.PHONY: fmt
fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

.PHONY: vet
vet:
	go vet -v $$(go list ./... | grep -v /vendor/)

.PHONY: lint
lint:
	$$(go env GOPATH)/bin/golint $$(go list ./... | grep -v /vendor/)

.PHONY: validate
validate: fmt vet

.PHONY: clean
clean:
	-rm -f aptomi

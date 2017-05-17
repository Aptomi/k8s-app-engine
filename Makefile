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

.PHONY: clean
clean:
	-rm -f aptomi

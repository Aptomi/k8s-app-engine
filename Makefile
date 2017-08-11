.PHONY: default
default: clean build test

.PHONY: vendor
vendor:
	glide install --strip-vendor

.PHONY: profile
profile:
	@echo "Profiling CPU for 15 seconds"
	go test -bench . -benchtime 15s ./pkg/slinga/engine -cpuprofile cpu.out
	go tool pprof -web engine.test cpu.out

.PHONY: test
test:
	go test -short -v ./cmd/...
	go test -short -v ./pkg/...
	@echo "\nAll unit tests passed"

.PHONY: alltest
alltest:
	go test -v ./cmd/...
	go test -v ./pkg/...
	@echo "\nAll unit & integration tests passed"

.PHONY: test-loop
test-loop:
	while go test -v ./pkg/slinga; do :; done

.PHONY: clean-run-noop
clean-run-noop:
	$(eval TMP := $(shell mktemp -d))
	APTOMI_DB=$(TMP) && tools/demo-local-policy-init.sh

.PHONY: smoke
smoke: alltest build install clean-run-noop

.PHONY: emulate
emulate:
	tools/demo-local-policy-init.sh
	aptomi policy apply --emulate
	aptomi policy apply --emulate --newrevision
	tools/dev-enable-all.sh
	aptomi policy apply --emulate
	aptomi policy apply --emulate --newrevision

.PHONY: build
build:
	CGO_ENABLED=0 go build -i -o aptomi

.PHONY: install
install:
	go install

.PHONY: fmt
fmt:
	go fmt $$(go list ./... | grep -v /vendor/)

.PHONY: vet
vet:
	go tool vet -all -shadow main.go
	go tool vet -all -shadow ./cmd ./pkg

.PHONY: lint
lint:
	$$(go env GOPATH)/bin/golint $$(go list ./... | grep -v /vendor/) | grep -v 'should not use dot imports'

.PHONY: validate
validate: fmt vet lint
	@echo "\nAll validations passed"

.PHONY: clean
clean:
	-rm -f aptomi
	go clean -r -i

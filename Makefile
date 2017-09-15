GIT_VERSION=$(shell git describe --tags --long --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GOENV=CGO_ENABLED=0
GOFLAGS=-ldflags "-X github.com/Aptomi/aptomi/pkg/slinga/version.gitVersion=${GIT_VERSION} -X github.com/Aptomi/aptomi/pkg/slinga/version.gitCommit=${GIT_COMMIT} -X github.com/Aptomi/aptomi/pkg/slinga/version.buildDate=${BUILD_DATE}"
GO=${GOENV} go

.PHONY: default
default: clean build test

.PHONY: vendor
vendor: prepare_glide
	${GOENV} glide install --strip-vendor

.PHONY: vendor-no-color
vendor-no-color:
	${GOENV} glide --no-color install --strip-vendor

.PHONY: profile-engine
profile-engine:
	@echo "Profiling CPU for 30 seconds"
	${GO} test -bench . -benchtime 30s ./pkg/slinga/engine/apply -cpuprofile cpu.out
	${GO} tool pprof -svg apply.test cpu.out > profile.svg

.PHONY: coverage
coverage:
	@echo "Calculating code coverage"
	touch coverage.tmp && echo 'mode: atomic' > coverage.out && ${GO} list ./... | xargs -n1 -I{} sh -c '${GO} test -short -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out' && rm coverage.tmp
	${GO} tool cover -html=coverage.out -o coverage.html

.PHONY: test
test:
	${GO} test -short -v ./...
	@echo "\nAll unit tests passed"

.PHONY: test-race
test-race:
	CGO_ENABLED=1 go test -race -short -v ./...
	@echo "\nNo race conditions detected. Unit tests passed"

.PHONY: alltest
alltest:
	${GO} test -v ./...
	@echo "\nAll unit & integration tests passed"

.PHONY: test-loop
test-loop:
	while ${GO} test -v ./...; do :; done

.PHONY: clean-run-noop
clean-run-noop:
	$(eval TMP := $(shell mktemp -d))
	${GOENV} APTOMI_DB=$(TMP) tools/demo-local-policy-init.sh

.PHONY: smoke
smoke: install alltest clean-run-noop
	-rm -f aptomi aptomictl

.PHONY: build
build:
	${GO} build ${GOFLAGS} -v -i ./...
	${GO} build ${GOFLAGS} -v -i -o aptomi github.com/Aptomi/aptomi/cmd/aptomi
	${GO} build ${GOFLAGS} -v -i -o aptomictl github.com/Aptomi/aptomi/cmd/aptomictl

.PHONY: install
install: build
	${GO} install -v ${GOFLAGS} github.com/Aptomi/aptomi/cmd/aptomi
	${GO} install -v ${GOFLAGS} github.com/Aptomi/aptomi/cmd/aptomictl

.PHONY: fmt
fmt:
	${GO} fmt ./...

.PHONY: vet
vet:
	${GO} tool vet -all -shadow ./cmd ./pkg || echo "\nSome vet checks failed\n"

.PHONY: lint
lint: prepare_gometalinter
	${GOENV} gometalinter --deadline=120s ./pkg/... ./cmd/... | grep -v 'should not use dot imports'

.PHONY: validate
validate: fmt vet lint
	@echo "\nAll validations passed"

.PHONY: clean
clean:
	-rm -f aptomi aptomictl
	${GO} clean -r -i

HAS_GOMETALINTER := $(shell command -v gometalinter)

.PHONY: prepare_gometalinter
prepare_gometalinter:
ifndef HAS_GOMETALINTER
	go get -u -v -d github.com/alecthomas/gometalinter && \
	go install -v github.com/alecthomas/gometalinter && \
	gometalinter --install --update
endif

HAS_GLIDE := $(shell command -v glide)

.PHONY: prepare_glide
prepare_glide:
ifndef HAS_GLIDE
	curl https://glide.sh/get | sh
endif


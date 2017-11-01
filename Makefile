GIT_VERSION=$(shell git describe --tags --long --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GOENV=CGO_ENABLED=0
GOFLAGS=-ldflags "-X github.com/Aptomi/aptomi/pkg/version.gitVersion=${GIT_VERSION} -X github.com/Aptomi/aptomi/pkg/version.gitCommit=${GIT_COMMIT} -X github.com/Aptomi/aptomi/pkg/version.buildDate=${BUILD_DATE}"
GO=${GOENV} go

.PHONY: default
default: clean build test

.PHONY: vendor
vendor: prepare_glide
ifndef JENKINS_HOME
	${GOENV} glide install --strip-vendor
else
	${GOENV} glide --no-color install --strip-vendor
endif
	${GO} build -o bin/protoc-gen-go ./vendor/github.com/golang/protobuf/protoc-gen-go
	tools/setup-apimachinery.sh

.PHONY: profile-engine
profile-engine:
	@echo "Profiling CPU for 30 seconds"
	${GO} test -bench . -benchtime 30s ./pkg/engine/apply -cpuprofile cpu.out
	${GO} tool pprof -svg apply.test cpu.out > profile.svg

.PHONY: coverage
coverage:
	@echo "Calculating code coverage (unit tests only)"
	echo 'mode: atomic' > coverage.out && ${GO} list ./... | xargs -n1 -I{} sh -c "echo -n '' > coverage.tmp && ${GO} test -short -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out" && rm coverage.tmp
	${GO} tool cover -html=coverage.out -o coverage.html

.PHONY: coverage-full
coverage-full:
	@echo "Calculating code coverage (unit and integration tests)"
	echo 'mode: atomic' > coverage.out && ${GO} list ./... | xargs -n1 -I{} sh -c "echo -n '' > coverage.tmp && ${GO} test -covermode=atomic -coverprofile=coverage.tmp {} && tail -n +2 coverage.tmp >> coverage.out" && rm coverage.tmp
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
alltest: prepare_go_junit_report
ifndef JENKINS_HOME
	${GO} test -v ./...
else
	${GO} test -v ./... 2>&1 | go-junit-report | tee junit.xml
endif
	${GO} test -bench . -count 1 ./pkg/engine/...
	@echo "\nAll tests passed (unit, integration, benchmark)"

.PHONY: test-loop
test-loop:
	while ${GO} test -v ./...; do :; done

.PHONY: smoke
smoke: install alltest
	tools/smoke.sh

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

.PHONY: clean
clean:
	-rm -f aptomi aptomictl
	${GO} clean -r -i

.PHONY: lint
lint: prepare_gometalinter
ifdef JENKINS_HOME
	${GOENV} gometalinter --config=gometalinter.json --checkstyle ./pkg/... ./cmd/... | tee checkstyle.xml
else
	${GOENV} gometalinter --concurrency=2 --config=gometalinter.json ./pkg/... ./cmd/...
endif

HAS_GOMETALINTER := $(shell command -v gometalinter)

.PHONY: prepare_gometalinter
prepare_gometalinter:
ifndef HAS_GOMETALINTER
	go get -u -v -d github.com/alecthomas/gometalinter && \
	go install -v github.com/alecthomas/gometalinter && \
	gometalinter --install --update
endif

.PHONY: toc
toc: prepare_doctoc
	doctoc README.md

HAS_DOCTOC := $(shell command -v doctoc)

.PHONY: prepare_doctoc
prepare_doctoc:
ifndef HAS_DOCTOC
	npm install -g doctoc
endif

HAS_GLIDE := $(shell command -v glide)

.PHONY: prepare_glide
prepare_glide:
ifndef HAS_GLIDE
	go get -u -v github.com/Masterminds/glide
endif

HAS_GO_JUNIT_REPORT := $(shell command -v go-junit-report)

.PHONY: prepare_go_junit_report
prepare_go_junit_report:
ifndef HAS_GO_JUNIT_REPORT
	go get -u -v github.com/jstemmer/go-junit-report
endif

.PHONY: w-dep
w-dep:
	cd webui; npm install

.PHONY: w-test
w-test:
	cd webui; npm test

.PHONY: w-test-unit
w-test-unit:
	cd webui; npm run unit

.PHONY: w-test-e2e
w-test-e2e:
	cd webui; npm run e2e

.PHONY: w-dev
w-dev:
	cd webui; npm run dev

.PHONY: w-build
w-build:
	cd webui; npm run build

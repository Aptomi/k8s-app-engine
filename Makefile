GIT_VERSION=dev-$(shell git describe --tags --long --dirty)
GIT_COMMIT=$(shell git rev-parse HEAD)
BUILD_DATE=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GOENV=env CGO_ENABLED=0
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
	cd webui; npm install

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

.PHONY: coverage-publish
coverage-publish: prepare_goveralls
	BUILD_NUMBER="" goveralls -coverprofile coverage.out

.PHONY: test
test:
	tools/with_etcd.sh ${GO} test -short -v ./...
	@echo "\nAll unit tests passed"

.PHONY: test-race
test-race:
	CGO_ENABLED=1 go test -race -short -v ./...
	@echo "\nNo race conditions detected. Unit tests passed"

.PHONY: alltest
alltest: prepare_go_junit_report
ifndef JENKINS_HOME
	tools/with_etcd.sh ${GO} test -v ./...
else
	tools/with_etcd.sh ${GO} test -v ./... 2>&1 | go-junit-report | tee junit.xml
endif
	${GO} test -bench . -count 1 ./pkg/engine/...
	@echo "\nAll tests passed (unit, integration, benchmark)"

.PHONY: test-loop
test-loop:
	while ${GO} test -v ./...; do :; done

.PHONY: smoke
smoke: install alltest
	tools/smoke.sh

.PHONY: embed-ui
embed-ui: prepare_filebox
	find webui -type f | grep -v '/node_modules' | grep -v '/dist' | sort | xargs shasum 2>/dev/null | shasum > .ui_hash_current
	if diff .ui_hash_current .ui_hash_previous; then echo 'No changes in UI. Skipping UI build'; else rm -rf pkg/server/ui/*b0x*; cd webui; npm run build; cd ..; fileb0x webui/b0x.yaml; fi
	cp .ui_hash_current .ui_hash_previous

#
# IMPORTANT
#
# Make sure to update .goreleaser.yml if doing any changes to build process
#
.PHONY: build
build: embed-ui
	${GO} build ${GOFLAGS} -v -i ./...
	${GO} build ${GOFLAGS} -v -i -o aptomi github.com/Aptomi/aptomi/cmd/aptomi
	${GO} build ${GOFLAGS} -v -i -o aptomictl github.com/Aptomi/aptomi/cmd/aptomictl

#
# IMPORTANT
#
# Make sure to update .goreleaser.yml if doing any changes to install process
#
.PHONY: install
install: build
	${GO} install -v ${GOFLAGS} github.com/Aptomi/aptomi/cmd/aptomi
	${GO} install -v ${GOFLAGS} github.com/Aptomi/aptomi/cmd/aptomictl

.PHONY: release
release: prepare_goreleaser
	goreleaser --rm-dist
	tools/publish-charts.sh

.PHONY: fmt
fmt:
	${GO} fmt ./...
	${GOENV} goimports -w cmd examples pkg

.PHONY: clean
clean:
	-rm -f aptomi aptomictl .ui_hash_current .ui_hash_previous
	${GO} clean -r -i

.PHONY: lint
lint: prepare_golangci_lint embed-ui
ifdef JENKINS_HOME
	${GOENV} golangci-lint run --out-format checkstyle | tee checkstyle.xml
else
	${GOENV} golangci-lint run
endif

HAS_GOLANGCI_LINT := $(shell command -v golangci-lint)

.PHONY: prepare_golangci_lint
prepare_golangci_lint:
ifndef HAS_GOLANGCI_LINT
	${GO} get -u gopkg.in/golangci/golangci-lint.v1/cmd/golangci-lint
endif

.PHONY: toc
toc: prepare_doctoc
	doctoc README.md
	doctoc docs/language.md

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
	${GO} get -u github.com/Masterminds/glide
endif

HAS_GO_JUNIT_REPORT := $(shell command -v go-junit-report)

.PHONY: prepare_go_junit_report
prepare_go_junit_report:
ifndef HAS_GO_JUNIT_REPORT
	${GO} get -u github.com/jstemmer/go-junit-report
endif

HAS_GOVERALLS := $(shell command -v goveralls)

.PHONY: prepare_goveralls
prepare_goveralls:
ifndef HAS_GOVERALLS
	${GO} get -u github.com/mattn/goveralls
endif

HAS_FILEBOX := $(shell command -v fileb0x)

.PHONY: prepare_filebox
prepare_filebox:
ifndef HAS_FILEBOX
	${GO} get -u github.com/UnnoTed/fileb0x
endif

HAS_GORELEASER := $(shell command -v goreleaser)

.PHONY: prepare_goreleaser
prepare_goreleaser:
ifndef HAS_GORELEASER
	${GO} get -u github.com/goreleaser/goreleaser
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

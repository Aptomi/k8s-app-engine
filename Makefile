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
	${GOENV} glide install --strip-vendor

.PHONY: vendor-no-color
vendor-no-color:
	${GOENV} glide --no-color install --strip-vendor

.PHONY: profile-engine
profile-engine:
	@echo "Profiling CPU for 30 seconds"
	${GO} test -bench . -benchtime 30s ./pkg/engine/apply -cpuprofile cpu.out
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
	${GO} test -bench . -count 1 ./pkg/engine/...
	@echo "\nAll tests passed (unit, integration, benchmark)"

.PHONY: test-loop
test-loop:
	while ${GO} test -v ./...; do :; done

.PHONY: clean-run-noop
clean-run-noop:
	$(eval TMP := $(shell mktemp -d))
	${GOENV} APTOMI_DB=$(TMP) tools/demo-local-policy-init.sh

.PHONY: smoke
smoke: install alltest

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

.PHONY: lint
lint: prepare_gometalinter
	${GOENV} gometalinter --config=gometalinter.json --deadline=180s ./pkg/... ./cmd/...

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

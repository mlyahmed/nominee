#!/usr/bin/make -f

### Variables
SHELL = /bin/bash

HARD_BUILD ?=
DOCKER_USERNAME ?= mlyahmed
DOCKER_PASSWORD ?=

export NOMINEE_NETWORK ?= nominee
export GOARCH ?= $(shell go env GOARCH)
export GOOS ?= $(shell go env GOOS)
export NOMINEE_DOCKER_REPO := nominee
export NOMINEE_ARTIFACTS := nominee-postgres nominee-haproxy
export NOMINEE_BIN_DIR := bin
export BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
export SIMPLE_VERSION := $(shell (test "$(shell git describe)" = "$(shell git describe --abbrev=0)" && echo $(shell git describe)) || echo $(shell git describe --abbrev=0)-$(shell git branch --show-current))
export GIT_VERSION := $(shell git describe --dirty --tags --always)
export GIT_COMMIT := $(shell git rev-parse HEAD)
export IMAGE_VERSION ?= $(SIMPLE_VERSION)
export MODULE := $(shell go list -m)
export GO111MODULE = on
export CGO_ENABLED = 0
export GO_ASMFLAGS = -asmflags=all=-trimpath=./...
export GO_GCFLAGS = -gcflags=all=-trimpath=./...
export GO_BUILD_ARGS = \
  $(GO_GCFLAGS) $(GO_ASMFLAGS) \
  -ldflags="-s -w \
    -X '$(MODULE)/pkg/version.Date=$(BUILD_DATE)' \
    -X '$(MODULE)/pkg/version.Platform=$(GOOS)/$(GOARCH)' \
    -X '$(MODULE)/pkg/version.SimpleVersion=$(SIMPLE_VERSION)' \
    -X '$(MODULE)/pkg/version.GitVersion=$(GIT_VERSION)' \
    -X '$(MODULE)/pkg/version.GitCommit=$(GIT_COMMIT)' \
    -X '$(MODULE)/pkg/version.ImageVersion=$(IMAGE_VERSION)'"

rm-all create-docker-network: export NOMINEE_NETWORK_EXISTS := $(shell docker network ls | grep $(NOMINEE_NETWORK))

### Debug Tools
print-%: ; @echo $* = '$($*)' from $(origin $*)

### Assertions
assert-command-present = $(if $(shell which $1),,$(error '$1' missing and needed for this build))
build-binaries test fix: export _check := $(call assert-command-present,go)
rm-% build-image%: export _check := $(call assert-command-present,docker)
start-% stop-% logs-%: export _check := $(call assert-command-present,docker-compose)

### Build Rules
.PHONY: all
all: clean fix test build-binaries build-images; $(info all done.)

.PHONY: build-images
build-images: $(foreach artifact, $(NOMINEE_ARTIFACTS), build-image-$(artifact))
build-image-%: build-binaries
	$(info build docker image $*)
	@docker build -t $(NOMINEE_DOCKER_REPO)/$*:$(IMAGE_VERSION) -f images/$*/Dockerfile .

.PHONY: push-images
push-images: docker-login $(foreach artifact, $(NOMINEE_ARTIFACTS), push-image-$(artifact)) docker-logout
push-image-%:
	$(info push docker image $*)
	docker push $(NOMINEE_DOCKER_REPO)/$*:$(IMAGE_VERSION)

.PHONY: docker-login
docker-login:
	$(if $(DOCKER_PASSWORD), @echo $(DOCKER_PASSWORD) | docker login -u $(DOCKER_USERNAME) --password-stdin)

.PHONY: docker-logout
docker-logout:
	$(if $(DOCKER_PASSWORD), @docker logout)

.PHONY: build-binaries
build-binaries:
	$(info build binaries...)
	@mkdir -p $(NOMINEE_BIN_DIR)
	go build $(GO_BUILD_ARGS) -o $(NOMINEE_BIN_DIR) ./...

.PHONY: test
test:
	$(info test...)
	@go test ./... -coverprofile report.coverage

.PHONY: fix
fix:
	$(info fix up...)
	@go mod vendor
	@go mod tidy
	@go fmt ./...

.PHONY: clean
clean:
	$(info clean up...)
	@rm -rf $(NOMINEE_BIN_DIR)
	@go clean $(if $(HARD_BUILD),-cache -testcache -modcache,)

### Run Rules
.PHONY: start-all start-postgres start-etcd start-haproxy
start-all: start-etcd start-haproxy start-postgres
start-postgres: start-docker-compose-etcd start-docker-compose-postgres
start-etcd: start-docker-compose-etcd
start-haproxy: start-docker-compose-etcd start-docker-compose-haproxy
start-docker-compose-%: create-docker-network
	$(info starting $*...)
	docker-compose -f hack/$*/docker-compose.yaml up -d

.PHONY: stop-all stop-postgres stop-etcd stop-haproxy
stop-all: stop-postgres stop-haproxy stop-etcd
stop-postgres: stop-docker-compose-postgres
stop-etcd: stop-docker-compose-etcd
stop-haproxy: stop-docker-compose-haproxy
stop-docker-compose-%:
	$(info stopping $*...)
	@docker-compose -f hack/$*/docker-compose.yaml stop

.PHONY: logs-postgres logs-etcd logs-haproxy
logs-postgres: logs-docker-compose-postgres
logs-etcd: logs-docker-compose-etcd
logs-haproxy: logs-docker-compose-haproxy
logs-docker-compose-%:
	@docker-compose -f hack/$*/docker-compose.yaml logs -f

.PHONY: rm-all rm-postgres rm-etcd rm-haproxy
rm-all: stop-all rm-postgres rm-haproxy rm-etcd
	$(if $(NOMINEE_NETWORK_EXISTS), @docker network rm $(NOMINEE_NETWORK), $(info the network $(NOMINEE_NETWORK) does not exist))
rm-postgres: stop-postgres rm-docker-compose-postgres
rm-etcd: stop-etcd rm-docker-compose-etcd
rm-haproxy: stop-haproxy rm-docker-compose-haproxy
rm-docker-compose-%:
	@docker-compose -f hack/$*/docker-compose.yaml rm -fsv

.PHONY: create-docker-network
create-docker-network:
	@docker network create --driver bridge $(NOMINEE_NETWORK) 2>/dev/null || true

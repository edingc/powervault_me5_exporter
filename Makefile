BINARY      := prometheus-powervault-me5-exporter
MODULE      := github.com/edingc/powervault_me5_exporter
IMAGE       := powervault-me5-exporter
GOLANG_CROSS_VERSION ?= v1.25.7

# Inject version info at build time
VERSION     ?= $(shell cat VERSION || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BRANCH      ?= $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := \
  -X github.com/prometheus/common/version.Version=$(VERSION) \
  -X github.com/prometheus/common/version.Revision=$(COMMIT) \
  -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE) \
  -X github.com/prometheus/common/version.Branch=$(BRANCH) \
  -w -s

.PHONY: all build test lint fmt vet clean docker-build help

all: build

## build: compile the binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/powervault_me5_exporter

## test: run all tests
test:
	go test -race -coverprofile=coverage.txt -covermode=atomic ./...

## lint: run golangci-lint (install: https://golangci-lint.run/usage/install/)
lint:
	golangci-lint run ./...

## fmt: format all Go source files
fmt:
	gofmt -s -w .

## vet: run go vet
vet:
	go vet ./...

## clean: remove build artifacts
clean:
	rm -f $(BINARY) coverage.txt

## docker-build: build Docker image
docker-build:
	docker build \
	  --build-arg VERSION=$(VERSION) \
	  --build-arg COMMIT=$(COMMIT) \
	  --build-arg BUILD_DATE=$(BUILD_DATE) \
	  --build-arg BRANCH=$(BRANCH) \
	  -t $(IMAGE):$(VERSION) \
	  -t $(IMAGE):latest \
	  .

release:
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
	--env-file .release-env -v `pwd`:/work -w /work \
	ghcr.io/goreleaser/goreleaser-cross:$(GOLANG_CROSS_VERSION) \
	release --clean

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

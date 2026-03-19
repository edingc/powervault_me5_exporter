BINARY      := prometheus-powervault-me5-exporter
MODULE      := github.com/edingc/powervault_me5_exporter
IMAGE       := powervault-me5-exporter

# Inject version info at build time
VERSION     ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT      ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE  ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

LDFLAGS := \
  -X github.com/prometheus/common/version.Version=$(VERSION) \
  -X github.com/prometheus/common/version.Revision=$(COMMIT) \
  -X github.com/prometheus/common/version.BuildDate=$(BUILD_DATE) \
  -X github.com/prometheus/common/version.Branch=$(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "unknown") \
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
	  -t $(IMAGE):$(VERSION) \
	  -t $(IMAGE):latest \
	  .

## help: show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //'

.PHONY: build debug install test upload lint

# go env
GOPROXY     := "https://goproxy.cn,direct"
GOOS        := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
CGO_LDFLAGS := "-static"
CGO_CFLAGS  := "-D_LARGEFILE64_SOURCE"
CC          := musl-gcc

GOENV := GO111MODULE=on
GOENV += GOPROXY=$(GOPROXY)
GOENV += CC=$(CC)
GOENV += CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_CFLAGS=$(CGO_CFLAGS)
GOENV += GOOS=$(GOOS) GOARCH=$(GOARCH)
GOLANGCILINT_VERSION ?= v1.50.0
GOBIN := $(shell go env GOPATH)/bin
GOBIN_GOLANGCILINT := $(shell which $(GOBIN)/golangci-lint)
# go
GO := go

# output
OUTPUT := bin/curveadm

# build flags
LDFLAGS := -s -w
LDFLAGS += -extldflags "-static -fpic"
LDFLAGS += -X github.com/opencurve/curveadm/cli/cli.CommitId=$(shell git rev-parse --short HEAD)

BUILD_FLAGS := -a
BUILD_FLAGS += -trimpath
BUILD_FLAGS += -ldflags '$(LDFLAGS)'
BUILD_FLAGS += $(EXTRA_FLAGS)

# debug flags
GCFLAGS := "all=-N -l"

DEBUG_FLAGS := -gcflags=$(GCFLAGS)

# go test
GO_TEST ?= $(GO) test

# test flags
CASE ?= "."

TEST_FLAGS := -v
TEST_FLAGS += -p 3
TEST_FLAGS += -cover
TEST_FLAGS += -count=1
TEST_FLAGS += $(DEBUG_FLAGS)
TEST_FLAGS += -run $(CASE)

# packages
PACKAGES := $(PWD)/cmd/curveadm/main.go

# tar
VERSION := "unknown"

build:
	$(GOENV) $(GO) build -o $(OUTPUT) $(BUILD_FLAGS) $(PACKAGES)

debug:
	$(GOENV) $(GO) build -o $(OUTPUT) $(DEBUG_FLAGS) $(PACKAGES)

install:
	cp bin/curveadm ~/.curveadm/bin

test:
	$(GO_TEST) $(TEST_FLAGS) ./...

upload:
	@NOSCMD=$(NOSCMD) bash build/package/upload.sh $(VERSION)

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VERSION)
	$(GOBIN_GOLANGCILINT) run -v


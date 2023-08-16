.PHONY: build debug test tar

# go env
GOPROXY     := "https://goproxy.cn,direct"
GOOS        := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
CGO_LDFLAGS := "-static"
CC          := musl-gcc

GOENV := GO111MODULE=on
GOENV += GOPROXY=$(GOPROXY)
GOENV += CC=$(CC)
GOENV += CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS)
GOENV += GOOS=$(GOOS) GOARCH=$(GOARCH)
GOLANGCILINT_VERSION ?= v1.50.0
GOBIN := $(shell go env GOPATH)/bin
GOBIN_GOLANGCILINT := $(shell which $(GOBIN)/golangci-lint)
# go
GO := go

# output
OUTPUT := bin/curveadm
SERVER_OUTPUT := bin/pigeon

# build flags
LDFLAGS := -s -w
LDFLAGS += -extldflags "-static -fpic"

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
SERVER_PACKAGES := $(PWD)/cmd/service/main.go

# tar
VERSION := "unknown"

build: fmt vet
	$(GOENV) $(GO) build -o $(OUTPUT) $(BUILD_FLAGS) $(PACKAGES)
	$(GOENV) $(GO) build -o $(SERVER_OUTPUT) $(BUILD_FLAGS) $(SERVER_PACKAGES)


debug: fmt vet
	$(GOENV) $(GO) build -o $(OUTPUT) $(DEBUG_FLAGS) $(PACKAGES)
	$(GOENV) $(GO) build -o $(SERVER_OUTPUT) $(DEBUG_FLAGS) $(SERVER_PACKAGES)


test:
	$(GO_TEST) $(TEST_FLAGS) ./...

upload:
	@NOSCMD=$(NOSCMD) bash build/package/upload.sh $(VERSION)

lint:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCILINT_VERSION)
	$(GOBIN_GOLANGCILINT) run -v

fmt:
	go fmt ./...

vet:
	go vet ./...

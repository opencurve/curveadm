.PHONY: build

# go env
GOOS        := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
GOARCH      := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))
CGO_LDFLAGS := "-static"
CC          := musl-gcc
GOENV       := CC=$(CC) CGO_ENABLED=1 CGO_LDFLAGS=$(CGO_LDFLAGS) GOOS=$(GOOS) GOARCH=$(GOARCH)

# go
GO := go

# output
OUTPUT := bin/curveadm

# build flags
LD_FLAGS := -s -w
LD_FLAGS += -extldflags "-static -fpic"

BUILD_FLAGS := -a
BUILD_FLAGS += -trimpath
BUILD_FLAGS += -ldflags '$(LD_FLAGS)'
BUILD_FLAGS += $(EXTRA_FLAGS)

# packages
PACKAGES := $(PWD)/cmd/curveadm/main.go

build:
	$(GOENV) $(GO) build -o $(OUTPUT) $(BUILD_FLAGS) $(PACKAGES)
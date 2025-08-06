NAME := notion-mindmap-server
GOOS ?= darwin
GOARCH ?= arm64

BUILD_DATE            := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT            := $(shell git rev-parse --short HEAD)
GIT_TAG               := $(shell git describe --exact-match --tags --abbrev=0  2> /dev/null || echo untagged)

# Check if GIT_TAG matches version pattern (v*.*.*)
ifneq ($(filter v%,$(GIT_TAG)),)
RELEASE_TAG           := true
else
RELEASE_TAG           := false
endif

VERSION ?= latest 

ifeq ($(VERSION),latest)
ifeq ($(RELEASE_TAG),true)
VERSION := $(GIT_TAG)
endif
endif

override LDFLAGS += \
  -X main.version=$(VERSION) \
  -X main.commitHash=$(GIT_COMMIT) \
  -X main.buildDate=$(BUILD_DATE)

IMAGE_REGISTRY_HOST ?= "ghcr.io/the-end-of-the-human-era-has-arrived"
PLATFORMS ?= linux/amd64,linux/arm64
CTR_CLI ?= docker

build:
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) GO111MODULE=on go build --ldflags "$(LDFLAGS)" -a -o $(BIN_DIR)/$(NAME) cmd/*.go

image-build:
	$(CTR_CLI) buildx build -t $(IMAGE_REGISTRY_HOST)/$(NAME):$(VERSION) --platform $(PLATFORMS) --build-arg LDFLAGS="$(LDFLAGS)" -f Dockerfile .

image-push:
	$(CTR_CLI) buildx build -t $(IMAGE_REGISTRY_HOST)/$(NAME):$(VERSION) --platform $(PLATFORMS) --build-arg LDFLAGS="$(LDFLAGS)" -f Dockerfile --push .

fmt: __golangci-lint
	$(GOLANGCI_LINT) fmt -c $(GOLANGCI_LINT_CFG)

lint: __golangci-lint
	$(GOLANGCI_LINT) run -c $(GOLANGCI_LINT_CFG)

BIN_DIR ?= bin
$(BIN_DIR):
	mkdir -p $(BIN_DIR)

GOLANGCI_LINT := $(BIN_DIR)/golangci-lint
GOLANGCI_LINT_CFG ?= .golangci.yaml
GOLANGCI_LINT_VERSION ?= v2.3.0

.PHONY: golangci-lint
__golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(BIN_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $(BIN_DIR) $(GOLANGCI_LINT_VERSION)

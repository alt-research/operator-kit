## Tool Versions
# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
# https://kubernetes.io/releases
ENVTEST_K8S_VERSION ?= 1.25.0

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

ARCH=$(shell uname -m)
ifeq (x86_64, $(ARCH))
	ARCH=amd64
else ifeq (aarch64, $(ARCH))
	ARCH=arm64
endif
OS=$(shell uname -s)
os=$(shell uname -s | tr '[:upper:]' '[:lower:]')

# Setting SHELL to bash allows bash commands to be executed by recipes.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: help

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development
.PHONY: lint-docker
lint-docker: ## Lint code
	docker run --rm -it -v $(shell pwd):/code -e GOPROXY=$(shell go env GOPROXY) -e HTTPS_PROXY -e https_proxy -w /code golangci/golangci-lint golangci-lint run -v --timeout 5m0s

.PHONY: lint
lint: linter ## Lint code
	$(LINTER) run -v

.PHONY: pre-commit
pre-commit: fmt lint ## Run this before doing commit to make sure everything is up to date

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test ./... -coverprofile cover.out

.PHONY: addlicense
addlicense: addlicense-bin ## Add license to all files
	addlicense -f LICENSE-HEADER -ignore ".github/**/*" -ignore "**/*.yaml" -ignore "**/*.yml" . 

##@ Build Dependencies

## Location to install dependencies to
SCRIPTS = scripts

LOCALBIN ?= bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

ABS_LOCALBIN=$(shell pwd)/$(LOCALBIN)
## Tool Binaries
ENVTEST ?= $(LOCALBIN)/setup-envtest
LINTER ?= $(LOCALBIN)/golangci-lint
ADDLICENSE ?= $(LOCALBIN)/addlicense

.PHONY: linter
linter: $(LINTER) ## instal golangci-lint
$(LINTER): $(LOCALBIN)
	echo $(LINTER)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(ABS_LOCALBIN)
	$(LINTER) --version

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(ABS_LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: addlicense-bin
addlicense-bin: $(ADDLICENSE) ## Download addlicense locally if necessary.
$(ADDLICENSE): $(LOCALBIN)
	test -s $(LOCALBIN)/addlicense || GOBIN=$(ABS_LOCALBIN) go install github.com/google/addlicense@latest

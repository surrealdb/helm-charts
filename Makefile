GO ?= go

# Pin tool versions here. Install into a gitignored local directory via `make tools`.
HELM_DOCS_VERSION ?= 1.14.2
TOOLS_DIR := $(CURDIR)/tools
TOOLS_BIN := $(TOOLS_DIR)/bin
HELM_DOCS := $(TOOLS_BIN)/helm-docs

UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)

ifeq ($(UNAME_S),Darwin)
	HELM_DOCS_OS := Darwin
else ifeq ($(UNAME_S),Linux)
	HELM_DOCS_OS := Linux
else
	$(error Unsupported OS for helm-docs install: $(UNAME_S))
endif

ifeq ($(UNAME_M),x86_64)
	HELM_DOCS_ARCH := x86_64
else ifeq ($(UNAME_M),amd64)
	HELM_DOCS_ARCH := x86_64
else ifeq ($(UNAME_M),arm64)
	HELM_DOCS_ARCH := arm64
else ifeq ($(UNAME_M),aarch64)
	HELM_DOCS_ARCH := arm64
else
	$(error Unsupported architecture for helm-docs install: $(UNAME_M))
endif

HELM_DOCS_ARCHIVE := helm-docs_$(HELM_DOCS_VERSION)_$(HELM_DOCS_OS)_$(HELM_DOCS_ARCH).tar.gz
HELM_DOCS_URL := https://github.com/norwoodj/helm-docs/releases/download/v$(HELM_DOCS_VERSION)/$(HELM_DOCS_ARCHIVE)

# Shared helm-docs template (relative to --chart-search-root). Chart-local README.md.gotmpl is not used.
HELM_DOCS_FLAGS := --chart-search-root=charts --template-files=./README.md.gotmpl --badge-style=for-the-badge

.PHONY: tools
tools: $(HELM_DOCS)

# Reinstall when HELM_DOCS_VERSION changes.
$(HELM_DOCS): $(TOOLS_BIN)/.helm-docs-$(HELM_DOCS_VERSION)
	@true

$(TOOLS_BIN)/.helm-docs-$(HELM_DOCS_VERSION):
	mkdir -p $(TOOLS_BIN)
	rm -f $(TOOLS_BIN)/helm-docs $(TOOLS_BIN)/.helm-docs-*
	curl -sSL "$(HELM_DOCS_URL)" | tar -xz -C $(TOOLS_BIN) helm-docs
	chmod +x $(HELM_DOCS)
	$(HELM_DOCS) --version
	touch $@

.PHONY: docs
docs: $(HELM_DOCS)
	$(HELM_DOCS) $(HELM_DOCS_FLAGS)

.PHONY: build
build: docs
	$(GO) build

.PHONY: clean
clean:
	$(GO) clean -modcache

.PHONY: test
test: docs
	$(GO) clean -testcache
	$(GO) test -v -cover ./...

.PHONY: update-test-snapshots
update-test-snapshots: docs
	$(GO) clean -testcache
	UPDATE_SNAPSHOT="deployment.yaml/*" $(GO) test -v -cover ./...

.PHONY: lint
lint:
	golangci-lint run

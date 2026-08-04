.PHONY: all build run test lint clean init deps build-all package checksum release github-release test-release

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet
GOMOD=$(GOCMD) mod

# Binary name and app
BINARY=portx
CMD_APP=cmd/portx
README=README.md

# Output directories
DIST=dist
BUILD_DIR=$(DIST)/build
PKG_DIR=$(DIST)/pkg

# Version ldflags (injected at build time)
VERSION ?= dev
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "")
DATE := $(shell date -u +%Y-%m-%dT%H:%M:%SZ 2>/dev/null || echo "")
LDFLAGS=-s -w -X github.com/alexperezortuno/portx/internal/version.Version=$(VERSION) -X github.com/alexperezortuno/portx/internal/version.Commit=$(COMMIT) -X github.com/alexperezortuno/portx/internal/version.Date=$(DATE)

# Platforms and architectures
PLATFORMS=linux darwin windows
ARCHS=amd64 arm64

all: init build test

init:
	$(GOMOD) tidy

run:
	$(GOCMD) run ./$(CMD_APP)

fmt:
	$(GOFMT) -w ./...

vet:
	$(GOVET) ./...

lint: fmt vet

clean:
	rm -rf $(DIST)

build:
	$(GOBUILD) -ldflags "$(LDFLAGS)" -o bin/$(BINARY) ./$(CMD_APP)

deps:
	$(GOMOD) download
	$(GOMOD) tidy

test:
	$(GOTEST) -v -race ./...

test-coverage:
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

# Cross-compile for all platforms
build-all:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-amd64/$(BINARY) ./$(CMD_APP)
	GOOS=linux GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/linux-arm64/$(BINARY) ./$(CMD_APP)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/darwin-amd64/$(BINARY) ./$(CMD_APP)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/darwin-arm64/$(BINARY) ./$(CMD_APP)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/windows-amd64/$(BINARY).exe ./$(CMD_APP)
	GOOS=windows GOARCH=arm64 $(GOBUILD) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/windows-arm64/$(BINARY).exe ./$(CMD_APP)

# Package each platform build into a zip with README.md
package: build-all
	mkdir -p $(PKG_DIR)
	cp $(README) $(BUILD_DIR)/linux-amd64/
	cd $(BUILD_DIR)/linux-amd64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_linux_amd64.zip" $(BINARY) $(README)
	cp $(README) $(BUILD_DIR)/linux-arm64/
	cd $(BUILD_DIR)/linux-arm64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_linux_arm64.zip" $(BINARY) $(README)
	cp $(README) $(BUILD_DIR)/darwin-amd64/
	cd $(BUILD_DIR)/darwin-amd64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_darwin_amd64.zip" $(BINARY) $(README)
	cp $(README) $(BUILD_DIR)/darwin-arm64/
	cd $(BUILD_DIR)/darwin-arm64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_darwin_arm64.zip" $(BINARY) $(README)
	cp $(README) $(BUILD_DIR)/windows-amd64/
	cd $(BUILD_DIR)/windows-amd64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_windows_amd64.zip" $(BINARY).exe $(README)
	cp $(README) $(BUILD_DIR)/windows-arm64/
	cd $(BUILD_DIR)/windows-arm64 && zip -q "$(abspath $(PKG_DIR))/$(BINARY)_$(VERSION)_windows_arm64.zip" $(BINARY).exe $(README)
	@echo "Packages created in $(PKG_DIR)/"

# Generate checksums for all zip files
checksum: package
	cd $(PKG_DIR) && \
		echo "# $(BINARY) $(VERSION) checksums" > $(BINARY)_$(VERSION)_checksums.txt && \
		for f in *.zip; do \
			if command -v sha256sum > /dev/null 2>&1; then \
				sha256sum "$$f" >> $(BINARY)_$(VERSION)_checksums.txt; \
			elif command -v shasum > /dev/null 2>&1; then \
				shasum -a 256 "$$f" >> $(BINARY)_$(VERSION)_checksums.txt; \
			else \
				openssl sha256 "$$f" | awk '{print $$2, $$1}' >> $(BINARY)_$(VERSION)_checksums.txt; \
			fi; \
		done && \
		echo "Checksums written to $(BINARY)_$(VERSION)_checksums.txt"

# Full release: clean + build + package + checksum
# Usage: make release VERSION=1.0.0
release: clean build-all package checksum
	@echo ""
	@echo "Release $(VERSION) ready in $(PKG_DIR)/"
	@ls -lh $(PKG_DIR)/

# Create GitHub release
# Usage: make github-release VERSION=1.0.0 (GITHUB_TOKEN must be set in environment)
github-release: release
	@if [ -z "$$GITHUB_TOKEN" ]; then \
		echo "GITHUB_TOKEN is not set, skipping GitHub release."; \
	elif ! command -v gh > /dev/null 2>&1; then \
		echo "gh CLI not found, skipping GitHub release."; \
	else \
		tag=$$(git describe --tags --exact-match 2>/dev/null | grep -E '^v[0-9]' || echo ""); \
		if [ -z "$$tag" ]; then \
			echo "Not on a version tag, skipping GitHub release."; \
		else \
			echo "Creating GitHub release for $$tag..."; \
			gh release create "$$tag" \
				--title "PortX $$tag" \
				--notes "PortX $$(echo $$tag | sed 's/v//')" \
				$(PKG_DIR)/$(BINARY)_$(VERSION)_checksums.txt \
				$(PKG_DIR)/*.zip; \
		fi; \
	fi

# Test the full build pipeline locally (no GitHub release)
# Usage: make test-release VERSION=1.0.0
test-release: release
	@echo "Verifying packages..."
	@for f in $(PKG_DIR)/*.zip; do \
		echo "Contents of $$f:"; \
		unzip -l "$$f"; \
		echo ""; \
	done
	@echo "Checksums:"; \
	cat $(PKG_DIR)/$(BINARY)_$(VERSION)_checksums.txt

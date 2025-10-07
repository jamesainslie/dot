.PHONY: build test test-tparse lint clean install check qa help version version-major version-minor version-patch release

# Build variables
BINARY_NAME := dot
CURRENT_VERSION := $(shell git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.0-dev")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Parse semantic version components
MAJOR := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f1)
MINOR := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f2)
PATCH := $(shell echo $(CURRENT_VERSION) | sed 's/v//' | cut -d. -f3 | cut -d- -f1)

# LDFLAGS for version embedding
LDFLAGS := -ldflags "\
	-X main.version=$(VERSION) \
	-X main.commit=$(COMMIT) \
	-X main.date=$(DATE)"

# Default target
.DEFAULT_GOAL := help

## help: Display this help message
help:
	@echo "Available targets:"
	@sed -n 's/^##//p' $(MAKEFILE_LIST) | column -t -s ':' | sed -e 's/^/ /'

## build: Build the application binary
build:
	go build -buildvcs=false $(LDFLAGS) -o $(BINARY_NAME) ./cmd/$(BINARY_NAME)

## test: Run all tests with race detection and coverage
test:
	go test -v -race -cover -coverprofile=coverage.out ./...

## lint: Run golangci-lint
lint:
	golangci-lint run --config .golangci.yml

## vet: Run go vet
vet:
	go vet ./...

## fmt: Check code formatting
fmt:
	@if [ "$(shell gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		gofmt -s -l .; \
		exit 1; \
	fi

## fmt-fix: Fix code formatting
fmt-fix:
	gofmt -s -w .
	goimports -w .

## clean: Remove build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	rm -rf dist/

## install: Install the binary
install: build
	go install -buildvcs=false $(LDFLAGS) ./cmd/$(BINARY_NAME)

## check: Run tests and linting (machine-readable output for CI/AI agents)
check: test lint vet

## qa: Run tests with tparse, linting, and vetting (human-friendly output)
qa: test-tparse lint vet

## test-tparse: Run tests with tparse for formatted output
test-tparse:
	@command -v tparse >/dev/null 2>&1 || { echo "Installing tparse..."; go install github.com/mfridman/tparse@latest; }
	@set -o pipefail && go test -json -race -cover -coverprofile=coverage.out ./... | tparse -all -progress -slow 10

## coverage: Generate coverage report
coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## deps-update: Update all dependencies
deps-update:
	go get -u ./...
	go mod tidy

## deps-verify: Verify dependencies
deps-verify:
	go mod verify

## version: Display current version
version:
	@echo "Current version: $(CURRENT_VERSION)"
	@echo "Next version (patch): v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1)"

## version-major: Bump major version
version-major:
	$(eval NEW_VERSION := v$(shell expr $(MAJOR) + 1).0.0)
	@echo "Bumping major version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release VERSION=$(NEW_VERSION)

## version-minor: Bump minor version
version-minor:
	$(eval NEW_VERSION := v$(MAJOR).$(shell expr $(MINOR) + 1).0)
	@echo "Bumping minor version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release VERSION=$(NEW_VERSION)

## version-patch: Bump patch version
version-patch:
	$(eval NEW_VERSION := v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))
	@echo "Bumping patch version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release VERSION=$(NEW_VERSION)

## release: Create new release (requires VERSION variable)
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set. Use: make release VERSION=v1.2.3"; exit 1; fi
	@echo "Creating release $(VERSION)..."
	@$(MAKE) check
	@echo "Release $(VERSION) ready. Tag with: git tag -a $(VERSION) -m 'Release $(VERSION)'"

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)


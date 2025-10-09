.PHONY: build test test-tparse lint clean install check qa help version version-major version-minor version-patch release release-tag changelog-update

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
check: test check-coverage lint vet

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

## check-coverage: Verify test coverage meets 80% threshold
check-coverage:
	@if [ ! -f coverage.out ]; then \
		echo "Error: coverage.out not found. Run 'make test' first."; \
		exit 1; \
	fi
	@COVERAGE=$$(go tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	THRESHOLD=80.0; \
	echo "Coverage: $${COVERAGE}% (threshold: $${THRESHOLD}%)"; \
	if [ "$$(echo "$${COVERAGE} < $${THRESHOLD}" | bc)" -eq 1 ]; then \
		echo ""; \
		echo "ERROR: Test coverage below threshold"; \
		echo "  Current:   $${COVERAGE}%"; \
		echo "  Required:  $${THRESHOLD}%"; \
		echo "  Shortfall: $$(echo "$${THRESHOLD} - $${COVERAGE}" | bc)%"; \
		echo ""; \
		echo "Add tests to reach 80% coverage."; \
		echo "Run: go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out"; \
		exit 1; \
	fi; \
	echo "Coverage check passed: $${COVERAGE}%"

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

## changelog: Generate CHANGELOG.md from git commits
changelog:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	git-chglog -o CHANGELOG.md

## changelog-next: Preview next version changelog
changelog-next:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	$(eval NEXT_VERSION := v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))
	@echo "Preview of changelog for $(NEXT_VERSION):"
	@echo ""
	git-chglog --next-tag $(NEXT_VERSION) $(CURRENT_VERSION)..

## changelog-update: Update changelog and commit it (internal target)
changelog-update:
	@command -v git-chglog >/dev/null 2>&1 || { echo "Installing git-chglog..."; go install github.com/git-chglog/git-chglog/cmd/git-chglog@latest; }
	@echo "Generating changelog..."
	@git-chglog -o CHANGELOG.md
	@echo "Committing changelog and configuration..."
	@git add CHANGELOG.md .chglog/
	@git commit -m "docs(changelog): update for $(VERSION) release" || true

## version-major: Bump major version and create release
version-major:
	$(eval NEW_VERSION := v$(shell expr $(MAJOR) + 1).0.0)
	@echo "Bumping major version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## version-minor: Bump minor version and create release
version-minor:
	$(eval NEW_VERSION := v$(MAJOR).$(shell expr $(MINOR) + 1).0)
	@echo "Bumping minor version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## version-patch: Bump patch version and create release
version-patch:
	$(eval NEW_VERSION := v$(MAJOR).$(MINOR).$(shell expr $(PATCH) + 1))
	@echo "Bumping patch version: $(CURRENT_VERSION) -> $(NEW_VERSION)"
	@$(MAKE) release-tag VERSION=$(NEW_VERSION)

## release: Verify release readiness (requires VERSION variable)
release:
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set. Use: make release VERSION=v1.2.3"; exit 1; fi
	@echo "Verifying release $(VERSION) readiness..."
	@$(MAKE) check
	@echo "✓ All quality checks passed"
	@echo "Release $(VERSION) verified and ready"

## release-tag: Complete release workflow with changelog and tagging
release-tag:
	@if [ -z "$(VERSION)" ]; then echo "VERSION not set. Use: make release-tag VERSION=v1.2.3"; exit 1; fi
	@echo "Starting release workflow for $(VERSION)..."
	@echo ""
	@echo "Step 1: Running quality checks..."
	@$(MAKE) check
	@echo "✓ Quality checks passed"
	@echo ""
	@echo "Step 2: Updating changelog..."
	@$(MAKE) changelog-update VERSION=$(VERSION)
	@echo "✓ Changelog updated"
	@echo ""
	@echo "Step 3: Creating git tag..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@echo "✓ Tag $(VERSION) created"
	@echo ""
	@echo "Step 4: Regenerating changelog with new tag..."
	@$(MAKE) changelog
	@git add CHANGELOG.md
	@git commit --amend --no-edit
	@echo "✓ Changelog finalized"
	@echo ""
	@echo "Step 5: Moving tag to amended commit..."
	@git tag -f $(VERSION)
	@echo "✓ Tag moved to final commit"
	@echo ""
	@echo "══════════════════════════════════════════════════════════"
	@echo "Release $(VERSION) ready to push!"
	@echo "══════════════════════════════════════════════════════════"
	@echo ""
	@echo "Push with:"
	@echo "  git push origin main"
	@echo "  git push --force origin $(VERSION)"
	@echo ""

## build-all: Build for all platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 ./cmd/$(BINARY_NAME)
	GOOS=linux GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 ./cmd/$(BINARY_NAME)
	GOOS=darwin GOARCH=arm64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 ./cmd/$(BINARY_NAME)
	GOOS=windows GOARCH=amd64 go build -buildvcs=false $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe ./cmd/$(BINARY_NAME)


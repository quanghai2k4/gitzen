# GitZen Makefile - Manual Build Automation
# Minh hoa quy trinh dong goi thu cong (Topic 3: Packaging & Deployment)

# Variables
BINARY_NAME := gitzen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.commit=$(shell git rev-parse HEAD 2>/dev/null || echo "none") -X main.date=$(BUILD_TIME)

# Build directories
BUILD_DIR := build
DIST_DIR := dist

# Go settings
GO := go
GOFLAGS := -trimpath
CGO_ENABLED := 0

# Platforms for cross-compilation
PLATFORMS := linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64

# Colors for output
GREEN := \033[0;32m
YELLOW := \033[0;33m
RED := \033[0;31m
NC := \033[0m # No Color

.PHONY: all build build-all clean install uninstall run test lint fmt vet tidy help version

## Default target
all: build

## Build binary for current platform
build:
	@printf "$(GREEN)%s$(NC)\n" "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gitzen
	@printf "$(GREEN)%s$(NC)\n" "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

## Build binaries for all platforms (cross-compilation)
build-all: clean
	@printf "$(GREEN)%s$(NC)\n" "Building $(BINARY_NAME) $(VERSION) for all platforms..."
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then output=$$output.exe; fi; \
		printf "$(YELLOW)%s$(NC)\n" "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=$(CGO_ENABLED) \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $$output ./cmd/gitzen || exit 1; \
	done
	@printf "$(GREEN)%s$(NC)\n" "All builds complete! Check $(BUILD_DIR)/"
	@ls -la $(BUILD_DIR)/

## Create release archives (manual packaging)
package: build-all
	@printf "$(GREEN)%s$(NC)\n" "Creating release packages..."
	@mkdir -p $(DIST_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		binary=$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then \
			binary=$$binary.exe; \
			zip -j $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$$os-$$arch.zip $$binary; \
		else \
			tar -czvf $(DIST_DIR)/$(BINARY_NAME)-$(VERSION)-$$os-$$arch.tar.gz -C $(BUILD_DIR) $$(basename $$binary); \
		fi; \
	done
	@printf "$(GREEN)%s$(NC)\n" "Packages created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

## Clean build artifacts
clean:
	@printf "$(YELLOW)%s$(NC)\n" "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(GO) clean
	@printf "$(GREEN)%s$(NC)\n" "Clean complete"

## Install binary to $GOPATH/bin
install: build
	@printf "$(GREEN)%s$(NC)\n" "Installing $(BINARY_NAME) to $(GOPATH)/bin/..."
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@printf "$(GREEN)%s$(NC)\n" "Installed: $(GOPATH)/bin/$(BINARY_NAME)"

## Uninstall binary from $GOPATH/bin
uninstall:
	@printf "$(YELLOW)%s$(NC)\n" "Uninstalling $(BINARY_NAME)..."
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@printf "$(GREEN)%s$(NC)\n" "Uninstalled"

## Build and run
run: build
	@$(BUILD_DIR)/$(BINARY_NAME)

## Run with specific repo
run-repo: build
	@$(BUILD_DIR)/$(BINARY_NAME) --repo $(REPO)

## Run all tests
test:
	@printf "$(GREEN)%s$(NC)\n" "Running tests..."
	$(GO) test -v -race -cover ./...

## Run tests with coverage report
test-coverage:
	@printf "$(GREEN)%s$(NC)\n" "Running tests with coverage..."
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@printf "$(GREEN)%s$(NC)\n" "Coverage report: coverage.html"

## Run linter (format + vet)
lint: fmt vet
	@printf "$(GREEN)%s$(NC)\n" "Lint complete"

## Format code
fmt:
	@printf "$(YELLOW)%s$(NC)\n" "Formatting code..."
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## Run go vet
vet:
	@printf "$(YELLOW)%s$(NC)\n" "Running go vet..."
	$(GO) vet ./...

## Tidy go modules
tidy:
	@printf "$(YELLOW)%s$(NC)\n" "Tidying go modules..."
	$(GO) mod tidy

## Verify dependencies
verify:
	@printf "$(YELLOW)%s$(NC)\n" "Verifying dependencies..."
	$(GO) mod verify

## Show version info
version:
	@printf "Version:    $(YELLOW)%s$(NC)\n" "$(VERSION)"
	@printf "Build Time: $(YELLOW)%s$(NC)\n" "$(BUILD_TIME)"
	@printf "Go Version: $(YELLOW)%s$(NC)\n" "$(GO_VERSION)"

## Show help
help:
	@printf "$(GREEN)%s$(NC)\n\n" "GitZen Makefile"
	@printf "Usage: make [target]\n\n"
	@printf "Targets:\n"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "build" "Build binary for current platform"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "build-all" "Build binaries for all platforms (cross-compile)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "package" "Create release archives (.tar.gz, .zip)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "clean" "Remove build artifacts"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "install" "Install binary to GOPATH/bin"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "uninstall" "Remove binary from GOPATH/bin"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "run" "Build and run"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "run-repo" "Run with specific repo (REPO=/path/to/repo)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "test" "Run all tests"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "test-coverage" "Run tests with coverage report"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "lint" "Run linter (fmt + vet)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "fmt" "Format code"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "vet" "Run go vet"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "tidy" "Tidy go modules"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "verify" "Verify dependencies"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "version" "Show version info"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "help" "Show this help"
	@printf "\nRelease targets:\n"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "release-patch" "Create patch release (bug fixes)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "release-minor" "Create minor release (new features)"
	@printf "  $(YELLOW)%-15s$(NC) %s\n" "release-major" "Create major release (breaking changes)"
	@printf "\nExamples:\n"
	@printf "  make build                  # Build for current OS\n"
	@printf "  make build-all              # Cross-compile for all platforms\n"
	@printf "  make package                # Create release archives\n"
	@printf "  make run-repo REPO=~/myrepo # Run with specific repo\n"
	@printf "  make release-patch          # Create patch release\n"

## Release targets
release-patch: ## Create a patch release (bug fixes)
	@./scripts/release.sh patch

release-minor: ## Create a minor release (new features)
	@./scripts/release.sh minor

release-major: ## Create a major release (breaking changes)  
	@./scripts/release.sh major

release: ## Interactive release (choose version)
	@./scripts/release.sh

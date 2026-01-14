# GitZen Makefile - Manual Build Automation
# Minh hoa quy trinh dong goi thu cong (Topic 3: Packaging & Deployment)

# Variables
BINARY_NAME := gitzen
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d_%H:%M:%S')
GO_VERSION := $(shell go version | cut -d ' ' -f 3)
LDFLAGS := -s -w -X main.version=$(VERSION) -X main.buildTime=$(BUILD_TIME)

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
	@echo "$(GREEN)Building $(BINARY_NAME) $(VERSION)...$(NC)"
	CGO_ENABLED=$(CGO_ENABLED) $(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $(BINARY_NAME) ./cmd/gitzen
	@echo "$(GREEN)Build complete: ./$(BINARY_NAME)$(NC)"

## Build binaries for all platforms (cross-compilation)
build-all: clean
	@echo "$(GREEN)Building $(BINARY_NAME) $(VERSION) for all platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@for platform in $(PLATFORMS); do \
		os=$$(echo $$platform | cut -d'/' -f1); \
		arch=$$(echo $$platform | cut -d'/' -f2); \
		output=$(BUILD_DIR)/$(BINARY_NAME)-$$os-$$arch; \
		if [ "$$os" = "windows" ]; then output=$$output.exe; fi; \
		echo "$(YELLOW)Building $$os/$$arch...$(NC)"; \
		GOOS=$$os GOARCH=$$arch CGO_ENABLED=$(CGO_ENABLED) \
			$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS)" -o $$output ./cmd/gitzen || exit 1; \
	done
	@echo "$(GREEN)All builds complete! Check $(BUILD_DIR)/$(NC)"
	@ls -la $(BUILD_DIR)/

## Create release archives (manual packaging)
package: build-all
	@echo "$(GREEN)Creating release packages...$(NC)"
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
	@echo "$(GREEN)Packages created in $(DIST_DIR)/$(NC)"
	@ls -la $(DIST_DIR)/

## Clean build artifacts
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -f $(BINARY_NAME)
	@rm -rf $(BUILD_DIR) $(DIST_DIR)
	@$(GO) clean
	@echo "$(GREEN)Clean complete$(NC)"

## Install binary to $GOPATH/bin
install: build
	@echo "$(GREEN)Installing $(BINARY_NAME) to $(GOPATH)/bin/...$(NC)"
	@cp $(BINARY_NAME) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)Installed: $(GOPATH)/bin/$(BINARY_NAME)$(NC)"

## Uninstall binary from $GOPATH/bin
uninstall:
	@echo "$(YELLOW)Uninstalling $(BINARY_NAME)...$(NC)"
	@rm -f $(GOPATH)/bin/$(BINARY_NAME)
	@echo "$(GREEN)Uninstalled$(NC)"

## Build and run
run: build
	@./$(BINARY_NAME)

## Run with specific repo
run-repo: build
	@./$(BINARY_NAME) --repo $(REPO)

## Run all tests
test:
	@echo "$(GREEN)Running tests...$(NC)"
	$(GO) test -v -race -cover ./...

## Run tests with coverage report
test-coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	$(GO) test -v -race -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report: coverage.html$(NC)"

## Run linter (format + vet)
lint: fmt vet
	@echo "$(GREEN)Lint complete$(NC)"

## Format code
fmt:
	@echo "$(YELLOW)Formatting code...$(NC)"
	$(GO) fmt ./...
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

## Run go vet
vet:
	@echo "$(YELLOW)Running go vet...$(NC)"
	$(GO) vet ./...

## Tidy go modules
tidy:
	@echo "$(YELLOW)Tidying go modules...$(NC)"
	$(GO) mod tidy

## Verify dependencies
verify:
	@echo "$(YELLOW)Verifying dependencies...$(NC)"
	$(GO) mod verify

## Show version info
version:
	@echo "Version:    $(VERSION)"
	@echo "Build Time: $(BUILD_TIME)"
	@echo "Go Version: $(GO_VERSION)"

## Show help
help:
	@echo "$(GREEN)GitZen Makefile$(NC)"
	@echo ""
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@echo "  $(YELLOW)build$(NC)          Build binary for current platform"
	@echo "  $(YELLOW)build-all$(NC)      Build binaries for all platforms (cross-compile)"
	@echo "  $(YELLOW)package$(NC)        Create release archives (.tar.gz, .zip)"
	@echo "  $(YELLOW)clean$(NC)          Remove build artifacts"
	@echo "  $(YELLOW)install$(NC)        Install binary to GOPATH/bin"
	@echo "  $(YELLOW)uninstall$(NC)      Remove binary from GOPATH/bin"
	@echo "  $(YELLOW)run$(NC)            Build and run"
	@echo "  $(YELLOW)run-repo$(NC)       Run with specific repo (REPO=/path/to/repo)"
	@echo "  $(YELLOW)test$(NC)           Run all tests"
	@echo "  $(YELLOW)test-coverage$(NC)  Run tests with coverage report"
	@echo "  $(YELLOW)lint$(NC)           Run linter (fmt + vet)"
	@echo "  $(YELLOW)fmt$(NC)            Format code"
	@echo "  $(YELLOW)vet$(NC)            Run go vet"
	@echo "  $(YELLOW)tidy$(NC)           Tidy go modules"
	@echo "  $(YELLOW)verify$(NC)         Verify dependencies"
	@echo "  $(YELLOW)version$(NC)        Show version info"
	@echo "  $(YELLOW)help$(NC)           Show this help"
	@echo ""
	@echo "Examples:"
	@echo "  make build                  # Build for current OS"
	@echo "  make build-all              # Cross-compile for all platforms"
	@echo "  make package                # Create release archives"
	@echo "  make run-repo REPO=~/myrepo # Run with specific repo"

.PHONY: build install install-user run clean deps test fmt lint build-all dev help version

# Binary name
BINARY_NAME=anime

# Version info
VERSION=$(shell cat VERSION 2>/dev/null || echo "1.0.0")
BUILD_TIME=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS=-ldflags "-X github.com/joshkornreich/anime/cmd.Version=$(VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$(shell pwd)"

# Detect OS
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Linux)
    OS=linux
endif
ifeq ($(UNAME_S),Darwin)
    OS=darwin
endif

# Detect architecture
UNAME_M := $(shell uname -m)
ifeq ($(UNAME_M),x86_64)
    ARCH=amd64
endif
ifeq ($(UNAME_M),arm64)
    ARCH=arm64
endif
ifeq ($(UNAME_M),aarch64)
    ARCH=arm64
endif

# Build directory
BUILD_DIR=build

# Default target
all: build

# Display version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit:  $(COMMIT)"
	@echo "OS:      $(OS)"
	@echo "Arch:    $(ARCH)"

# Create source tarball for embedding (only essential source files)
embed-source:
	@echo "Creating source tarball for embedding..."
	@tar -czf cmd/anime-src.tar.gz \
		--exclude='.git' \
		--exclude='build' \
		--exclude='*.tar.gz' \
		--exclude='.DS_Store' \
		--exclude='__pycache__' \
		--exclude='node_modules' \
		--exclude='B2002' \
		--exclude='animated' \
		--exclude='*.md' \
		--exclude='*.yml' \
		--exclude='setup-gh200.yml' \
		--exclude='*.ini' \
		--exclude='coverage.*' \
		--exclude='*/anime' \
		--exclude='anime-linux-*' \
		--exclude='anime-darwin-*' \
		--exclude='*.pem' \
		--exclude='*.key' \
		--exclude='.env' \
		--exclude='.env.*' \
		--exclude='*.bak' \
		--exclude='*.bak2' \
		--exclude='*.bak3' \
		--exclude='*.backup' \
		--exclude='*.orig' \
		-h \
		-C .. cli
	@echo "✓ Created cmd/anime-src.tar.gz ($$(du -h cmd/anime-src.tar.gz | cut -f1))"

# Pull agents and commands from ~/.claude before building
pull-claude:
	@echo "Pulling Claude agents and commands..."
	@if [ -d "$(HOME)/.claude/agents" ]; then \
		rm -rf cmd/embedded/agents/*.md 2>/dev/null || true; \
		cp -r $(HOME)/.claude/agents/*.md cmd/embedded/agents/ 2>/dev/null || echo "  No agents to pull"; \
		echo "✓ Pulled agents from ~/.claude/agents"; \
	else \
		echo "  No ~/.claude/agents directory found"; \
	fi
	@if [ -d "$(HOME)/.claude/commands" ]; then \
		rm -rf cmd/embedded/commands/*.md 2>/dev/null || true; \
		cp -r $(HOME)/.claude/commands/*.md cmd/embedded/commands/ 2>/dev/null || echo "  No commands to pull"; \
		echo "✓ Pulled commands from ~/.claude/commands"; \
	else \
		echo "  No ~/.claude/commands directory found"; \
	fi

# Build the anime binary for current platform
build: pull-claude embed-source
	@echo "Building $(BINARY_NAME) v$$(cat VERSION) for $(OS)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/$(BINARY_NAME) main.go
	@echo "✓ Built: $(BUILD_DIR)/$(BINARY_NAME) v$$(cat VERSION)"
	@echo ""
	@# Verify source embedding in built binary
	@echo "Verifying source embedding..."
	@$(BUILD_DIR)/$(BINARY_NAME) --verify-embed || (echo "✗ FATAL: Source embedding verification failed!" && exit 1)
	@echo "✓ Source embedding verified"
	@echo ""
	@echo "Installing to ~/go/bin and ~/.local/bin..."
	@go install -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)"
	@mkdir -p $(HOME)/.local/bin
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(HOME)/.local/bin/$(BINARY_NAME)
	@# Re-sign binaries on macOS to prevent SIGKILL
	@if [ "$(OS)" = "darwin" ]; then \
		codesign -s - -f $(HOME)/go/bin/$(BINARY_NAME) 2>/dev/null || true; \
		codesign -s - -f $(HOME)/.local/bin/$(BINARY_NAME) 2>/dev/null || true; \
	fi
	@echo "✓ Installed: ~/go/bin/$(BINARY_NAME) v$$(cat VERSION)"
	@echo "✓ Installed: ~/.local/bin/$(BINARY_NAME) v$$(cat VERSION)"

# Install anime to /usr/local/bin (requires sudo on some systems)
install: build
	@echo "Installing $(BINARY_NAME) to /usr/local/bin..."
	@if [ "$(OS)" = "darwin" ]; then \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	else \
		sudo install -m 755 $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/$(BINARY_NAME); \
	fi
	@echo "✓ $(BINARY_NAME) installed to /usr/local/bin/$(BINARY_NAME)"

# Install to user's go bin (no sudo required)
install-user: embed-source
	@echo "Installing $(BINARY_NAME) v$$(cat VERSION) to user Go bin..."
	go install -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)"
	@echo "✓ $(BINARY_NAME) installed to ~/go/bin/$(BINARY_NAME) v$$(cat VERSION)"
	@echo "  Make sure ~/go/bin is in your PATH"

# Run anime config
run:
	go run $(LDFLAGS) main.go config

# Run anime with arguments
run-args:
	go run $(LDFLAGS) main.go $(ARGS)

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf $(BUILD_DIR)
	rm -f $(BINARY_NAME)
	rm -f anime-*-*
	go clean
	@echo "✓ Clean complete"

# Download and tidy dependencies
deps:
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies updated"

# Run all tests
test:
	@echo "Running all tests..."
	go test -v -race -coverprofile=coverage.out ./...
	@echo ""
	@echo "Coverage summary:"
	@go tool cover -func=coverage.out | grep total:

# Run unit tests only (fast, no integration tests)
test-unit:
	@echo "Running unit tests..."
	go test -v -short -race -coverprofile=coverage-unit.out ./internal/...
	@echo ""
	@echo "Unit test coverage:"
	@go tool cover -func=coverage-unit.out | grep total:

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -run Integration ./...

# Verify build integrity (source embedding and build vars)
test-build: build
	@echo "Running build verification tests..."
	ANIME_TEST_BUILD_VERIFICATION=1 go test -v -run 'TestEmbedded|TestBuildDir|TestExtract' ./cmd/...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Coverage report: coverage.html"

# Generate coverage report for internal packages only
test-coverage-internal:
	@echo "Running tests for internal packages..."
	go test -v -coverprofile=coverage-internal.out ./internal/...
	@echo ""
	@echo "Generating coverage report for internal packages..."
	go tool cover -html=coverage-internal.out -o coverage-internal.html
	@echo "✓ Coverage report: coverage-internal.html"
	@echo ""
	@echo "Internal packages coverage:"
	@go tool cover -func=coverage-internal.out | grep total:

# Run tests with verbose output and show all failures
test-verbose:
	@echo "Running tests with verbose output..."
	go test -v -race -count=1 ./...

# Run a specific test by pattern
test-run:
	@if [ -z "$(TEST)" ]; then \
		echo "Usage: make test-run TEST=TestName"; \
		exit 1; \
	fi
	@echo "Running tests matching: $(TEST)"
	go test -v -run $(TEST) ./...

# Run benchmark tests
test-bench:
	@echo "Running benchmark tests..."
	go test -bench=. -benchmem ./...

# Format code
fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✓ Code formatted"

# Lint code (requires golangci-lint)
lint:
	@echo "Linting code..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi

# Vet code
vet:
	@echo "Vetting code..."
	go vet ./...
	@echo "✓ Vet complete"

# Check code quality (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "✓ All checks passed"

# Build for multiple platforms
build-all:
	@echo "Building for multiple platforms (v$$(cat VERSION))..."
	@mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=arm64 go build -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 main.go
	@echo "✓ Built binaries (v$$(cat VERSION)):"
	@ls -lh $(BUILD_DIR)/

# Build release binaries with optimizations
release: clean
	@echo "Building release binaries (v$$(cat VERSION))..."
	@mkdir -p $(BUILD_DIR)/release
	GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-arm64 main.go
	GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-darwin-amd64 main.go
	GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-arm64 main.go
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-X github.com/joshkornreich/anime/cmd.Version=$$(cat VERSION) -X github.com/joshkornreich/anime/cmd.BuildTime=$(BUILD_TIME) -X github.com/joshkornreich/anime/cmd.Commit=$(COMMIT) -X github.com/joshkornreich/anime/cmd.BuildDir=$$(pwd)" -o $(BUILD_DIR)/release/$(BINARY_NAME)-linux-amd64 main.go
	@cd $(BUILD_DIR)/release && sha256sum * > checksums.txt
	@echo "✓ Release binaries built in $(BUILD_DIR)/release/ (v$$(cat VERSION))"
	@ls -lh $(BUILD_DIR)/release/

# Quick TUI demo
demo:
	go run $(LDFLAGS) main.go config

# Development mode - watch and rebuild
dev:
	@echo "Development mode - watching for changes..."
	@which air > /dev/null || (echo "Installing air..." && go install github.com/air-verse/air@latest)
	air

# Install development tools
dev-tools:
	@echo "Installing development tools..."
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install github.com/air-verse/air@latest
	@echo "✓ Development tools installed"

# Help
help:
	@echo "anime - Lambda GH200 Management CLI"
	@echo ""
	@echo "Build targets:"
	@echo "  make build           - Build and install to ~/go/bin (bumps version)"
	@echo "  make build-all       - Build for all platforms (darwin/linux, amd64/arm64)"
	@echo "  make release         - Build optimized release binaries with checksums"
	@echo "  make clean           - Remove build artifacts"
	@echo ""
	@echo "Installation:"
	@echo "  make install         - Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user    - Same as 'make build' (bumps version, installs to ~/go/bin)"
	@echo ""
	@echo "Development:"
	@echo "  make run             - Run anime config"
	@echo "  make run-args ARGS='status' - Run anime with arguments"
	@echo "  make demo            - Quick TUI demo"
	@echo "  make dev             - Watch mode with auto-reload (requires air)"
	@echo "  make dev-tools       - Install development tools"
	@echo ""
	@echo "Code quality:"
	@echo "  make deps                   - Download and tidy dependencies"
	@echo "  make test                   - Run all tests with race detection"
	@echo "  make test-unit              - Run unit tests only (fast)"
	@echo "  make test-integration       - Run integration tests only"
	@echo "  make test-build             - Verify source embedding and build integrity"
	@echo "  make test-coverage          - Run tests with coverage report (60% threshold)"
	@echo "  make test-coverage-internal - Coverage report for internal packages only"
	@echo "  make test-verbose           - Run tests with verbose output"
	@echo "  make test-run TEST=Name     - Run specific test by pattern"
	@echo "  make test-bench             - Run benchmark tests"
	@echo "  make fmt                    - Format code"
	@echo "  make vet                    - Run go vet"
	@echo "  make lint                   - Run golangci-lint (requires golangci-lint)"
	@echo "  make check                  - Run all checks (fmt, vet, lint, test)"
	@echo ""
	@echo "Other:"
	@echo "  make version         - Display version information"
	@echo "  make help            - Show this help message"
	@echo ""
	@echo "Current platform: $(OS)/$(ARCH)"
	@echo "Version: $(VERSION)"

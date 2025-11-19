.PHONY: build install run clean deps test

# Build the anime binary
build:
	go build -o anime main.go

# Install anime to /usr/local/bin
install: build
	sudo mv anime /usr/local/bin/anime
	@echo "✓ anime installed to /usr/local/bin/anime"

# Install to user's go bin
install-user:
	go install

# Run anime config
run:
	go run main.go config

# Run anime with arguments
run-args:
	go run main.go $(ARGS)

# Clean build artifacts
clean:
	rm -f anime
	go clean

# Download dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test -v ./...

# Format code
fmt:
	go fmt ./...

# Lint code
lint:
	golangci-lint run

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=arm64 go build -o anime-darwin-arm64
	GOOS=darwin GOARCH=amd64 go build -o anime-darwin-amd64
	GOOS=linux GOARCH=arm64 go build -o anime-linux-arm64
	GOOS=linux GOARCH=amd64 go build -o anime-linux-amd64

# Quick test of TUI
demo:
	go run main.go config

# Help
help:
	@echo "anime - Lambda GH200 Management CLI"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build anime binary"
	@echo "  make install     - Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user - Install to ~/go/bin"
	@echo "  make run         - Run anime config"
	@echo "  make clean       - Remove build artifacts"
	@echo "  make deps        - Download dependencies"
	@echo "  make test        - Run tests"
	@echo "  make fmt         - Format code"
	@echo "  make demo        - Quick TUI demo"
	@echo ""
	@echo "Usage after install:"
	@echo "  anime config     - Configure servers"
	@echo "  anime deploy SERVER - Deploy to server"
	@echo "  anime status SERVER - Check server status"
	@echo "  anime list       - List all servers"

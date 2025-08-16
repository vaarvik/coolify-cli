.PHONY: build clean install install-local test release dev-config dev-test run-logs

# Version information
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

# Build flags
LDFLAGS = -ldflags "-X main.Version=$(VERSION) -X main.Commit=$(COMMIT) -X main.BuildTime=$(BUILD_TIME)"

# Installation paths
INSTALL_DIR = /usr/local/bin
CONFIG_DIR = $(HOME)/.coolify-cli
BINARY_NAME = coolify-cli

# Build the CLI binary
build:
	go build $(LDFLAGS) -o $(BINARY_NAME) .

# Build for multiple platforms
build-all: clean
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BINARY_NAME)-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME) $(BINARY_NAME)-*

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Install the CLI system-wide
install: build
	@echo "Installing $(BINARY_NAME) to $(INSTALL_DIR)..."
	@if [ -w "$(INSTALL_DIR)" ]; then \
		cp $(BINARY_NAME) $(INSTALL_DIR)/; \
	else \
		sudo cp $(BINARY_NAME) $(INSTALL_DIR)/; \
	fi
	@echo "Creating config directory at $(CONFIG_DIR)..."
	@mkdir -p $(CONFIG_DIR)
	@if [ ! -f "$(CONFIG_DIR)/config.json" ]; then \
		$(INSTALL_DIR)/$(BINARY_NAME) config init; \
	fi
	@echo "Installation complete! Run '$(BINARY_NAME) --help' to get started."

# Install locally for development
install-local: build
	@echo "Installing $(BINARY_NAME) to GOPATH/bin..."
	go install .

# Create a release
release: build-all
	@echo "Creating release artifacts..."
	@mkdir -p dist
	@cp $(BINARY_NAME)-* dist/
	@cd dist && \
		shasum -a 256 * > checksums.txt && \
		for file in $(BINARY_NAME)-* ; do \
			tar czf $$file.tar.gz $$file ; \
			rm $$file ; \
		done
	@echo "Release artifacts created in dist/"

# Development helpers
dev-config:
	./$(BINARY_NAME) config init

dev-test:
	./$(BINARY_NAME) config test

# Example: run logs command
run-logs:
	./$(BINARY_NAME) logs nk4kcskcsswg0wskk88skcsg

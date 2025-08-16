.PHONY: build clean install test run-logs

# Build the CLI binary
build:
	go build -o coolify-cli .

# Build for multiple platforms
build-all:
	GOOS=linux GOARCH=amd64 go build -o coolify-cli-linux-amd64 .
	GOOS=darwin GOARCH=amd64 go build -o coolify-cli-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 go build -o coolify-cli-darwin-arm64 .
	GOOS=windows GOARCH=amd64 go build -o coolify-cli-windows-amd64.exe .

# Clean build artifacts
clean:
	rm -f coolify-cli coolify-cli-*

# Install dependencies
deps:
	go mod download
	go mod tidy

# Run tests
test:
	go test ./...

# Install the CLI to GOPATH/bin
install:
	go install .

# Development helpers
dev-config:
	./coolify-cli config init

dev-test:
	./coolify-cli config test

# Example: run logs command (replace with actual app ID)
run-logs:
	./coolify-cli logs nk4kcskcsswg0wskk88skcsg

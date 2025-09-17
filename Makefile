# Makefile for crgodicom
BINARY_NAME=crgodicom
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -s -w"

.PHONY: build build-all clean test install deps docker-test-setup

# Build for current platform
build:
	@echo "Building ${BINARY_NAME} for current platform..."
	@mkdir -p bin
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/crgodicom/main.go

# Build for all target platforms
build-all:
	@echo "Building ${BINARY_NAME} for all platforms..."
	@mkdir -p bin
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe cmd/crgodicom/main.go
	GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-arm64.exe cmd/crgodicom/main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/crgodicom/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 cmd/crgodicom/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/crgodicom/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 cmd/crgodicom/main.go
	@echo "Build complete. Binaries available in bin/"

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	go mod download
	go mod tidy

# Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -f coverage.out coverage.html

# Install binary to GOPATH/bin
install: build
	@echo "Installing ${BINARY_NAME} to GOPATH/bin..."
	go install ${LDFLAGS} cmd/crgodicom/main.go

# Setup Docker test PACS servers
docker-test-setup:
	@echo "Setting up Docker Orthanc PACS test servers..."
	docker run -d --name orthanc-test-1 -p 4900:4242 -p 4901:4243 orthancteam/orthanc:latest
	docker run -d --name orthanc-test-2 -p 4902:4242 -p 4903:4243 orthancteam/orthanc:latest
	@echo "Test PACS servers started:"
	@echo "  Orthanc 1: localhost:4900 (Web UI: localhost:4901)"
	@echo "  Orthanc 2: localhost:4902 (Web UI: localhost:4903)"

# Stop Docker test PACS servers
docker-test-stop:
	@echo "Stopping Docker test PACS servers..."
	docker stop orthanc-test-1 orthanc-test-2 || true
	docker rm orthanc-test-1 orthanc-test-2 || true

# Development setup
dev-setup: deps
	@echo "Development setup complete"
	@echo "Run 'make docker-test-setup' to start test PACS servers"

# Show help
help:
	@echo "Available targets:"
	@echo "  build          - Build for current platform"
	@echo "  build-all      - Build for all target platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  deps           - Install dependencies"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install binary to GOPATH/bin"
	@echo "  docker-test-setup - Start Docker test PACS servers"
	@echo "  docker-test-stop - Stop Docker test PACS servers"
	@echo "  dev-setup      - Setup development environment"
	@echo "  help           - Show this help message"

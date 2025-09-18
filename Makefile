# Makefile for crgodicom
BINARY_NAME=crgodicom
VERSION?=$(shell git describe --tags --always --dirty)
BUILD_DATE?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
GIT_COMMIT?=$(shell git rev-parse HEAD)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.BuildDate=${BUILD_DATE} -X main.GitCommit=${GIT_COMMIT} -s -w"

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

# Create installers for all platforms
installer-all:
	@echo "Creating installers for all platforms..."
	@mkdir -p dist
	@$(MAKE) installer-windows
	@$(MAKE) installer-macos
	@$(MAKE) installer-linux

# Create Windows installer
installer-windows:
	@echo "Creating Windows installer..."
	@mkdir -p dist/windows
	@$(MAKE) build-all
	@cp bin/${BINARY_NAME}-windows-*.exe dist/windows/
	@cp README.md CHANGELOG.md LICENSE crgodicom.yaml dist/windows/
	@cp -r examples dist/windows/
	@echo "Windows installer files prepared in dist/windows/"

# Create macOS installer
installer-macos:
	@echo "Creating macOS installer..."
	@mkdir -p dist/macos
	@$(MAKE) build-all
	@cp bin/${BINARY_NAME}-darwin-* dist/macos/
	@cp README.md CHANGELOG.md LICENSE crgodicom.yaml dist/macos/
	@cp -r examples dist/macos/
	@echo "macOS installer files prepared in dist/macos/"

# Create Linux installer
installer-linux:
	@echo "Creating Linux installer..."
	@mkdir -p dist/linux
	@$(MAKE) build-all
	@cp bin/${BINARY_NAME}-linux-* dist/linux/
	@cp README.md CHANGELOG.md LICENSE crgodicom.yaml dist/linux/
	@cp -r examples dist/linux/
	@echo "Linux installer files prepared in dist/linux/"

# Create AppImage for Linux
appimage:
	@echo "Creating Linux AppImage..."
	@mkdir -p dist/appimage
	@$(MAKE) build-all
	@scripts/create-appimage.sh
	@echo "AppImage created in dist/appimage/"

# Create DMG for macOS
dmg:
	@echo "Creating macOS DMG..."
	@mkdir -p dist/dmg
	@$(MAKE) build-all
	@scripts/create-dmg.sh
	@echo "DMG created in dist/dmg/"

# Create MSI for Windows
msi:
	@echo "Creating Windows MSI..."
	@mkdir -p dist/msi
	@$(MAKE) build-all
	@scripts/create-msi.sh
	@echo "MSI created in dist/msi/"

# Clean installer artifacts
clean-installers:
	@echo "Cleaning installer artifacts..."
	rm -rf dist/

# Show help
help:
	@echo "Available targets:"
	@echo "  build              - Build for current platform"
	@echo "  build-all          - Build for all target platforms"
	@echo "  test               - Run tests"
	@echo "  test-coverage      - Run tests with coverage report"
	@echo "  deps               - Install dependencies"
	@echo "  clean              - Clean build artifacts"
	@echo "  clean-installers   - Clean installer artifacts"
	@echo "  install            - Install binary to GOPATH/bin"
	@echo "  installer-all      - Create installers for all platforms"
	@echo "  installer-windows  - Create Windows installer"
	@echo "  installer-macos    - Create macOS installer"
	@echo "  installer-linux    - Create Linux installer"
	@echo "  appimage           - Create Linux AppImage"
	@echo "  dmg                - Create macOS DMG"
	@echo "  msi                - Create Windows MSI"
	@echo "  docker-test-setup  - Start Docker test PACS servers"
	@echo "  docker-test-stop   - Stop Docker test PACS servers"
	@echo "  dev-setup          - Setup development environment"
	@echo "  help               - Show this help message"

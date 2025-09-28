# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

CRGoDICOM is a cross-platform CLI utility written in Go for creating synthetic DICOM data and sending it to PACS systems. It's a Go implementation of the Python dicom-maker with improved performance and easier distribution.

**Key Features:**
- Synthetic DICOM data generation with configurable fields
- PACS integration via DCMTK (C-ECHO, C-STORE operations)
- Study templates for common modalities
- Export capabilities (PNG, JPEG, PDF)
- Cross-platform single binary distribution

## Development Commands

### Build and Test
```bash
# Build for current platform
make build

# Build for all platforms (Windows, macOS, Linux - amd64/arm64)
make build-all

# Run tests with race detection and coverage
make test

# Generate coverage report
make test-coverage

# Install dependencies and tidy modules
make deps

# Clean build artifacts
make clean
```

### Development Setup
```bash
# Setup development environment
make dev-setup

# Start Docker PACS test servers
make docker-test-setup

# Stop Docker PACS test servers
make docker-test-stop
```

### Testing Individual Components
```bash
# Run specific package tests
go test -v ./internal/dicom/...
go test -v ./internal/cli/...
go test -v ./internal/orm/parser/...

# Run single test function
go test -v ./internal/cli/ -run TestCreateCommand

# Run tests with specific flags
go test -race -coverprofile=coverage.out ./...
```

## Architecture Overview

### Package Structure
```
cmd/crgodicom/          # Main CLI application entry point
internal/
├── cli/                # CLI commands and handlers (create, list, send, etc.)
├── config/             # YAML configuration management and validation
├── dcmtk/              # DCMTK integration layer (network, reader, writer)
├── dicom/              # Core DICOM data generation and manipulation
├── export/             # Export functionality (PNG, JPEG, PDF)
├── image/              # Image generation with noise patterns
├── orm/                # HL7 ORM message parsing and processing
│   ├── parser/         # HL7 message parsing logic
│   ├── generator/      # DICOM generation from ORM data
│   └── config/         # ORM-specific configuration
├── pacs/               # PACS client operations (C-FIND, C-MOVE)
└── storage/            # Local storage management for studies
```

### Core Components

**CLI Framework:** Uses urfave/cli/v2 with command pattern. Each command is defined in `internal/cli/` with corresponding handler functions.

**Configuration:** YAML-based config system (`internal/config/`) with CLI flag overrides. Default config in `crgodicom.yaml`, extended config in `crgodicom.example.yaml`.

**DICOM Generation:** Core DICOM creation in `internal/dicom/` using suyashkumar/dicom library. Templates defined in config system.

**DCMTK Integration:** External DCMTK toolkit integration in `internal/dcmtk/` for PACS operations. Uses system calls to DCMTK binaries (storescu, echoscu, findscu).

**Study Templates:** Built-in templates for common modalities (chest-xray, ct-chest, mri-brain, etc.) with configurable series/image counts.

## DCMTK Integration Patterns

The application uses DCMTK (DICOM Toolkit) for PACS communication:

- **Manager Pattern:** `dcmtk.NewManager()` handles DCMTK installation detection and tool path resolution
- **System Integration:** Calls external DCMTK binaries via exec commands rather than CGO bindings
- **Tool Detection:** Automatic detection of storescu, echoscu, findscu, movescu tools
- **Installation Helper:** Built-in check-dcmtk command provides platform-specific installation instructions

## Configuration System

The app uses a hierarchical configuration system:
1. **Default Configuration:** Hard-coded defaults in `internal/config/config.go`
2. **YAML Configuration:** `crgodicom.yaml` in working directory
3. **CLI Flag Overrides:** Command-line flags take highest precedence

### Configuration Loading Pattern
```go
cfg, err := config.LoadConfig(configPath)
if err != nil {
    cfg = config.DefaultConfig() // Fallback to defaults
}
// Override with CLI flags
```

## Testing Strategy

- **Unit Tests:** `*_test.go` files alongside source code
- **Integration Tests:** Docker-based PACS servers for network testing
- **CLI Testing:** Command execution testing in `internal/cli/*_test.go`
- **Coverage:** Use `make test-coverage` for HTML coverage reports

### Running Tests
```bash
# All tests
go test -v -race ./...

# Specific package
go test -v ./internal/dicom/

# With coverage
go test -race -coverprofile=coverage.out ./... && go tool cover -html=coverage.out
```

## Development Workflow

### Adding New Commands
1. Create command handler in `internal/cli/`
2. Add command to `cmd/crgodicom/main.go` commands slice
3. Add tests in `internal/cli/*_test.go`
4. Update configuration schema if needed

### Adding New Templates
1. Add template to built-in templates in `internal/config/config.go`
2. Or add to `crgodicom.yaml` for user-defined templates
3. Test with `crgodicom create --template your-template`

### DCMTK Integration
- Use `dcmtk.NewManager()` for DCMTK operations
- Always check availability with `manager.CheckAvailability()`
- Use `internal/cli/check-dcmtk.go` patterns for tool detection
- Handle missing DCMTK gracefully with helpful error messages

## Important Files

- **`Makefile`:** Complete build, test, and development automation
- **`crgodicom.yaml`:** Default configuration with built-in templates
- **`crgodicom.example.yaml`:** Extended configuration example
- **`cmd/crgodicom/main.go`:** CLI application entry point with command registration
- **`internal/config/config.go`:** Configuration structs and validation logic
- **`internal/dcmtk/dcmtk_*.go`:** DCMTK integration and network operations

## Common Development Tasks

### Add New DICOM Modality
1. Add template to `getBuiltInTemplates()` in `internal/config/config.go`
2. Test generation: `make build && ./bin/crgodicom create --template your-modality`

### Debug PACS Operations
1. Use `crgodicom check-dcmtk --verbose` to verify DCMTK setup
2. Start test PACS: `make docker-test-setup`
3. Test connectivity: `crgodicom echo --host localhost --port 4900`

### Add Export Format
1. Implement exporter in `internal/export/`
2. Add format to export command flags in `internal/cli/export.go`
3. Add tests in `internal/export/*_test.go`
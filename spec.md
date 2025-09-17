# CRGoDICOM Gen Utility Specification

**Project Name**: `crgodicom`  
**Version**: 1.0.0  
**Target Platforms**: Windows, macOS, Linux (amd64, arm64)  
**License**: MIT  
**Copyright**: flatmapit.com  
**Python Reference**: https://github.com/flatmapit/dicom-maker  

## Overview

This specification describes a cross-platform, single-binary CLI utility written in Go that provides identical functionality to the existing Python `dicom-maker` application. The Go implementation should offer improved performance, easier distribution, and zero external dependencies while maintaining complete feature parity.

## Core Requirements

### 1. Cross-Platform Compatibility
- **Target OS**: Windows, macOS, Linux
- **Target Architectures**: amd64, arm64
- **Single Binary**: No external dependencies or runtime requirements
- **Identical CLI Interface**: Command syntax and options must match Python version
- **Native Performance**: Leverage Go's compilation advantages

### 2. Distribution & Installation
- **GitHub Releases**: Automated builds for all platforms
- **Package Managers**: Homebrew (macOS), Chocolatey (Windows), APT/YUM (Linux)
- **Direct Download**: Pre-compiled binaries with checksums
- **Docker Image**: Multi-arch container support
- **Installation Script**: One-liner installation for all platforms

### 3. Build System
- **Go Modules**: Modern dependency management
- **Cross Compilation**: `GOOS`/`GOARCH` matrix builds
- **CI/CD**: GitHub Actions for automated builds and releases
- **Version Embedding**: Git tag/commit information in binary
- **Size Optimization**: UPX compression for smaller binaries

## Functional Requirements

### CLI Commands Structure

The Go utility must implement these exact commands and options:

#### Global Options
```bash
crgodicom [GLOBAL OPTIONS] COMMAND [COMMAND OPTIONS] [ARGUMENTS...]

GLOBAL OPTIONS:
   --config FILE         Configuration file path (default: crgodicom.yaml in cwd)
   --log-file FILE       Log file path
   --log-level LEVEL     Log level (DEBUG, INFO, WARNING, ERROR, CRITICAL)
   --version, -v         Show version information
   --help, -h            Show help
```

#### Command: `create`
```bash
crgodicom create [OPTIONS]

Create synthetic DICOM studies

OPTIONS:
   --study-count COUNT          Number of studies to create (default: 1)
   --series-count COUNT         Number of series per study (default: 1)
   --image-count COUNT          Number of images per series (default: 1)
   --modality MODALITY          DICOM modality: CR, CT, MR, US, DX, MG (default: CR)
   --template TEMPLATE          Study template name
   --anatomical-region REGION   Anatomical region (default: chest)
   --patient-id ID              Patient ID
   --patient-name NAME          Patient name (format: LAST^FIRST^MIDDLE)
   --accession-number NUMBER    Accession number
   --study-description DESC     Study description
   --output-dir DIR             Output directory (default: studies)
   --config FILE                Configuration file path
```

#### Command: `list`
```bash
crgodicom list [OPTIONS]

List local DICOM studies

OPTIONS:
   --output-dir DIR    Studies directory (default: studies)
   --format FORMAT     Output format: table, json, csv (default: table)
   --verbose, -v       Show detailed information
```

#### Command: `send`
```bash
crgodicom send [OPTIONS]

Send DICOM study to PACS

OPTIONS:
   --study-id UID      Study Instance UID (required)
   --host HOST         PACS host address (required)
   --port PORT         PACS port (default: 11112)
   --aec AEC           Application Entity Caller (required)
   --aet AET           Application Entity Title (required)
   --output-dir DIR    Studies directory (default: studies)
   --timeout SECONDS   Connection timeout (default: 30)
   --retries COUNT     Retry attempts (default: 3)
```

#### Command: `verify`
```bash
crgodicom verify [OPTIONS]

Verify PACS connection using C-ECHO

OPTIONS:
   --host HOST         PACS host address (required)
   --port PORT         PACS port (default: 11112)
   --aec AEC           Application Entity Caller (required)
   --aet AET           Application Entity Title (required)
   --timeout SECONDS   Connection timeout (default: 10)
```

#### Command: `export`
```bash
crgodicom export [OPTIONS]

Export DICOM study to various formats

OPTIONS:
   --study-id UID         Study Instance UID (required)
   --format FORMAT        Export format: png, pdf (required)
   --output-dir DIR       Output directory (for PNG format)
   --output-file FILE     Output file path (for PDF format)
   --input-dir DIR        Studies directory (default: studies)
   --include-metadata     Include metadata files (PNG format)
```

#### Command: `query` (Future Enhancement)
```bash
crgodicom query [OPTIONS]

Query PACS for studies using C-FIND

OPTIONS:
   --host HOST           PACS host address (required)
   --port PORT           PACS port (default: 11112)
   --aec AEC             Application Entity Caller (required)
   --aet AET             Application Entity Title (required)
   --patient-id ID       Patient ID filter
   --patient-name NAME   Patient name filter
   --study-date DATE     Study date filter (YYYYMMDD)
   --modality MODALITY   Modality filter
   --accession NUMBER    Accession number filter
   --format FORMAT       Output format: table, json (default: table)
```

### Core Functionality

#### 1. DICOM Data Generation
- **Realistic Synthetic Data**: Generate anatomically plausible pixel data
- **DICOM 3.0 Compliance**: Full standard compliance with proper UIDs
- **Modality Support**: CR, CT, MR, US, DX, MG with appropriate SOP Classes
- **Burnt-in Text**: Metadata overlay on generated images
- **Field Validation**: Automatic generation of mandatory DICOM fields
- **Template System**: Predefined study templates for common examinations

**Required DICOM Fields Generation**:
```go
// Patient Module
PatientName         // User-provided or "DOE^JOHN^M"
PatientID          // User-provided or random 8-char hex
PatientBirthDate   // Random date (18-80 years old)

// Study Module  
StudyInstanceUID   // Generated using configurable org.root + timestamp + random
StudyDate          // Current date (YYYYMMDD)
StudyTime          // Current time (HHMMSS)
AccessionNumber    // User-provided or "YYYYMMDD-XXXX" format

// Series Module
SeriesInstanceUID  // Generated UID using configurable org.root
SeriesNumber       // Sequential (1, 2, 3...)
Modality          // User-specified or template default

// Image Module
SOPInstanceUID     // Generated UID per image using configurable org.root
SOPClassUID        // Based on modality
InstanceNumber     // Sequential (1, 2, 3...)
```

#### 2. Image Generation
- **Abstract Noise Patterns**: Generate noise-based synthetic medical images (no anatomical simulation required)
- **Modality-Specific**: Different noise characteristics per modality based on Python dicom-maker examples
- **Configurable Dimensions**: Support standard DICOM image sizes
- **Pixel Data Types**: 8-bit and 16-bit grayscale support
- **Burnt-in Metadata**: Overlay patient/study information on images

**Image Specifications**:
```go
// Standard image dimensions by modality
CR: 2048x2048, 16-bit  // Chest X-Ray
CT: 512x512, 16-bit    // CT slices  
MR: 256x256, 16-bit    // MR images
US: 640x480, 8-bit     // Ultrasound
DX: 2048x2048, 16-bit  // Digital X-Ray
MG: 4096x3328, 16-bit  // Mammography
```

#### 3. PACS Communication
- **DICOM Network Services**: C-ECHO, C-STORE, C-FIND support
- **Association Management**: Proper DICOM association handling
- **Transfer Syntax**: Implicit VR Little Endian support
- **Error Handling**: Robust network error handling and retries
- **Connection Pooling**: Efficient connection reuse

**Supported SOP Classes**:
```go
// Verification
"1.2.840.10008.1.1" // Verification SOP Class

// Image Storage  
"1.2.840.10008.5.1.4.1.1.1"   // Computed Radiography
"1.2.840.10008.5.1.4.1.1.2"   // CT Image Storage
"1.2.840.10008.5.1.4.1.1.4"   // MR Image Storage  
"1.2.840.10008.5.1.4.1.1.6"   // Ultrasound Image Storage
"1.2.840.10008.5.1.4.1.1.1.1" // Digital X-Ray Image Storage
"1.2.840.10008.5.1.4.1.1.1.2" // Digital Mammography X-Ray Image Storage
```

#### 4. Export Functionality
- **PNG Export**: Individual images with metadata text files
- **PDF Export**: Multi-page reports with images and metadata
- **Metadata Extraction**: Comprehensive DICOM tag extraction
- **File Organization**: Structured output directory layout

**PNG Export Structure**:
```
output_dir/
├── study_metadata.json
├── series_001/
│   ├── image_001.png
│   ├── image_001.txt
│   ├── image_002.png
│   └── image_002.txt
└── series_002/
    ├── image_001.png
    ├── image_001.txt
    └── ...
```

**PDF Export Features**:
- Study summary page with patient/study metadata
- One image per page with individual metadata
- Series Instance UIDs included
- Professional medical report formatting

#### 5. Study Management
- **Local Storage**: Organized DICOM file storage
- **Study Indexing**: Fast study lookup and listing
- **Metadata Caching**: Efficient study information retrieval
- **File Validation**: DICOM file integrity checking

**Storage Structure**:
```
studies/
└── {StudyInstanceUID}/
    ├── study_metadata.json
    └── series_001/
        ├── image_001.dcm
        ├── image_002.dcm
        └── ...
```

### Go-Specific Implementation Requirements

#### 1. Dependencies
```go
// Core DICOM library
github.com/suyashkumar/dicom v1.0.5

// CLI framework  
github.com/urfave/cli/v2 v2.25.7

// Image processing
github.com/fogleman/gg v1.3.0
golang.org/x/image v0.10.0

// PDF generation
github.com/jung-kurt/gofpdf v1.16.2

// Logging
github.com/sirupsen/logrus v1.9.3

// Configuration
gopkg.in/yaml.v3 v3.0.1

// Testing
github.com/stretchr/testify v1.8.4
```

#### 2. Project Structure
```
crgodicom/
├── cmd/
│   └── crgodicom/
│       └── main.go
├── internal/
│   ├── cli/
│   │   ├── create.go
│   │   ├── list.go
│   │   ├── send.go
│   │   ├── verify.go
│   │   ├── export.go
│   │   └── query.go
│   ├── dicom/
│   │   ├── generator.go
│   │   ├── validator.go
│   │   └── writer.go
│   ├── pacs/
│   │   ├── client.go
│   │   ├── echo.go
│   │   ├── store.go
│   │   └── find.go
│   ├── image/
│   │   ├── generator.go
│   │   └── overlay.go
│   ├── export/
│   │   ├── png.go
│   │   └── pdf.go
│   ├── storage/
│   │   ├── manager.go
│   │   └── index.go
│   └── config/
│       └── config.go
├── pkg/
│   └── types/
│       └── dicom.go
├── test/
├── scripts/
│   ├── build.sh
│   ├── release.sh
│   └── install.sh
├── .github/
│   └── workflows/
│       ├── build.yml
│       ├── release.yml
│       └── test.yml
├── go.mod
├── go.sum
├── README.md
├── LICENSE
└── Makefile
```

#### 3. Build Configuration
```makefile
# Makefile
BINARY_NAME=crgodicom
VERSION?=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-X main.Version=${VERSION} -s -w"

.PHONY: build build-all clean test

build:
	go build ${LDFLAGS} -o bin/${BINARY_NAME} cmd/crgodicom/main.go

build-all:
	GOOS=windows GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-amd64.exe cmd/crgodicom/main.go
	GOOS=windows GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-windows-arm64.exe cmd/crgodicom/main.go
	GOOS=darwin GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-amd64 cmd/crgodicom/main.go
	GOOS=darwin GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-darwin-arm64 cmd/crgodicom/main.go
	GOOS=linux GOARCH=amd64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-amd64 cmd/crgodicom/main.go
	GOOS=linux GOARCH=arm64 go build ${LDFLAGS} -o bin/${BINARY_NAME}-linux-arm64 cmd/crgodicom/main.go

test:
	go test -v -race -coverprofile=coverage.out ./...

clean:
	rm -rf bin/
```

#### 4. GitHub Actions CI/CD
```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        goos: [linux, windows, darwin]
        goarch: [amd64, arm64]
    
    steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    
    - name: Build
      env:
        GOOS: ${{ matrix.goos }}
        GOARCH: ${{ matrix.goarch }}
      run: |
        make build-all
    
    - name: Create Release
      uses: softprops/action-gh-release@v1
      with:
        files: bin/*
```

### Configuration System

#### Configuration File Format
```yaml
# crgodicom.yaml (default config file in current working directory)
# CLI flags override configuration file values

dicom:
  org_root: "1.2.840.10008.5.1.4.1.1"  # Configurable UID organization root

default_pacs:
  host: "localhost"
  port: 11112
  aec: "CRGODICOM"
  aet: "PACS_SERVER"
  timeout: 30

study_templates:
  # Built-in templates (always available)
  chest-xray:
    modality: "CR"
    series_count: 1
    image_count: 2
    anatomical_region: "chest"
    study_description: "Chest X-Ray"
  
  ct-chest:
    modality: "CT"
    series_count: 2
    image_count: 50
    anatomical_region: "chest"
    study_description: "CT Chest"
  
  ultrasound-abdomen:
    modality: "US"
    series_count: 1
    image_count: 10
    anatomical_region: "abdomen"
    study_description: "Ultrasound Abdomen"

  # User-defined templates can be added here
  # Use 'crgodicom create-template' to generate template structure

logging:
  level: "INFO"
  file: "crgodicom.log"
  format: "json"

storage:
  base_dir: "studies"
  compression: false
  index_cache: true

# Test PACS configuration for development/testing
test_pacs:
  orthanc_1:
    host: "localhost"
    port: 4900
    aec: "CRGODICOM"
    aet: "ORTHANC1"
    docker_image: "orthancteam/orthanc"
  
  orthanc_2:
    host: "localhost"
    port: 4901
    aec: "CRGODICOM"
    aet: "ORTHANC2"
    docker_image: "orthancteam/orthanc"
```

### Error Handling & Logging

#### Logging Requirements
- **Structured Logging**: JSON format for machine parsing
- **Log Levels**: DEBUG, INFO, WARNING, ERROR, CRITICAL
- **Context**: Include operation context and timing
- **File Rotation**: Automatic log file rotation
- **Console Output**: Colored, human-readable console output

#### Error Categories
```go
// Error types that must be handled
type ErrorCategory int

const (
    NetworkError ErrorCategory = iota  // PACS connection issues
    ValidationError                    // Invalid DICOM data
    FileSystemError                   // File I/O problems
    ConfigurationError               // Invalid configuration
    GenerationError                  // DICOM generation failures
    ExportError                     // Export operation failures
)
```

### Testing Requirements

#### Test Coverage
- **Unit Tests**: 90%+ coverage for all packages
- **Integration Tests**: End-to-end CLI command testing
- **PACS Tests**: Mock PACS server for network testing
- **Cross-Platform Tests**: Automated testing on all target platforms
- **Performance Tests**: Benchmarks for large study generation

#### Test Structure
```
test/
├── unit/
│   ├── dicom_test.go
│   ├── pacs_test.go
│   └── export_test.go
├── integration/
│   ├── cli_test.go
│   └── e2e_test.go
├── fixtures/
│   ├── sample_dicoms/
│   └── test_configs/
└── mocks/
    ├── pacs_server.go
    └── file_system.go
```

### Performance Requirements

#### Benchmarks
- **Study Generation**: 1000 images in <30 seconds
- **PACS Transmission**: 100MB study in <60 seconds
- **PDF Export**: 50-image study in <10 seconds
- **Memory Usage**: <500MB for 1000-image study
- **Binary Size**: <50MB after compression

#### Optimization Strategies
- **Concurrent Processing**: Parallel image generation
- **Memory Pooling**: Reuse image buffers
- **Streaming**: Stream large datasets to avoid memory spikes
- **Compression**: UPX binary compression
- **Caching**: Cache frequently accessed data

### Security Considerations

#### Data Protection
- **No Sensitive Data**: All generated data is synthetic
- **Secure Defaults**: Safe default configuration values
- **Input Validation**: Sanitize all user inputs
- **Network Security**: Secure DICOM network communication
- **File Permissions**: Appropriate file system permissions

### Documentation Requirements

#### User Documentation
- **Installation Guide**: Platform-specific installation instructions
- **User Manual**: Complete command reference with examples
- **Configuration Guide**: Configuration file documentation
- **Troubleshooting**: Common issues and solutions
- **API Documentation**: Go package documentation

#### Developer Documentation
- **Architecture**: System design and component interaction
- **Contributing**: Development setup and contribution guidelines
- **Testing**: Test execution and coverage requirements
- **Release Process**: Build and release procedures

### Compatibility Matrix

#### Python CLI Equivalence
| Python Command | Go Equivalent | Status |
|---|---|---|
| `python -m dicom_maker.cli create` | `crgodicom create` | ✅ Required |
| `python -m dicom_maker.cli list` | `crgodicom list` | ✅ Required |
| `python -m dicom_maker.cli send` | `crgodicom send` | ✅ Required |
| `python -m dicom_maker.cli export` | `crgodicom export` | ✅ Required |
| `python -m dicom_maker.cli verify` | `crgodicom verify` | ✅ Required |
| N/A | `crgodicom query` | 🔄 Future |
| N/A | `crgodicom create-template` | ✅ Required |

#### Feature Parity Checklist
- [ ] DICOM 3.0 compliant data generation
- [ ] All supported modalities (CR, CT, MR, US, DX, MG)
- [ ] Burnt-in text with metadata overlay
- [ ] C-ECHO and C-STORE PACS operations
- [ ] PNG and PDF export functionality
- [ ] Study templates and configuration
- [ ] Comprehensive logging system
- [ ] Cross-platform binary distribution
- [ ] Identical CLI interface and options
- [ ] Performance improvements over Python version

### Success Criteria

#### Functional Success
1. **Complete Feature Parity**: All Python CLI functionality replicated
2. **Cross-Platform Operation**: Identical behavior on Windows, macOS, Linux
3. **PACS Compatibility**: Works with all PACS systems supported by Python version
4. **Export Quality**: Generated PNGs and PDFs match Python version quality
5. **CLI Compatibility**: Drop-in replacement for Python CLI commands

#### Performance Success
1. **Speed Improvement**: 3x faster than Python version for large studies
2. **Memory Efficiency**: 50% lower memory usage than Python version
3. **Binary Size**: Single executable under 50MB
4. **Startup Time**: <100ms cold start time
5. **Network Performance**: Improved PACS transmission speeds

#### Distribution Success
1. **Easy Installation**: One-command installation on all platforms
2. **Zero Dependencies**: No runtime dependencies required
3. **Automatic Updates**: Built-in update mechanism
4. **Package Manager Support**: Available via Homebrew, Chocolatey, etc.
5. **Docker Support**: Multi-architecture container images

### Migration Path

#### From Python to Go
1. **Parallel Development**: Go version developed alongside Python version
2. **Feature Validation**: Each Go feature validated against Python equivalent
3. **Beta Testing**: Extensive testing with existing Python users
4. **Documentation Update**: Migration guide and compatibility notes
5. **Gradual Rollout**: Phased replacement of Python version

#### Backward Compatibility
- Configuration files should be compatible between versions
- Study storage format must remain identical
- CLI command syntax must be identical
- Export file formats must be identical
- PACS communication behavior must be identical

---

## Implementation Timeline

### Phase 1: Core Infrastructure (Weeks 1-2)
- Project setup and dependency management
- CLI framework and command structure
- Configuration system and logging
- Basic DICOM data structures

### Phase 2: DICOM Generation (Weeks 3-4)
- DICOM field validation and generation
- Image generation with modality support
- Burnt-in text overlay functionality
- Study template system

### Phase 3: PACS Communication (Weeks 5-6)
- DICOM network protocol implementation
- C-ECHO and C-STORE operations
- Association management and error handling
- Network timeout and retry logic

### Phase 4: Export Functionality (Weeks 7-8)
- PNG export with metadata files
- PDF generation with professional formatting
- Study management and local storage
- File organization and indexing

### Phase 5: Testing & Optimization (Weeks 9-10)
- Comprehensive test suite development
- Performance optimization and benchmarking
- Cross-platform testing and validation
- Documentation and user guides

### Phase 6: Distribution & Release (Weeks 11-12)
- CI/CD pipeline setup
- Binary building and packaging
- Package manager integration
- Release automation and documentation

---

This specification provides a complete blueprint for creating a production-ready, cross-platform Go DICOM utility that maintains full compatibility with the existing Python implementation while offering improved performance and easier distribution.

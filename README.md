# CRGoDICOM

A cross-platform CLI utility written in Go for creating synthetic DICOM data and sending it to PACS systems. This is a Go implementation of the Python [dicom-maker](https://github.com/flatmapit/dicom-maker) with improved performance and easier distribution.

## Features

- **Cross-Platform**: Single binary for Windows, macOS, and Linux (amd64, arm64)
- **DICOM 3.0 Compliant**: Full support for DICOM standard with configurable fields
- **Study Templates**: Built-in and user-defined templates for common modalities
- **PACS Integration**: C-ECHO and C-STORE operations for PACS communication
- **Export Capabilities**: Export to PNG+text files or PDF with metadata
- **Zero Dependencies**: No external runtime requirements
- **Configuration**: YAML-based configuration with CLI overrides

## Quick Start

### Installation

```bash
# Download the latest release for your platform
# Or build from source:
git clone https://github.com/flatmapit/crgodicom.git
cd crgodicom
make build
```

### Basic Usage

```bash
# Create a synthetic DICOM study
crgodicom create --study-count 1 --series-count 2 --image-count 10

# Create study from template
crgodicom create --template chest-xray --series-count 1 --image-count 2

# List local studies
crgodicom list

# Verify PACS connection
crgodicom verify --host localhost --port 11112 --aec CLIENT --aet PACS

# Send study to PACS
crgodicom send --study-id <study-uid> --host localhost --port 11112 --aec CLIENT --aet PACS

# Export study to PNG files
crgodicom export --study-id <study-uid> --format png --output-dir exports/

# Create a new study template
crgodicom create-template --name my-template --modality CT --series-count 2 --image-count 20
```

## Templates

CRGoDICOM comes with built-in templates for common medical imaging modalities:

### Built-in Templates
- **chest-xray**: Computed Radiography (CR) chest imaging
- **ct-chest**: Computed Tomography (CT) chest with multiple series
- **ultrasound-abdomen**: Ultrasound (US) abdominal imaging
- **mammography**: Mammography (MG) breast imaging
- **digital-xray**: Digital X-Ray (DX) imaging
- **mri-brain**: Magnetic Resonance Imaging (MR) brain studies

### Template Examples
📚 **[View Complete Template Examples](docs/template-examples/README.md)** - Comprehensive examples showing:
- Generated images with burnt-in metadata
- PDF reports with all images and technical details
- DICOM metadata dumps
- Usage examples and customization options
- Clinical context and use cases

### Custom Templates
```bash
# Create a custom template
crgodicom create-template --name cardiac-mri \
  --modality MR --series-count 4 --image-count 25 \
  --anatomical-region heart --study-description "Cardiac MRI"

# Use custom template
crgodicom create --template cardiac-mri
```

## Configuration

The application uses a YAML configuration file (`crgodicom.yaml`) in the current working directory. CLI flags override configuration file values.

```yaml
# crgodicom.yaml
dicom:
  org_root: "1.2.840.10008.5.1.4.1.1"

default_pacs:
  host: "localhost"
  port: 11112
  aec: "CRGODICOM"
  aet: "PACS_SERVER"

study_templates:
  chest-xray:
    modality: "CR"
    series_count: 1
    image_count: 2
    anatomical_region: "chest"
    study_description: "Chest X-Ray"
```

## Development

### Building

```bash
# Build for current platform
make build

# Build for all platforms
make build-all

# Run tests
make test

# Development setup
make dev-setup
```

### Testing with Docker PACS

```bash
# Start test PACS servers
make docker-test-setup

# Stop test PACS servers
make docker-test-stop
```

## Status

🚧 **Work in Progress** - This is an active development project.

### Implemented
- ✅ CLI framework with all commands
- ✅ Configuration system (YAML + CLI overrides)
- ✅ Study templates (built-in and user-defined)
- ✅ Basic project structure

### In Progress
- 🔄 DICOM data generation
- 🔄 Image generation (noise patterns)
- 🔄 PACS communication (C-ECHO, C-STORE)

### Planned
- ⏳ Export functionality (PNG, PDF)
- ⏳ Comprehensive testing
- ⏳ Docker test setup
- ⏳ Performance optimization

## License

MIT License - see [LICENSE](LICENSE) file for details.

## Copyright

© 2025 flatmapit.com

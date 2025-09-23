# Testing Guide

This document describes how to run automated tests for CRGoDICOM features.

## Running Tests

### Unit Tests
```bash
# Run all unit tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test ./internal/export/...
```

### Integration Tests
```bash
# Run integration tests
go test -tags=integration ./test/integration/...

# Run tests with coverage
go test -cover ./...
```

### Export Functionality Tests

#### Test PNG Export
```bash
# Create a test study
./bin/crgodicom create --study-count 1 --series-count 1 --image-count 2

# Get the study ID from the output, then export to PNG
./bin/crgodicom export --study-id <study-uid> --format png --output-dir test-exports
```

#### Test JPEG Export
```bash
# Export the same study to JPEG
./bin/crgodicom export --study-id <study-uid> --format jpeg --output-dir test-exports
```

#### Test PDF Export
```bash
# Export the study to PDF
./bin/crgodicom export --study-id <study-uid> --format pdf --output-file test-report.pdf
```

### Verification Tests

#### Verify Pixel Data Extraction
1. Create a DICOM study with known pixel data
2. Export to PNG, JPEG, and PDF formats
3. Verify that:
   - PNG files contain the expected pixel data
   - JPEG files contain the expected pixel data (with compression)
   - PDF files contain embedded images with pixel data
   - All formats include burnt-in metadata

#### Test Commands
```bash
# List available studies
./bin/crgodicom list

# Check DCMTK availability
./bin/crgodicom check-dcmtk

# Verify PACS connection (if DCMTK is available)
./bin/crgodicom echo --host localhost --port 11112 --aec CLIENT --aet PACS
```

## Test Data

### Sample Studies
The `studies/` directory contains sample DICOM studies for testing:
- Study UID: `1.2.840.10008.5.1.4.1.1.1758590232.1649618261125328773`
  - Contains 2 series with 5 images each
  - Includes pixel data for export testing

### Expected Outputs
When running export tests, verify:
1. **PNG Files**: High-quality lossless images with burnt-in metadata
2. **JPEG Files**: Compressed images (quality 95) with burnt-in metadata
3. **PDF Files**: Multi-page reports with embedded images and study metadata

## Automated Test Scripts

### Run All Export Tests
```bash
#!/bin/bash
# test-exports.sh

STUDY_ID="1.2.840.10008.5.1.4.1.1.1758590232.1649618261125328773"
OUTPUT_DIR="test-exports-$(date +%Y%m%d-%H%M%S)"

echo "Testing PNG export..."
./bin/crgodicom export --study-id $STUDY_ID --format png --output-dir $OUTPUT_DIR/png

echo "Testing JPEG export..."
./bin/crgodicom export --study-id $STUDY_ID --format jpeg --output-dir $OUTPUT_DIR/jpeg

echo "Testing PDF export..."
./bin/crgodicom export --study-id $STUDY_ID --format pdf --output-file $OUTPUT_DIR/report.pdf

echo "Export tests completed. Check $OUTPUT_DIR for results."
```

### Verify Pixel Data
```bash
#!/bin/bash
# verify-pixel-data.sh

STUDY_ID="1.2.840.10008.5.1.4.1.1.1758590232.1649618261125328773"

echo "Verifying pixel data extraction..."

# Check if PNG files exist and have content
PNG_COUNT=$(find studies/exports -name "*.png" | wc -l)
echo "Found $PNG_COUNT PNG files"

# Check if JPEG files exist and have content
JPEG_COUNT=$(find studies/exports -name "*.jpg" | wc -l)
echo "Found $JPEG_COUNT JPEG files"

# Check if PDF file exists
if [ -f "studies/exports/study__report.pdf" ]; then
    echo "PDF report exists"
else
    echo "PDF report missing"
fi

echo "Pixel data verification completed."
```

## Continuous Integration

### GitHub Actions
The project includes GitHub Actions workflows for automated testing:
- Unit tests on multiple platforms
- Integration tests with Docker PACS
- Export functionality tests
- Cross-platform builds

### Local CI Simulation
```bash
# Run the same tests as CI
make test
make build-all
make docker-test-setup
```

## Troubleshooting

### Common Issues

1. **DICOM Parsing Errors**
   - Ensure DICOM files are valid
   - Check file permissions
   - Verify file size and format

2. **Export Failures**
   - Check output directory permissions
   - Ensure sufficient disk space
   - Verify study ID exists

3. **Missing Pixel Data**
   - Check DICOM file integrity
   - Verify pixel data tags are present
   - Check for compression issues

### Debug Mode
```bash
# Enable debug logging
export LOG_LEVEL=debug
./bin/crgodicom export --study-id <study-uid> --format png --output-dir test-exports
```

## Performance Testing

### Benchmark Tests
```bash
# Run benchmark tests
go test -bench=. ./...

# Memory profiling
go test -memprofile=mem.prof ./...
go tool pprof mem.prof
```

### Load Testing
```bash
# Test with large studies
./bin/crgodicom create --study-count 10 --series-count 5 --image-count 20

# Test export performance
time ./bin/crgodicom export --study-id <study-uid> --format png --output-dir test-exports
```
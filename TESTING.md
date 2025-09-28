# CRGoDICOM Testing Guide

This document provides comprehensive testing procedures for CRGoDICOM, covering unit tests, integration tests, conformance validation, and performance testing.

## Table of Contents

1. [Testing Overview](#testing-overview)
2. [Unit Testing](#unit-testing)
3. [Integration Testing](#integration-testing)
4. [DICOM Conformance Testing](#dicom-conformance-testing)
5. [Performance Testing](#performance-testing)
6. [Test Data Management](#test-data-management)
7. [Continuous Integration](#continuous-integration)

## Testing Overview

CRGoDICOM implements a comprehensive testing strategy to ensure:
- **DICOM Compliance**: Full conformance to DICOM 3.0 standard
- **Data Integrity**: Accurate metadata and pixel data generation
- **Cross-Platform Compatibility**: Consistent behavior across platforms
- **Performance**: Efficient generation and processing
- **Integration**: Reliable PACS communication

### Test Categories

- **Unit Tests**: Individual component testing
- **Integration Tests**: End-to-end workflow testing
- **Conformance Tests**: DICOM standard compliance validation
- **Performance Tests**: Load and stress testing
- **Regression Tests**: Automated testing for bug prevention

## Unit Testing

### Running Unit Tests

```bash
# Run all unit tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/dicom/...
go test ./internal/cli/...
go test ./pkg/types/...
```

### Test Structure

```
test/
├── unit/
│   ├── dicom/
│   │   ├── generator_test.go
│   │   ├── metadata_test.go
│   │   ├── uid_generator_test.go
│   │   └── conformance_test.go
│   ├── cli/
│   │   ├── create_test.go
│   │   ├── export_test.go
│   │   └── list_test.go
│   └── types/
│       └── dicom_test.go
├── integration/
│   ├── pacs_test.go
│   ├── export_test.go
│   └── workflow_test.go
└── fixtures/
    ├── test-studies/
    ├── expected-metadata/
    └── conformance-cases/
```

### Key Unit Test Areas

#### DICOM Generator Tests
- UID generation uniqueness and format validation
- Metadata completeness and correctness
- Image generation for all modalities
- Pixel data integrity and size validation

#### Conformance Tests
- Type 1 attribute validation (required)
- Type 2 attribute validation (required but may be empty)
- Type 3 attribute validation (optional)
- UID format compliance
- Date/time format validation

#### CLI Command Tests
- Parameter validation
- Configuration loading
- Error handling
- Output formatting

## Integration Testing

### End-to-End Workflow Testing

```bash
# Run integration tests
go test -tags=integration ./test/integration/...

# Test complete study generation workflow
go test -tags=integration ./test/integration/workflow_test.go

# Test PACS integration (requires DCMTK)
go test -tags=integration ./test/integration/pacs_test.go
```

### Integration Test Scenarios

#### Study Generation Workflow
1. **Template Loading**: Load built-in and custom templates
2. **Study Creation**: Generate complete studies with all metadata
3. **File Writing**: Write DICOM files to disk
4. **Validation**: Verify file integrity and DICOM compliance
5. **Export**: Test export to various formats

#### PACS Integration Workflow
1. **Connection Testing**: C-ECHO operations
2. **Study Transmission**: C-STORE operations
3. **Error Handling**: Network failures and timeouts
4. **Authentication**: AE Title validation

#### Export Workflow
1. **Image Export**: PNG, JPEG format generation
2. **PDF Reports**: Multi-page PDF creation
3. **Metadata Extraction**: DICOM tag extraction
4. **File Organization**: Directory structure validation

## DICOM Conformance Testing

### Conformance Levels

CRGoDICOM supports multiple conformance levels:

- **Basic Conformance**: Essential DICOM attributes only
- **Partial Conformance**: Most Type 1 and Type 2 attributes
- **Full Conformance**: Complete DICOM 3.0 compliance
- **Enterprise Conformance**: Advanced features and validation

### Running Conformance Tests

```bash
# Run conformance tests
go test ./internal/dicom/conformance_test.go

# Test specific conformance level
go test -run TestFullConformance ./internal/dicom/

# Generate conformance report
go test -run TestConformanceReport ./internal/dicom/
```

### Conformance Test Cases

#### Patient Module Conformance
- Patient Name format validation (LAST^FIRST^MIDDLE)
- Patient ID format validation
- Patient Birth Date format (YYYYMMDD)
- Patient Sex validation (M/F/O)

#### Study Module Conformance
- Study Instance UID uniqueness and format
- Study Date/Time format validation
- Accession Number format validation
- Study Description content validation

#### Series Module Conformance
- Series Instance UID uniqueness and format
- Modality code validation
- Series Number format validation
- Series Description content validation

#### Image Module Conformance
- SOP Instance UID uniqueness and format
- SOP Class UID validation
- Instance Number format validation
- Image dimensions validation
- Pixel data integrity validation

### Conformance Reporting

The conformance checker generates detailed reports:

```bash
# Generate detailed conformance report
crgodicom create --modality CT --conformance-check --verbose

# Example output:
# ✅ Study 1 passed conformance check (Score: 95.2%)
# ⚠️  Patient: PatientBirthDate: Patient Birth Date is recommended (Type 2)
# ⚠️  Study: StudyID: Study ID is recommended (Type 2)
```

## Performance Testing

### Load Testing

```bash
# Generate large studies for performance testing
crgodicom create --study-count 100 --series-count 5 --image-count 20 --modality CT

# Benchmark UID generation
go test -bench=BenchmarkUIDGeneration ./internal/dicom/

# Benchmark image generation
go test -bench=BenchmarkImageGeneration ./internal/dicom/
```

### Performance Benchmarks

#### UID Generation Performance
- **Target**: 10,000 UIDs/second
- **Memory**: < 1MB for 100,000 UIDs
- **Uniqueness**: 100% unique across concurrent generation

#### Image Generation Performance
- **Target**: 100 images/second (512x512, 16-bit)
- **Memory**: < 100MB for 1000 images
- **Quality**: Consistent modality-specific patterns

#### Study Generation Performance
- **Target**: 10 complete studies/second
- **Memory**: < 500MB for 100 studies
- **Metadata**: Complete Type 1/Type 2 coverage

### Memory Profiling

```bash
# Run with memory profiling
go test -memprofile=mem.prof ./internal/dicom/
go tool pprof mem.prof

# Run with CPU profiling
go test -cpuprofile=cpu.prof ./internal/dicom/
go tool pprof cpu.prof
```

## Test Data Management

### Test Fixtures

Test fixtures are stored in `test/fixtures/` and include:

- **Expected Metadata**: Reference DICOM metadata for validation
- **Conformance Cases**: Test cases for each conformance level
- **Sample Studies**: Pre-generated studies for regression testing
- **Error Cases**: Invalid data for error handling tests

### Test Data Generation

```bash
# Generate test data for specific modality
go run scripts/generate-test-data.go --modality CT --count 10

# Generate conformance test cases
go run scripts/generate-conformance-cases.go --level full

# Clean up test data
go run scripts/cleanup-test-data.go
```

### Test Data Validation

```bash
# Validate test data integrity
go test -run TestDataIntegrity ./test/fixtures/

# Validate conformance test cases
go test -run TestConformanceCases ./test/fixtures/
```

## Continuous Integration

### GitHub Actions Workflow

The CI pipeline includes:

1. **Code Quality**: Linting, formatting, and static analysis
2. **Unit Tests**: All unit tests with coverage reporting
3. **Integration Tests**: End-to-end workflow testing
4. **Conformance Tests**: DICOM compliance validation
5. **Performance Tests**: Benchmark validation
6. **Cross-Platform**: Testing on Windows, macOS, and Linux

### Local CI Simulation

```bash
# Run full CI pipeline locally
make ci-test

# Run specific CI stages
make lint
make test-unit
make test-integration
make test-conformance
make test-performance
```

### Test Coverage Requirements

- **Unit Tests**: > 90% coverage
- **Integration Tests**: > 80% coverage
- **Conformance Tests**: 100% Type 1 attribute coverage
- **Performance Tests**: All critical paths benchmarked

## Troubleshooting Tests

### Common Test Issues

#### DCMTK Not Available
```bash
# Check DCMTK installation
crgodicom check-dcmtk

# Install DCMTK for testing
# macOS: brew install dcmtk
# Ubuntu: sudo apt-get install dcmtk
```

#### Test Data Cleanup
```bash
# Clean up test data
rm -rf test-data/
rm -rf studies/
rm -rf exports/

# Reset test environment
make test-clean
```

#### Memory Issues
```bash
# Run tests with memory limits
go test -memprofile=mem.prof -memprofilerate=1 ./...

# Analyze memory usage
go tool pprof mem.prof
```

### Test Debugging

```bash
# Run tests with debug output
go test -v -run TestSpecificTest ./internal/dicom/

# Run tests with race detection
go test -race ./...

# Run tests with timeout
go test -timeout 30s ./...
```

## Test Automation

### Automated Test Execution

```bash
# Run all tests automatically
make test-all

# Run tests with reporting
make test-report

# Run tests in parallel
make test-parallel
```

### Test Reporting

Test reports include:
- **Coverage Reports**: Line and branch coverage
- **Conformance Reports**: DICOM compliance scores
- **Performance Reports**: Benchmark results
- **Error Reports**: Failed test analysis

### Test Maintenance

- **Weekly**: Run full test suite
- **Daily**: Run critical path tests
- **On Change**: Run affected component tests
- **Before Release**: Run complete validation suite

## Conclusion

This testing framework ensures CRGoDICOM maintains high quality and DICOM compliance across all supported platforms and use cases. Regular testing and validation are essential for maintaining the tool's reliability in medical system integration scenarios.
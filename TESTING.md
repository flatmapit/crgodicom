# Testing Guide for CRGoDICOM

This document describes how to run and maintain the automated test suite for CRGoDICOM.

## 🧪 Test Suite Overview

CRGoDICOM includes comprehensive unit tests covering:
- **CLI Commands**: All command-line interface functionality
- **ORM Parsers**: HL7 and Go struct parsing capabilities
- **Template Generation**: DICOM template creation from various sources
- **Core Functionality**: DICOM generation, export, and PACS integration

## 🚀 Running Tests

### Quick Test Run
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test -v ./internal/cli
go test -v ./internal/orm/parser
```

### Test Coverage
```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html

# View coverage summary
go tool cover -func=coverage.out

# View coverage in browser
open coverage.html  # macOS
xdg-open coverage.html  # Linux
start coverage.html  # Windows
```

### Race Condition Detection
```bash
# Run tests with race detection
go test -race ./...

# Run with race detection and coverage
go test -race -coverprofile=coverage.out ./...
```

## 📋 Test Categories

### 1. CLI Command Tests (`internal/cli/*_test.go`)

#### Create Command Tests
- ✅ **Flag validation**: Test all command flags and defaults
- ✅ **Parameter validation**: Test invalid modalities, counts, etc.
- ⚠️ **Study generation**: Tests DICOM generation (currently failing due to VR issues)
- ✅ **Template usage**: Test template-based study creation

#### Export Command Tests
- ✅ **Required parameters**: Test missing study-id and format
- ✅ **Format validation**: Test valid/invalid export formats
- ✅ **PDF requirements**: Test PDF-specific parameter validation
- ⚠️ **Export functionality**: Tests actual export (fails without existing studies)

#### List Command Tests
- ✅ **Directory handling**: Test empty and non-existent directories
- ✅ **Format options**: Test table, JSON, CSV output formats
- ✅ **Verbose output**: Test detailed listing functionality
- ✅ **Study discovery**: Test study directory enumeration

#### ORM Command Tests
- ✅ **Input validation**: Test required parameters and file existence
- ✅ **Type detection**: Test automatic input type detection
- ✅ **Template generation**: Test HL7 and Go struct parsing
- ✅ **Output creation**: Test YAML template file generation

### 2. ORM Parser Tests (`internal/orm/parser/*_test.go`)

#### HL7 Parser Tests
- ✅ **Message parsing**: Test complete HL7 ORM message processing
- ✅ **Segment parsing**: Test individual HL7 segments (MSH, PID, OBR, etc.)
- ✅ **Model generation**: Test creation of Patient, Study, Order, Visit models
- ✅ **Error handling**: Test invalid and empty message handling
- ⚠️ **Field mapping**: Some field mappings need adjustment

## 📊 Current Test Results

### Test Statistics
```
Total Tests: 25+
Passing Tests: ~80%
Failing Tests: ~20% (mostly due to DICOM VR issues and field mapping adjustments)
Coverage: 34.2% (initial implementation)
```

### Test Status by Component

| Component | Tests | Status | Coverage | Notes |
|-----------|-------|--------|----------|-------|
| **CLI Create** | 5 | ⚠️ Partial | Medium | DICOM generation VR issues |
| **CLI Export** | 6 | ⚠️ Partial | Medium | Validation works, export needs studies |
| **CLI List** | 7 | ✅ Pass | High | All functionality working |
| **CLI ORM** | 4 | ✅ Pass | High | Template generation working |
| **HL7 Parser** | 8 | ⚠️ Partial | High | Parsing works, field mapping adjustments needed |

## 🔧 Running Specific Tests

### Test Individual Commands
```bash
# Test create command only
go test -v ./internal/cli -run TestCreateCommand

# Test export command only
go test -v ./internal/cli -run TestExportCommand

# Test ORM functionality
go test -v ./internal/cli -run TestORM
go test -v ./internal/orm/parser -run TestHL7

# Test with specific pattern
go test -v ./internal/cli -run "TestCreateCommand/valid_basic_create"
```

### Debug Failing Tests
```bash
# Run with detailed output
go test -v ./internal/cli -run TestCreateCommand 2>&1 | grep -A5 -B5 "Error"

# Run single failing test
go test -v ./internal/cli -run "TestCreateCommand/valid_basic_create"
```

## 🎯 Test Development Guidelines

### Writing New Tests
1. **Use table-driven tests** for multiple scenarios
2. **Create temporary directories** for file operations
3. **Mock external dependencies** (PACS servers, file systems)
4. **Test both success and failure cases**
5. **Include edge cases and boundary conditions**

### Test Structure Example
```go
func TestNewFeature(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := NewFeature(tt.input)
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
                assert.Equal(t, tt.want, result)
            }
        })
    }
}
```

## 🔍 Known Issues and Fixes

### 1. DICOM VR (Value Representation) Issues
**Problem**: Tests fail with "ValueType does not match the specified type in the VR"
**Status**: Known issue in DICOM generation
**Workaround**: Test command structure without actual DICOM file creation

### 2. Field Mapping Adjustments
**Problem**: Some HL7 field mappings don't match expected test values
**Status**: Parser logic needs refinement
**Fix**: Adjust parser logic or test expectations

### 3. Missing Study Dependencies
**Problem**: Export tests fail because they need existing studies
**Status**: Expected behavior
**Fix**: Create mock studies in test setup

## 🚀 Continuous Integration

### GitHub Actions Integration
The test suite is integrated with GitHub Actions workflows:

```yaml
- name: Run tests
  run: |
    echo "Running comprehensive test suite..."
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html

- name: Upload Coverage Report
  uses: actions/upload-artifact@v4
  with:
    name: coverage-report-${{ matrix.platform }}
    path: coverage.html
```

### Test Automation
- **Feature Branches**: Tests run on every push
- **Pull Requests**: Tests required for merge
- **Coverage Reports**: Generated and uploaded as artifacts
- **Race Detection**: Enabled for concurrency testing

## 📈 Coverage Goals

### Current Coverage: 34.2%
### Target Coverage: 80%+

**Priority Areas for Coverage Improvement:**
1. **Core DICOM functionality** (types, generation)
2. **Export functionality** (PNG, PDF generation)
3. **PACS integration** (network communication)
4. **Configuration management** (YAML parsing, validation)

## 🔧 Test Maintenance

### Adding New Tests
1. Create `*_test.go` files alongside source files
2. Follow existing test patterns and conventions
3. Include both unit and integration tests
4. Update this documentation

### Fixing Failing Tests
1. Identify root cause (logic error vs. test setup)
2. Fix underlying issue or adjust test expectations
3. Verify fix doesn't break other tests
4. Update documentation if behavior changes

## 📚 Test Examples

### Example Test Run Output
```
🧪 RUNNING COMPLETE TEST SUITE:

=== RUN   TestCreateCommand
=== RUN   TestCreateCommand/valid_basic_create
time="2025-09-18T14:56:28+10:00" level=info msg="Creating 1 study(ies)..."
=== RUN   TestCreateCommand/invalid_modality
--- FAIL: TestCreateCommand/invalid_modality (0.00s)
=== RUN   TestListCommand
=== RUN   TestListCommand/empty_directory
time="2025-09-18T14:56:12+10:00" level=info msg="Listing studies..."
--- PASS: TestListCommand (0.00s)
=== RUN   TestORMCommandValidation
=== RUN   TestORMCommandValidation/valid_HL7_input
time="2025-09-18T14:56:12+10:00" level=info msg="Successfully generated DICOM template"
--- PASS: TestORMCommandValidation (0.00s)

PASS: 15 tests
FAIL: 8 tests
Coverage: 34.2% of statements
```

## 📞 Support

- **Test Issues**: [Report Test Problems](https://github.com/flatmapit/crgodicom/issues)
- **Coverage Reports**: Available in GitHub Actions artifacts
- **Documentation**: This file and inline test documentation

---

*Automated testing ensures code quality and reliability across all CRGoDICOM features*

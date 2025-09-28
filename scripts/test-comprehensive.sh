#!/bin/bash

# Comprehensive DICOM Test Suite
# Runs both generation and export tests

set -e

echo "🧪 Comprehensive DICOM Test Suite"
echo "=================================="
echo ""

# Check if binary exists
if [ ! -f "./bin/crgodicom" ]; then
    echo "❌ Binary not found: ./bin/crgodicom"
    echo "   Please run 'make build' first"
    exit 1
fi

# Check DCMTK installation
if ! command -v dcmdump &> /dev/null; then
    echo "❌ DCMTK not found in PATH"
    echo "   Please install DCMTK first"
    exit 1
fi

echo "✅ Prerequisites check passed"
echo ""

# Run generation tests
echo "🔬 Running DICOM generation tests..."
if ./scripts/test-dicom-generation.sh; then
    echo "✅ Generation tests passed"
else
    echo "❌ Generation tests failed"
    exit 1
fi
echo ""

# Run export tests
echo "📤 Running DICOM export tests..."
if ./scripts/test-dicom-export.sh; then
    echo "✅ Export tests passed"
else
    echo "❌ Export tests failed"
    exit 1
fi
echo ""

# Final validation
echo "🔍 Final validation..."
echo "======================"

# Check file counts
studies_count=$(find test-data/studies -type d -mindepth 1 -maxdepth 1 | wc -l)
dicoms_count=$(find test-data/studies -name "*.dcm" | wc -l)
exports_count=$(find test-data/exports -type f 2>/dev/null | wc -l)

echo "📁 Studies generated: $studies_count"
echo "📄 DICOM files: $dicoms_count"
echo "📤 Exported files: $exports_count"

# Check file sizes
studies_size=$(du -sh test-data/studies | cut -f1)
exports_size=$(du -sh test-data/exports 2>/dev/null | cut -f1 || echo "0")

echo "💾 Studies size: $studies_size"
echo "💾 Exports size: $exports_size"

# Validate a few DICOM files
echo ""
echo "🔍 DICOM validation sample:"
validation_count=0
for dcm_file in $(find test-data/studies -name "*.dcm" | head -3); do
    if dcmdump "$dcm_file" > /dev/null 2>&1; then
        validation_count=$((validation_count + 1))
    fi
done
echo "✅ $validation_count/3 sample DICOM files validated"

echo ""
echo "🎉 All tests completed successfully!"
echo "📊 Test data available in test-data/ directory"
echo "   (This directory is ignored by git)"

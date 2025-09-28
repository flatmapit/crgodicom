#!/bin/bash

# Cleanup Test Data Script
# Removes all generated test data

echo "🧹 Cleaning up DICOM test data..."
echo "=================================="
echo ""

# Remove test data directories
if [ -d "test-data" ]; then
    echo "🗑️  Removing test-data directory..."
    rm -rf test-data
    echo "✅ test-data directory removed"
else
    echo "ℹ️  test-data directory not found"
fi

# Remove main studies and exports directories
if [ -d "studies" ]; then
    echo "🗑️  Removing studies directory..."
    rm -rf studies
    echo "✅ studies directory removed"
else
    echo "ℹ️  studies directory not found"
fi

if [ -d "exports" ]; then
    echo "🗑️  Removing exports directory..."
    rm -rf exports
    echo "✅ exports directory removed"
else
    echo "ℹ️  exports directory not found"
fi

# Remove any stray DICOM files
dicom_files=$(find . -maxdepth 1 -name "*.dcm" 2>/dev/null | wc -l)
if [ "$dicom_files" -gt 0 ]; then
    echo "🗑️  Removing $dicom_files stray DICOM files..."
    find . -maxdepth 1 -name "*.dcm" -delete
    echo "✅ Stray DICOM files removed"
else
    echo "ℹ️  No stray DICOM files found"
fi

echo ""
echo "🎉 Cleanup completed!"
echo "📁 All test data has been removed"
echo "   (Directories will be recreated on next test run)"

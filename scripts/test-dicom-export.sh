#!/bin/bash

# DICOM Export Test Script
# Tests the export functionality for various formats

set -e

echo "📤 DICOM Export Test Suite"
echo "=========================="
echo ""

# Check if we have studies to export
if [ ! -d "test-data/studies" ] || [ -z "$(find test-data/studies -name "*.dcm" 2>/dev/null)" ]; then
    echo "❌ No DICOM studies found in test-data/studies/"
    echo "   Please run test-dicom-generation.sh first"
    exit 1
fi

# Clean up previous exports
echo "🧹 Cleaning up previous exports..."
rm -rf test-data/exports/*
echo "✅ Cleanup complete"
echo ""

# Find all studies
studies=($(find test-data/studies -type d -mindepth 1 -maxdepth 1))

if [ ${#studies[@]} -eq 0 ]; then
    echo "❌ No studies found to export"
    exit 1
fi

echo "📊 Found ${#studies[@]} studies to export"
echo ""

# Test export formats
formats=("png" "jpeg" "pdf")

for study_dir in "${studies[@]}"; do
    study_id=$(basename "$study_dir")
    echo "🔬 Testing exports for study: $study_id"
    
    # Get study info
    first_dcm=$(find "$study_dir" -name "*.dcm" | head -1)
    if [ -n "$first_dcm" ]; then
        patient_name=$(dcmdump "$first_dcm" | grep "PatientName" | head -1 | sed 's/.*\[\(.*\)\].*/\1/')
        modality=$(dcmdump "$first_dcm" | grep "Modality" | head -1 | sed 's/.*\[\(.*\)\].*/\1/')
        echo "   👤 Patient: $patient_name"
        echo "   🔬 Modality: $modality"
    fi
    
    # Test each export format
    for format in "${formats[@]}"; do
        echo "   📤 Testing $format export..."
        
        case $format in
            "png"|"jpeg")
                output_dir="test-data/exports/${study_id}_${format}"
                if ./bin/crgodicom export \
                    --study-id "$study_id" \
                    --format "$format" \
                    --output-dir "$output_dir" \
                    --input-dir test-data/studies > /dev/null 2>&1; then
                    file_count=$(find "$output_dir" -name "*.$format" 2>/dev/null | wc -l)
                    echo "     ✅ $format export successful ($file_count files)"
                else
                    echo "     ❌ $format export failed"
                fi
                ;;
            "pdf")
                output_file="test-data/exports/${study_id}_report.pdf"
                if ./bin/crgodicom export \
                    --study-id "$study_id" \
                    --format "pdf" \
                    --output-file "$output_file" \
                    --input-dir test-data/studies > /dev/null 2>&1; then
                    if [ -f "$output_file" ]; then
                        file_size=$(du -h "$output_file" | cut -f1)
                        echo "     ✅ PDF export successful ($file_size)"
                    else
                        echo "     ❌ PDF file not created"
                    fi
                else
                    echo "     ❌ PDF export failed"
                fi
                ;;
        esac
    done
    echo ""
done

echo "📊 Export Test Summary:"
echo "========================"
total_exports=$(find test-data/exports -type f 2>/dev/null | wc -l)
echo "📄 Total exported files: $total_exports"
echo "💾 Total export size: $(du -sh test-data/exports 2>/dev/null | cut -f1 || echo "0")"
echo ""
echo "🎉 DICOM export tests completed!"

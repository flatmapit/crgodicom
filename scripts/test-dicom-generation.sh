#!/bin/bash

# DICOM Generation Test Script
# Tests the creation of various DICOM modalities

set -e

echo "🧪 DICOM Generation Test Suite"
echo "================================"
echo ""

# Clean up previous test data
echo "🧹 Cleaning up previous test data..."
rm -rf test-data/studies/*
rm -rf test-data/exports/*
rm -rf test-data/temp/*
echo "✅ Cleanup complete"
echo ""

# Test different modalities
modalities=("CR" "CT" "MR" "US" "MG")
patient_names=("TEST^PATIENT^CR" "TEST^PATIENT^CT" "TEST^PATIENT^MR" "TEST^PATIENT^US" "TEST^PATIENT^MG")

for i in "${!modalities[@]}"; do
    modality="${modalities[$i]}"
    patient_name="${patient_names[$i]}"
    
    echo "🔬 Testing $modality modality generation..."
    
    # Generate study
    ./bin/crgodicom create \
        --modality "$modality" \
        --series-count 1 \
        --image-count 2 \
        --patient-name "$patient_name" \
        --patient-id "${modality}001" \
        --study-description "${modality} Test Study" \
        --output-dir test-data/studies \
        --verbose
    
    # Validate generated files
    study_dir=$(find test-data/studies -name "*.dcm" | head -1 | xargs dirname | xargs dirname)
    if [ -d "$study_dir" ]; then
        echo "✅ $modality study generated: $(basename "$study_dir")"
        
        # Count DICOM files
        dicom_count=$(find "$study_dir" -name "*.dcm" | wc -l)
        echo "   📊 Generated $dicom_count DICOM files"
        
        # Validate with DCMTK
        first_dcm=$(find "$study_dir" -name "*.dcm" | head -1)
        if dcmdump "$first_dcm" > /dev/null 2>&1; then
            echo "   ✅ DCMTK validation passed"
        else
            echo "   ❌ DCMTK validation failed"
        fi
    else
        echo "❌ $modality study generation failed"
    fi
    echo ""
done

echo "📊 Generation Test Summary:"
echo "============================"
total_studies=$(find test-data/studies -type d -mindepth 1 -maxdepth 1 | wc -l)
total_dicoms=$(find test-data/studies -name "*.dcm" | wc -l)
echo "📁 Total studies generated: $total_studies"
echo "📄 Total DICOM files: $total_dicoms"
echo "💾 Total size: $(du -sh test-data/studies | cut -f1)"
echo ""
echo "🎉 DICOM generation tests completed!"

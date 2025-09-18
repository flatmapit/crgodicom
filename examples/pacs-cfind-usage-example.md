# PACS CFIND Template Generation Examples

This document demonstrates how to use the `crgodicom pacs-cfind` command to generate DICOM templates from PACS studies using DICOM CFIND operations.

## Overview

The `pacs-cfind` command allows you to:
- Query PACS servers for existing studies using DICOM CFIND operations
- Generate DICOM templates from PACS study metadata
- Work with saved CFIND response files for offline template generation

## Prerequisites

- DCMTK installed and available (check with `crgodicom check-dcmtk`)
- Access to a PACS server with appropriate DICOM services
- Network connectivity to the PACS server

## Example 1: Generate Template from CFIND Response File

Use a pre-saved CFIND response JSON file to generate a template:

```bash
# Generate template from existing CFIND response
crgodicom pacs-cfind \
  --input examples/pacs-cfind-response-example.json \
  --output templates/pacs-ct-chest-template.yaml \
  --template-name "pacs-ct-chest-contrast" \
  --modality "CT" \
  --series-count 3 \
  --image-count 25 \
  --verbose
```

**Expected Output:**
```
📋 Generated Template Summary:
• Template Name: pacs-ct-chest-contrast
• Modality: CT
• Series Count: 3
• Image Count: 25
• Patient: SMITH^JOHN^MICHAEL (ID: P123456789)
• Accession: 2025TY0000001
• Study Description: CT Chest with Contrast

🎯 Usage:
crgodicom create --template pacs-ct-chest-contrast

📁 Output File: templates/pacs-ct-chest-template.yaml
```

## Example 2: Query PACS by Study Instance UID

Query a PACS server for a specific study and generate a template:

```bash
# Query PACS for specific study
crgodicom pacs-cfind \
  --study-uid "1.2.840.113619.2.5.1762583153.215519.978957063.78" \
  --host pacs.hospital.local \
  --port 4242 \
  --aec DICOM_CLIENT \
  --aet PACS_SERVER \
  --template-name "hospital-ct-study" \
  --modality "CT" \
  --test-connection \
  --verbose
```

## Example 3: Query PACS by Patient ID

Search for studies by patient ID:

```bash
# Find all studies for a patient
crgodicom pacs-cfind \
  --patient-id "P123456789" \
  --host pacs.hospital.local \
  --port 4242 \
  --aec DICOM_CLIENT \
  --aet PACS_SERVER \
  --template-name "patient-studies" \
  --test-connection \
  --verbose
```

## Example 4: Save CFIND Response for Later Use

Save the CFIND response to a file for later template generation:

```bash
# Query PACS and save response
crgodicom pacs-cfind \
  --study-uid "1.2.840.113619.2.5.1762583153.215519.978957063.78" \
  --host pacs.hospital.local \
  --port 4242 \
  --aec DICOM_CLIENT \
  --aet PACS_SERVER \
  --save-response cfind-response.json \
  --test-connection \
  --verbose

# Later, generate template from saved response
crgodicom pacs-cfind \
  --input cfind-response.json \
  --output templates/delayed-template.yaml \
  --template-name "delayed-generation"
```

## Example 5: Batch Processing Multiple Studies

Process multiple studies from a CFIND response:

```bash
# Generate templates for all studies in response file
crgodicom pacs-cfind \
  --input examples/pacs-cfind-response-example.json \
  --output templates/batch-template.yaml \
  --template-name "multi-study-template" \
  --series-count 2 \
  --image-count 15 \
  --verbose
```

## Generated Template Structure

The generated templates include DICOM tags extracted from PACS studies:

```yaml
study_templates:
  pacs-ct-chest-contrast:
    modality: "CT"
    series_count: 3
    image_count: 25
    anatomical_region: "chest"
    study_description: "CT Chest with Contrast - Generated from PACS"
    patient_name: "SMITH^JOHN^MICHAEL"
    patient_id: "P123456789"
    accession_number: "2025TY0000001"
    
    custom_tags:
      patient:
        "(0010,0010)": "{{ .Patient.PatientName }}"      # Patient Name
        "(0010,0020)": "{{ .Patient.PatientID }}"        # Patient ID
        "(0010,0030)": "{{ .Patient.PatientBirthDate }}" # Birth Date
        "(0010,0040)": "{{ .Patient.PatientSex }}"       # Patient Sex
        
      study:
        "(0020,000D)": "{{ .Study.StudyInstanceUID }}"   # Study Instance UID
        "(0008,0020)": "{{ .Study.StudyDate }}"          # Study Date
        "(0008,0030)": "{{ .Study.StudyTime }}"          # Study Time
        "(0008,0050)": "{{ .Study.AccessionNumber }}"    # Accession Number
        "(0008,1030)": "{{ .Study.StudyDescription }}"   # Study Description
        "(0008,0060)": "{{ .Study.Modality }}"           # Modality
        "(0008,0090)": "{{ .Study.ReferringPhysician }}" # Referring Physician
        
      institution:
        "(0008,0080)": "{{ .Institution.InstitutionName }}" # Institution Name
    
    metadata:
      generated_from: "PACS_CFIND"
      generated_at: "2025-09-18T15:45:00Z"
      source_study_uid: "1.2.840.113619.2.5.1762583153.215519.978957063.78"
      instance_count: "125"
      series_count: "3"
```

## Using Generated Templates

Once generated, use the templates to create DICOM studies:

```bash
# Create study using the generated template
crgodicom create --template pacs-ct-chest-contrast

# Customize patient data
crgodicom create --template pacs-ct-chest-contrast \
  --patient-name "DOE^JANE^ELIZABETH" \
  --patient-id "P999888777" \
  --accession-number "2025TY0000003"

# Export the generated study
crgodicom export --study-id <study-id> --format pdf \
  --output-file pacs-generated-study.pdf
```

## Error Handling

### Common Issues

1. **DCMTK Not Available**
   ```
   Error: DCMTK not available: findscu not found
   Solution: Install DCMTK or run 'crgodicom check-dcmtk --install-help'
   ```

2. **PACS Connection Failed**
   ```
   Error: PACS connection test failed
   Solution: Check network connectivity, host/port, and PACS configuration
   ```

3. **No Studies Found**
   ```
   Error: no studies found matching the query criteria
   Solution: Verify Study Instance UID or query parameters
   ```

### Troubleshooting

- Use `--test-connection` to verify PACS connectivity
- Use `--verbose` for detailed logging
- Check DCMTK installation with `crgodicom check-dcmtk`
- Verify PACS server supports CFIND operations

## Advanced Usage

### Custom Field Mappings

The PACS parser automatically maps common DICOM tags, but you can customize the generated templates:

```bash
# Generate template with custom series count
crgodicom pacs-cfind \
  --input cfind-response.json \
  --template-name "custom-series" \
  --series-count 5 \
  --image-count 50
```

### Integration with Hospital Workflow

```bash
#!/bin/bash
# hospital-pacs-integration.sh

# 1. Query PACS for new studies
STUDY_UID="1.2.840.113619.2.5.1762583153.215519.978957063.78"

# 2. Generate template from PACS study
crgodicom pacs-cfind \
  --study-uid "$STUDY_UID" \
  --host pacs.hospital.local --port 4242 \
  --aec HOSPITAL_CLIENT --aet PACS_SERVER \
  --output "templates/pacs-$(date +%Y%m%d).yaml" \
  --template-name "pacs-study-$(date +%Y%m%d)" \
  --test-connection

# 3. Create new study using template
STUDY_ID=$(crgodicom create --template "pacs-study-$(date +%Y%m%d)" --output-json | jq -r '.study_id')

# 4. Export for review
crgodicom export --study-id "$STUDY_ID" --format pdf \
  --output-file "reports/pacs-study-$STUDY_ID.pdf"

echo "✅ Generated study $STUDY_ID from PACS template"
```

## Best Practices

1. **Always test PACS connectivity** with `--test-connection`
2. **Save CFIND responses** for offline processing and debugging
3. **Use descriptive template names** for easy identification
4. **Validate generated templates** before using in production
5. **Monitor DICOM tag mappings** to ensure data integrity
6. **Keep DCMTK updated** for best compatibility

## Related Commands

- `crgodicom check-dcmtk` - Verify DCMTK installation
- `crgodicom create` - Create studies from templates
- `crgodicom list` - List generated studies
- `crgodicom export` - Export studies to various formats
- `crgodicom dcmtk` - Send studies to PACS using DCMTK

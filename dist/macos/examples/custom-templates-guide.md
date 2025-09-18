# Custom DICOM Templates with Custom Tags Guide

This guide shows you how to create and use custom DICOM templates with custom tags in crgodicom.

## Table of Contents
1. [Current Template System](#current-template-system)
2. [Creating Basic Templates](#creating-basic-templates)
3. [Custom Tags Implementation](#custom-tags-implementation)
4. [Examples](#examples)
5. [Advanced Usage](#advanced-usage)

## Current Template System

### Built-in Templates
crgodicom comes with several built-in templates:

```bash
# List available templates
./bin/crgodicom create --help

# Use a built-in template
./bin/crgodicom create --template chest-xray
./bin/crgodicom create --template ct-chest
./bin/crgodicom create --template ultrasound-abdomen
```

### Template Configuration (crgodicom.yaml)
```yaml
study_templates:
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
```

## Creating Basic Templates

### 1. Using CLI Command
```bash
# Create a custom template
./bin/crgodicom create-template \
  --name "my-ct-brain" \
  --modality "CT" \
  --series-count 3 \
  --image-count 25 \
  --anatomical-region "brain" \
  --study-description "CT Brain with Contrast" \
  --output-file "my-ct-brain-template.yaml"
```

### 2. Manual Template Creation
Create a template file `custom-us-cardiac.yaml`:
```yaml
template:
  name: "us-cardiac"
  modality: "US"
  series_count: 2
  image_count: 15
  anatomical_region: "heart"
  study_description: "Cardiac Ultrasound"
  patient_name: "CARDIAC^PATIENT"
  patient_id: "CARDIAC001"
  accession_number: "ACC123456"

usage:
  description: "Cardiac ultrasound study template"
  example: "crgodicom create --template us-cardiac"
```

### 3. Add to Configuration
Add your template to `crgodicom.yaml`:
```yaml
study_templates:
  # ... existing templates ...
  
  us-cardiac:
    modality: "US"
    series_count: 2
    image_count: 15
    anatomical_region: "heart"
    study_description: "Cardiac Ultrasound"
    patient_name: "CARDIAC^PATIENT"
    patient_id: "CARDIAC001"
    accession_number: "ACC123456"
```

## Custom Tags Implementation

### Extended Template Configuration
To support custom DICOM tags, we need to extend the template system:

```yaml
study_templates:
  custom-ct-brain:
    modality: "CT"
    series_count: 2
    image_count: 30
    anatomical_region: "brain"
    study_description: "Advanced CT Brain Study"
    
    # Standard DICOM fields
    patient_name: "BRAIN^STUDY^PATIENT"
    patient_id: "BRAIN001"
    accession_number: "ACC789012"
    
    # Custom DICOM tags
    custom_tags:
      # Patient-level custom tags
      patient:
        "(0010,1010)": "45Y"                    # Patient Age
        "(0010,1020)": "180.5"                  # Patient Size (cm)
        "(0010,1030)": "75.0"                   # Patient Weight (kg)
        "(0010,2160)": "SMOKER"                 # Patient Medical Record Locator
        "(0010,2180)": "HYPERTENSION"           # Patient Occupation
      
      # Study-level custom tags
      study:
        "(0008,103E)": "BRAIN_CT_PROTOCOL"      # Series Description
        "(0008,1080)": "NEUROLOGY"              # Admitting Diagnoses Description
        "(0018,0015)": "BRAIN"                  # Body Part Examined
        "(0018,1030)": "HEAD_FIRST_SUPINE"      # Protocol Name
        "(0020,0010)": "STUDY123"               # Study ID
      
      # Series-level custom tags
      series:
        "(0018,0010)": "IV_CONTRAST"            # Contrast/Bolus Agent
        "(0018,0012)": "IOPAMIDOL"              # Contrast/Bolus Agent Sequence
        "(0018,0014)": "RADIAL"                 # Contrast/Bolus Route
        "(0018,1040)": "120"                    # Contrast/Bolus Volume
        "(0018,1041)": "ML"                     # Contrast/Bolus Total Dose
        "(0018,1048)": "2.5"                    # Contrast/Bolus Flow Rate
        "(0018,1049)": "ML/S"                   # Contrast/Bolus Flow Rate Units
        "(0018,1050)": "75"                     # Contrast/Bolus Start Time
        "(0018,1072)": "ML"                     # Contrast/Bolus Stop Time
      
      # Equipment-level custom tags
      equipment:
        "(0008,0070)": "ACME_MEDICAL"           # Manufacturer
        "(0008,1090)": "CT_SCANNER_PRO_2024"    # Manufacturer's Model Name
        "(0018,1000)": "SN123456789"            # Device Serial Number
        "(0018,1020)": "v2.1.4"                 # Software Version(s)
      
      # Institution-level custom tags
      institution:
        "(0008,0080)": "ACME_HOSPITAL"          # Institution Name
        "(0008,0081)": "123 MAIN ST"            # Institution Address
        "(0008,1010)": "CT_STATION_01"          # Station Name
        "(0008,1040)": "RADIOLOGY_DEPT"         # Institutional Department Name
        "(0008,1048)": "DR_SMITH"               # Physician(s) of Record
        "(0008,1049)": "DR_JONES"               # Physician(s) Reading Study
        "(0008,1050)": "TECH_JOHNSON"           # Performing Physician's Name
        "(0008,1060)": "RADIOLOGY"              # Name of Physician(s) Reading Study
```

### Using Custom Tags Template
```bash
# Create study with custom tags
./bin/crgodicom create --template custom-ct-brain

# Override specific custom tags via CLI
./bin/crgodicom create \
  --template custom-ct-brain \
  --custom-tag "(0010,1010)" "50Y" \
  --custom-tag "(0018,0010)" "ORAL_CONTRAST"
```

## Examples

### Example 1: Cardiac MRI Template
```yaml
study_templates:
  cardiac-mri:
    modality: "MR"
    series_count: 4
    image_count: 20
    anatomical_region: "heart"
    study_description: "Cardiac MRI with Contrast"
    
    custom_tags:
      patient:
        "(0010,1010)": "55Y"
        "(0010,1020)": "175.0"
        "(0010,1030)": "80.0"
        "(0010,2160)": "CARDIAC_PATIENT"
      
      study:
        "(0008,103E)": "CARDIAC_MRI_PROTOCOL"
        "(0018,0015)": "HEART"
        "(0018,1030)": "SUPINE_FEET_FIRST"
      
      series:
        "(0018,0010)": "GADOLINIUM"
        "(0018,0012)": "MULTIHANCE"
        "(0018,1040)": "15"
        "(0018,1041)": "ML"
        "(0018,0020)": "CINE"                   # Scanning Sequence
        "(0018,0021)": "SE"                     # Sequence Variant
        "(0018,0022)": "SP"                     # Scan Options
        "(0018,0023)": "AXIAL"                  # MR Acquisition Type
        "(0018,0024)": "2D"                     # Sequence Name
        "(0018,0025)": "CINE"                   # Angio Flag
      
      equipment:
        "(0008,0070)": "SIEMENS"
        "(0008,1090)": "MAGNETOM_SKYRA"
        "(0018,1000)": "SN987654321"
        "(0018,1020)": "syngo_MR_VE11C"
```

### Example 2: Mammography Template
```yaml
study_templates:
  screening-mammography:
    modality: "MG"
    series_count: 4
    image_count: 1
    anatomical_region: "breast"
    study_description: "Screening Mammography"
    
    custom_tags:
      patient:
        "(0010,1010)": "45Y"
        "(0010,1020)": "165.0"
        "(0010,1030)": "65.0"
        "(0010,2160)": "SCREENING_PATIENT"
      
      study:
        "(0008,103E)": "SCREENING_MAMMOGRAPHY"
        "(0018,0015)": "BREAST"
        "(0018,1030)": "STANDARD_PROTOCOL"
        "(0020,0010)": "MAMMO001"
      
      series:
        "(0018,0014)": "BILATERAL"
        "(0018,1508)": "LEFT_CC"                # Position Reference Indicator
        "(0018,150A)": "RIGHT_CC"               # Position Reference Indicator
        "(0018,150C)": "LEFT_MLO"               # Position Reference Indicator
        "(0018,150E)": "RIGHT_MLO"              # Position Reference Indicator
        "(0018,11A2)": "25"                     # Body Part Thickness
        "(0018,11A4)": "MM"                     # Body Part Thickness Units
        "(0018,7060)": "STANDARD"               # Frame of Reference Type
      
      equipment:
        "(0008,0070)": "HOLOGIC"
        "(0008,1090)": "SELENIA_DIMENSIONS"
        "(0018,1000)": "SN111222333"
        "(0018,1020)": "v8.0.1"
      
      institution:
        "(0008,0080)": "WOMENS_IMAGING_CENTER"
        "(0008,1010)": "MG_STATION_01"
        "(0008,1040)": "MAMMOGRAPHY"
        "(0008,1048)": "DR_BREAST_SPECIALIST"
        "(0008,1060)": "MAMMOGRAPHY_TECH"
```

### Example 3: Emergency CT Template
```yaml
study_templates:
  emergency-ct-trauma:
    modality: "CT"
    series_count: 3
    image_count: 40
    anatomical_region: "abdomen"
    study_description: "Emergency CT Trauma Protocol"
    
    custom_tags:
      patient:
        "(0010,1010)": "35Y"
        "(0010,2160)": "TRAUMA_PATIENT"
        "(0010,2180)": "EMERGENCY"
      
      study:
        "(0008,103E)": "TRAUMA_CT_PROTOCOL"
        "(0018,0015)": "ABDOMEN_PELVIS"
        "(0018,1030)": "EMERGENCY_PROTOCOL"
        "(0008,1080)": "TRAUMA_RULE_OUT"
        "(0020,0010)": "TRAUMA001"
      
      series:
        "(0018,0010)": "IV_CONTRAST"
        "(0018,0012)": "IOPAMIDOL"
        "(0018,0014)": "IV"
        "(0018,1040)": "100"
        "(0018,1041)": "ML"
        "(0018,1048)": "3.0"
        "(0018,1049)": "ML/S"
        "(0018,0020)": "HELICAL"                # Scanning Sequence
        "(0018,0021)": "SE"                     # Sequence Variant
        "(0018,0022)": "FAST"                   # Scan Options
        "(0018,0060)": "120"                    # KVP
        "(0018,1151)": "120"                    # X-Ray Tube Current
        "(0018,1152)": "MA"                     # X-Ray Tube Current Units
      
      equipment:
        "(0008,0070)": "GE_HEALTHCARE"
        "(0008,1090)": "REVOLUTION_CT"
        "(0018,1000)": "SN444555666"
        "(0018,1020)": "v5.2.1"
      
      institution:
        "(0008,0080)": "EMERGENCY_HOSPITAL"
        "(0008,1010)": "CT_EMERGENCY"
        "(0008,1040)": "EMERGENCY_RADIOLOGY"
        "(0008,1048)": "DR_TRAUMA_SPECIALIST"
        "(0008,1060)": "EMERGENCY_TECH"
```

## Advanced Usage

### Dynamic Custom Tags
You can use variables in custom tags:

```yaml
study_templates:
  dynamic-ct:
    modality: "CT"
    series_count: 2
    image_count: 25
    anatomical_region: "chest"
    study_description: "Dynamic CT Study"
    
    custom_tags:
      patient:
        "(0010,1010)": "${PATIENT_AGE}Y"       # Will be replaced with actual age
        "(0010,1020)": "${PATIENT_HEIGHT}"      # Will be replaced with actual height
        "(0010,1030)": "${PATIENT_WEIGHT}"      # Will be replaced with actual weight
      
      study:
        "(0008,103E)": "CT_${TIMESTAMP}"        # Will include generation timestamp
        "(0020,0010)": "STUDY_${STUDY_COUNT}"   # Will include study counter
```

### Template Inheritance
Create base templates and extend them:

```yaml
# Base CT template
study_templates:
  base-ct:
    modality: "CT"
    series_count: 2
    image_count: 25
    
    custom_tags:
      equipment:
        "(0008,0070)": "ACME_MEDICAL"
        "(0008,1090)": "CT_SCANNER_PRO_2024"
        "(0018,1000)": "SN123456789"
        "(0018,1020)": "v2.1.4"
      
      institution:
        "(0008,0080)": "ACME_HOSPITAL"
        "(0008,1010)": "CT_STATION_01"
        "(0008,1040)": "RADIOLOGY_DEPT"
  
  # Extended chest CT template
  chest-ct-extended:
    extends: "base-ct"                          # Inherit from base template
    anatomical_region: "chest"
    study_description: "Enhanced Chest CT"
    
    custom_tags:
      # Override or add to inherited tags
      study:
        "(0008,103E)": "CHEST_CT_ENHANCED"
        "(0018,0015)": "CHEST"
      
      series:
        "(0018,0010)": "IV_CONTRAST"
        "(0018,1040)": "100"
        "(0018,1041)": "ML"
```

### Command Line Custom Tags
Override template custom tags via CLI:

```bash
# Use template but override specific custom tags
./bin/crgodicom create \
  --template custom-ct-brain \
  --custom-tag "(0010,1010)" "60Y" \
  --custom-tag "(0018,0010)" "ORAL_CONTRAST" \
  --custom-tag "(0008,0080)" "DIFFERENT_HOSPITAL"

# Add new custom tags not in template
./bin/crgodicom create \
  --template base-ct \
  --custom-tag "(0018,0015)" "SPINE" \
  --custom-tag "(0008,103E)" "SPINE_CT_PROTOCOL" \
  --custom-tag "(0018,0010)" "IV_CONTRAST"
```

## Implementation Notes

### DICOM Tag Format
- Use standard DICOM tag format: `(GGGG,EEEE)`
- Group and Element are 4-digit hexadecimal
- Example: `(0010,0010)` for Patient Name
- Example: `(0018,0015)` for Body Part Examined

### Value Representations (VR)
The system automatically determines the correct VR based on the tag:
- String values: PN, LO, SH, CS, ST, LT, UT
- Numeric values: IS, DS, FL, FD, SS, US, SL, UL
- Date values: DA, DT, TM
- Binary values: OB, OW, OF, OD, OL, OV

### Validation
- All custom tags are validated against DICOM standards
- Invalid tags are logged as warnings
- Critical tags (required for DICOM compliance) cannot be overridden

### Performance
- Custom tags are processed during study generation
- Large numbers of custom tags may impact generation time
- Consider using templates for frequently used tag sets

## Best Practices

1. **Organize by Category**: Group custom tags by patient, study, series, equipment, etc.
2. **Use Meaningful Names**: Choose template names that describe the use case
3. **Document Custom Tags**: Include comments explaining non-standard tags
4. **Validate Templates**: Test templates before deploying to production
5. **Version Control**: Keep templates in version control for consistency
6. **Inheritance**: Use base templates to reduce duplication
7. **Standards Compliance**: Ensure custom tags follow DICOM standards

## Troubleshooting

### Common Issues
1. **Invalid Tag Format**: Ensure tags use `(GGGG,EEEE)` format
2. **Missing VR**: System auto-detects VR, but some tags may need explicit VR
3. **Template Not Found**: Check template name spelling and configuration file
4. **Custom Tag Conflicts**: Later values override earlier ones

### Debugging
```bash
# Enable debug logging
./bin/crgodicom --log-level DEBUG create --template my-template

# Validate template syntax
./bin/crgodicom validate-template --template-file my-template.yaml

# List all available tags
./bin/crgodicom list-dicom-tags
```

This guide provides a comprehensive foundation for using custom DICOM tags in crgodicom templates. The system is designed to be flexible while maintaining DICOM compliance.

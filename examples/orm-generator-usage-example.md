# ORM Template Generator Usage Examples

This document shows how to use the ORM template generator to automatically create DICOM templates from various input sources.

## 🎯 Overview

The ORM template generator can create DICOM templates from:
- **HL7 ORM messages** - Hospital order management messages
- **Go struct definitions** - Go ORM models with DICOM tags
- **SQL schema files** - Database table definitions
- **JSON schema** - API or database schema definitions

## 🏥 Example 1: HL7 ORM Message to DICOM Template

### Input: HL7 ORM Message (TESTRIS System)
```hl7
MSH|^~\&|TESTRIS|TESTRIS|EXT_DEF||20241215143022+1100||ORM^O01^|abc123xyz-def|P|2.4
PID|||200000001^^^2006630&&GUID^005~7777777^^^E335&&GUID^017^E335||JOHNSON^SARAH^MARIE^""^MS||19820623|F|||MAPLE STREET 789^""^MELBOURNE^VIC^3001^1201^home^""~CANADA^""^""^""^""^""^Birth^""|||||||4444444||||3|||||||""
PV1||I|CNR^^^E304^^^^^Melbourne General Hospital|||||789012CD^ATTENDINGDR||||||||||||MC
ORC|XO|4339239594|2024WS0000001|2024WS0000001|E|||||||789012CD^ATTENDINGDR^^^^Dr
OBR|1|4444444444|2024WS0000001-1|MRIBRAINCON^MRI Brain with Contrast^WS-MGH.ORDERABLES|||||||||Research Acc: Y||^^^Neurological, Brain, Head, Neck|||2024WS0000001-1||||MR|||^^^20241215140500+1100^^Routine||||^Clinical History: Progressive headaches with visual disturbances||SRV-VICG-EXT-DEF@vichealth.net||||||||||||MRIBRAINCON^MRI Brain with Contrast^WS-MGH.PROCEDURES
```

### Command: Generate Template
```bash
# Generate DICOM template from HL7 ORM message
crgodicom orm-generate \
  --input examples/hl7-orm-example.txt \
  --output templates/testris-mri-brain-template.yaml \
  --template-name "testris-mri-brain-contrast" \
  --modality "MR" \
  --series-count 3 \
  --image-count 25 \
  --verbose

# Output:
# Successfully generated DICOM template: templates/testris-mri-brain-template.yaml
# 
# 📋 Generated Template Summary:
# • Template Name: testris-mri-brain-contrast
# • Modality: MR
# • Series Count: 3
# • Image Count: 25
# • Custom Tag Categories: 6
#   - patient: 8 tags
#   - study: 10 tags
#   - series: 11 tags
#   - equipment: 4 tags
#   - institution: 6 tags
#   - procedure: 8 tags
#   - clinical: 7 tags
```

### Generated Template Structure
```yaml
study_templates:
  testris-mri-brain-contrast:
    modality: "MR"
    series_count: 3
    image_count: 25
    anatomical_region: "brain"
    study_description: "MRI Brain with Contrast"
    
    custom_tags:
      patient:
        "(0010,0010)": "JOHNSON^SARAH^MARIE^MS"     # Patient Name
        "(0010,0020)": "200000001"                  # Patient ID
        "(0010,0030)": "19820623"                   # Patient Birth Date
        "(0010,0040)": "F"                          # Patient Sex
        "(0010,1040)": "MAPLE STREET 789^MELBOURNE^VIC^3001" # Patient Address
        
      study:
        "(0008,0020)": "20241215"                   # Study Date
        "(0008,0030)": "140500.000000"              # Study Time
        "(0008,0050)": "2024WS0000001"              # Accession Number
        "(0008,1030)": "MRI Brain with Contrast"    # Study Description
        "(0008,0090)": "DR^ATTENDING^PHYSICIAN"     # Referring Physician Name
        
      # ... additional categories
```

### Usage: Generate Study from Template
```bash
# Use the generated template to create a DICOM study
crgodicom create --template testris-mri-brain-contrast

# Customize patient data while using the template structure
crgodicom create --template testris-mri-brain-contrast \
  --patient-name "SMITH^JOHN^MICHAEL" \
  --patient-id "200000002" \
  --accession-number "2024WS0000002"

# Export the study
crgodicom export --study-id <study-id> --format pdf --output-file testris-mri-study.pdf
```

## 🔧 Example 2: Go Struct to DICOM Template

### Input: Go Struct with DICOM Tags
```go
// patient_model.go
package models

import "time"

type Patient struct {
    ID           uint      `gorm:"primaryKey" dicom:"(0010,0020)" json:"id"`
    MRN          string    `gorm:"uniqueIndex" dicom:"(0010,2160)" json:"mrn"`
    LastName     string    `gorm:"size:255" dicom:"(0010,0010):family" json:"last_name"`
    FirstName    string    `gorm:"size:255" dicom:"(0010,0010):given" json:"first_name"`
    MiddleName   string    `gorm:"size:255" dicom:"(0010,0010):middle" json:"middle_name"`
    DateOfBirth  time.Time `dicom:"(0010,0030)" transform:"date_format:20060102" json:"date_of_birth"`
    Sex          string    `gorm:"size:1" dicom:"(0010,0040)" validate:"enum:M,F,O" json:"sex"`
    PhoneNumber  string    `gorm:"size:20" dicom:"(0010,2154)" json:"phone_number"`
    Address      string    `gorm:"size:500" dicom:"(0010,1040)" json:"address"`
}

type Study struct {
    ID              uint      `gorm:"primaryKey" json:"id"`
    StudyInstanceUID string   `gorm:"uniqueIndex" dicom:"(0020,000D)" json:"study_instance_uid"`
    StudyDate       time.Time `dicom:"(0008,0020)" transform:"date_format:20060102" json:"study_date"`
    StudyTime       time.Time `dicom:"(0008,0030)" transform:"time_format:150405.000000" json:"study_time"`
    AccessionNumber string    `gorm:"uniqueIndex" dicom:"(0008,0050)" json:"accession_number"`
    StudyDescription string   `gorm:"size:255" dicom:"(0008,1030)" json:"study_description"`
    Modality        string    `gorm:"size:10" dicom:"(0008,0060)" json:"modality"`
    PatientID       uint      `gorm:"foreignKey" json:"patient_id"`
    Patient         Patient   `gorm:"references:ID" json:"patient"`
}
```

### Command: Generate from Go Structs
```bash
# Generate DICOM template from Go struct definitions
crgodicom orm-generate \
  --input models/patient_model.go \
  --type go \
  --output templates/go-model-template.yaml \
  --template-name "go-patient-study" \
  --modality "CT" \
  --verbose
```

## 📊 Example 3: SQL Schema to DICOM Template

### Input: SQL Schema
```sql
-- hospital_schema.sql
CREATE TABLE patients (
    id SERIAL PRIMARY KEY,
    mrn VARCHAR(50) UNIQUE NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    first_name VARCHAR(255) NOT NULL,
    date_of_birth DATE NOT NULL,
    sex CHAR(1) CHECK (sex IN ('M', 'F', 'O')),
    phone_number VARCHAR(20),
    address TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE studies (
    id SERIAL PRIMARY KEY,
    study_instance_uid VARCHAR(255) UNIQUE NOT NULL,
    study_date DATE NOT NULL,
    study_time TIME NOT NULL,
    accession_number VARCHAR(50) UNIQUE NOT NULL,
    study_description VARCHAR(255),
    modality VARCHAR(10) NOT NULL,
    patient_id INTEGER REFERENCES patients(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE series (
    id SERIAL PRIMARY KEY,
    series_instance_uid VARCHAR(255) UNIQUE NOT NULL,
    series_number INTEGER NOT NULL,
    series_description VARCHAR(255),
    modality VARCHAR(10) NOT NULL,
    body_part_examined VARCHAR(50),
    study_id INTEGER REFERENCES studies(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

### Command: Generate from SQL
```bash
# Generate DICOM template from SQL schema
crgodicom orm-generate \
  --input schema/hospital_schema.sql \
  --type sql \
  --output templates/sql-schema-template.yaml \
  --template-name "hospital-database-template" \
  --modality "CT" \
  --series-count 2 \
  --image-count 50
```

## 🌐 Example 4: JSON Schema to DICOM Template

### Input: JSON Schema
```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "Medical Imaging API Schema",
  "type": "object",
  "properties": {
    "patient": {
      "type": "object",
      "properties": {
        "patient_id": {"type": "string", "dicom_tag": "(0010,0020)"},
        "full_name": {"type": "string", "dicom_tag": "(0010,0010)"},
        "birth_date": {"type": "string", "format": "date", "dicom_tag": "(0010,0030)"},
        "gender": {"type": "string", "enum": ["M", "F", "O"], "dicom_tag": "(0010,0040)"}
      }
    },
    "study": {
      "type": "object", 
      "properties": {
        "study_uid": {"type": "string", "dicom_tag": "(0020,000D)"},
        "study_date": {"type": "string", "format": "date", "dicom_tag": "(0008,0020)"},
        "modality": {"type": "string", "dicom_tag": "(0008,0060)"},
        "description": {"type": "string", "dicom_tag": "(0008,1030)"}
      }
    }
  }
}
```

### Command: Generate from JSON Schema
```bash
# Generate DICOM template from JSON schema
crgodicom orm-generate \
  --input api/medical-imaging-schema.json \
  --type json \
  --output templates/api-schema-template.yaml \
  --template-name "medical-imaging-api"
```

## 🔄 Workflow Integration

### 1. Automated Template Generation
```bash
#!/bin/bash
# generate-templates.sh - Automated template generation script

echo "🏥 Generating DICOM templates from hospital systems..."

# Generate from HL7 ORM messages
for hl7_file in data/hl7/*.hl7; do
    template_name=$(basename "$hl7_file" .hl7)
    crgodicom orm-generate \
        --input "$hl7_file" \
        --output "templates/${template_name}-template.yaml" \
        --template-name "$template_name"
done

# Generate from Go models
for go_file in models/*.go; do
    template_name=$(basename "$go_file" .go)
    crgodicom orm-generate \
        --input "$go_file" \
        --type go \
        --output "templates/${template_name}-template.yaml" \
        --template-name "$template_name"
done

echo "✅ Template generation complete!"
```

### 2. CI/CD Integration
```yaml
# .github/workflows/template-generation.yml
name: Generate DICOM Templates

on:
  push:
    paths:
      - 'models/**'
      - 'schema/**'
      - 'data/hl7/**'

jobs:
  generate-templates:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build crgodicom
        run: make build
        
      - name: Generate templates from models
        run: |
          mkdir -p generated-templates
          
          # Generate from HL7 messages
          for file in data/hl7/*.hl7; do
            if [ -f "$file" ]; then
              name=$(basename "$file" .hl7)
              ./bin/crgodicom orm-generate \
                --input "$file" \
                --output "generated-templates/${name}.yaml" \
                --template-name "$name"
            fi
          done
          
          # Generate from Go models
          for file in models/*.go; do
            if [ -f "$file" ]; then
              name=$(basename "$file" .go)
              ./bin/crgodicom orm-generate \
                --input "$file" \
                --type go \
                --output "generated-templates/${name}.yaml" \
                --template-name "$name"
            fi
          done
      
      - name: Upload generated templates
        uses: actions/upload-artifact@v4
        with:
          name: generated-dicom-templates
          path: generated-templates/
```

## 🎯 Advanced Usage

### Custom Field Mappings
```bash
# Generate with custom field mappings
crgodicom orm-generate \
  --input patient-data.hl7 \
  --output custom-template.yaml \
  --template-name "custom-patient-template" \
  --field-mapping "patient_id:(0010,0020)" \
  --field-mapping "mrn:(0010,2160)" \
  --field-mapping "study_date:(0008,0020)"
```

### Batch Processing
```bash
# Process multiple HL7 files
find data/hl7/ -name "*.hl7" -exec crgodicom orm-generate \
  --input {} \
  --output templates/{}.yaml \
  --template-name $(basename {} .hl7) \;

# Generate templates for different modalities
for modality in CT MR US CR DX; do
  crgodicom orm-generate \
    --input models/patient.go \
    --type go \
    --output "templates/${modality,,}-patient-template.yaml" \
    --template-name "${modality,,}-patient-study" \
    --modality "$modality"
done
```

## 🔍 Validation and Testing

### Validate Generated Templates
```bash
# Generate and immediately test the template
crgodicom orm-generate --input data.hl7 --output test-template.yaml
crgodicom create --template test-template --study-count 1
crgodicom export --study-id <study-id> --format pdf --output-file validation.pdf
```

### Template Quality Check
```bash
# Generate with verbose output to check quality
crgodicom orm-generate \
  --input complex-hl7-message.hl7 \
  --output complex-template.yaml \
  --verbose

# Check the generated template
cat complex-template.yaml | grep -E "(completeness|confidence|manual_review)"
```

## 📚 Integration Examples

### Hospital Information System (HIS)
```bash
# Generate templates from HIS HL7 exports
crgodicom orm-generate \
  --input his-exports/radiology-orders.hl7 \
  --output templates/his-radiology-template.yaml \
  --template-name "his-radiology-study"
```

### Research Database
```bash
# Generate from research database schema
crgodicom orm-generate \
  --input research-db-schema.sql \
  --type sql \
  --output templates/research-template.yaml \
  --template-name "research-study-protocol"
```

### Electronic Health Record (EHR)
```bash
# Generate from EHR API schema
crgodicom orm-generate \
  --input ehr-api-schema.json \
  --type json \
  --output templates/ehr-template.yaml \
  --template-name "ehr-imaging-study"
```

## 🎯 Benefits

### For Developers
- **Rapid Integration** - Quickly create DICOM templates from existing systems
- **Consistency** - Ensure consistent DICOM tag usage across applications
- **Automation** - Reduce manual template creation work

### For Hospitals
- **Standards Compliance** - Generate templates that match existing data standards
- **Data Integration** - Seamlessly integrate with HIS, PACS, and EHR systems
- **Quality Assurance** - Automated validation and quality metrics

### For Researchers
- **Protocol Compliance** - Generate templates that match research protocols
- **Data Standardization** - Ensure consistent metadata across studies
- **Workflow Automation** - Automate template creation from database schemas

## 🔧 Configuration

### Custom Mapping Rules
Create a `orm-mapping.yaml` file to customize field mappings:

```yaml
# orm-mapping.yaml
field_mappings:
  # Patient mappings
  patient_id: "(0010,0020)"
  medical_record_number: "(0010,2160)"
  full_name: "(0010,0010)"
  birth_date: "(0010,0030)"
  
  # Study mappings
  accession_number: "(0008,0050)"
  study_description: "(0008,1030)"
  modality: "(0008,0060)"
  
transformations:
  date_format: "20060102"           # DICOM DA format
  time_format: "150405.000000"      # DICOM TM format
  name_format: "LAST^FIRST^MIDDLE"  # DICOM PN format

validation_rules:
  required_fields:
    - "patient_id"
    - "accession_number"
    - "study_date"
  
  field_constraints:
    sex: ["M", "F", "O"]
    modality: ["CT", "MR", "US", "CR", "DX", "MG"]
```

## 📞 Support

- **Documentation**: [ORM Template Generator Design](../docs/ORM_TEMPLATE_GENERATOR.md)
- **Examples**: [Template Examples](../docs/template-examples/README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This feature enables seamless integration between existing healthcare data systems and DICOM study generation*

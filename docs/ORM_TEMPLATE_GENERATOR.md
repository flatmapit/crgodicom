# ORM Template Generator Design

This document outlines the design for automatically generating DICOM study templates from ORM (Object-Relational Mapping) models and database schemas.

## 🎯 Overview

The ORM Template Generator will allow users to automatically create DICOM study templates by analyzing:
- Database schemas
- ORM model definitions
- Struct definitions
- JSON/YAML schema files

This enables seamless integration between existing data models and DICOM study generation.

## 🏗️ Architecture

### Input Sources
1. **Go Structs** - Parse Go struct definitions with DICOM tags
2. **SQL Schema** - Analyze database table structures
3. **JSON Schema** - Parse JSON schema definitions
4. **YAML Models** - Parse YAML model definitions
5. **Database Introspection** - Connect to live databases and extract schema

### Output
- **DICOM Template YAML** - Generated template with custom DICOM tags
- **Mapping Configuration** - Field mapping rules and transformations
- **Validation Rules** - Data validation and constraints

## 📋 Supported ORM Frameworks

### Phase 1: Go Ecosystem
- **GORM** - Most popular Go ORM
- **Go Structs** - Native Go struct parsing
- **SQL Schema** - Direct SQL DDL parsing

### Phase 2: Multi-Language Support
- **Python**: SQLAlchemy, Django ORM
- **Java**: Hibernate annotations
- **C#**: Entity Framework models

## 🔧 Implementation Plan

### 1. Go Struct Parser
```go
type Patient struct {
    ID           uint   `gorm:"primaryKey" dicom:"(0010,0020)"`
    Name         string `gorm:"size:255" dicom:"(0010,0010)"`
    BirthDate    time.Time `dicom:"(0010,0030)"`
    Sex          string `gorm:"size:1" dicom:"(0010,0040)"`
    MedicalID    string `gorm:"uniqueIndex" dicom:"(0010,2160)"`
}

type Study struct {
    ID              uint      `gorm:"primaryKey"`
    StudyInstanceUID string   `gorm:"uniqueIndex" dicom:"(0020,000D)"`
    StudyDate       time.Time `dicom:"(0008,0020)"`
    StudyTime       time.Time `dicom:"(0008,0030)"`
    Modality        string    `dicom:"(0008,0060)"`
    PatientID       uint      `gorm:"foreignKey"`
    Patient         Patient   `gorm:"references:ID"`
}
```

### 2. Template Generation Logic
```yaml
# Generated template from Go structs above
study_templates:
  patient-study-orm:
    modality: "{{ .Study.Modality }}"
    series_count: 1
    image_count: 10
    patient_name: "{{ .Patient.Name }}"
    patient_id: "{{ .Patient.ID }}"
    
    custom_tags:
      patient:
        "(0010,0020)": "{{ .Patient.ID }}"
        "(0010,0010)": "{{ .Patient.Name }}"
        "(0010,0030)": "{{ .Patient.BirthDate.Format \"20060102\" }}"
        "(0010,0040)": "{{ .Patient.Sex }}"
        "(0010,2160)": "{{ .Patient.MedicalID }}"
      
      study:
        "(0020,000D)": "{{ .Study.StudyInstanceUID }}"
        "(0008,0020)": "{{ .Study.StudyDate.Format \"20060102\" }}"
        "(0008,0030)": "{{ .Study.StudyTime.Format \"150405.000000\" }}"
        "(0008,0060)": "{{ .Study.Modality }}"
```

### 3. CLI Interface
```bash
# Generate template from Go struct file
crgodicom orm-generate --input models/patient.go --output templates/patient-template.yaml

# Generate from SQL schema
crgodicom orm-generate --input schema.sql --type sql --output templates/schema-template.yaml

# Generate from JSON schema
crgodicom orm-generate --input patient-schema.json --type json --output templates/json-template.yaml

# Generate from database connection
crgodicom orm-generate --database "postgres://user:pass@localhost/db" --table patients --output templates/db-template.yaml
```

## 🔄 Mapping Rules

### Field Type Mapping
| ORM Type | DICOM VR | DICOM Tag Category | Example |
|----------|----------|-------------------|---------|
| `string` | LO, SH, PN | Various | Patient Name, Study Description |
| `time.Time` | DA, TM, DT | Date/Time | Study Date, Birth Date |
| `uint`, `int` | IS, US | Numeric | Patient ID, Series Number |
| `float64` | DS | Decimal | Patient Weight, Height |
| `bool` | CS | Code String | Pregnancy Status |
| `[]byte` | OB, OW | Binary | Image Data |

### Tag Category Assignment
- **Patient fields** → (0010,xxxx) tags
- **Study fields** → (0008,xxxx), (0020,xxxx) tags
- **Series fields** → (0018,xxxx), (0020,xxxx) tags
- **Equipment fields** → (0008,xxxx), (0018,xxxx) tags
- **Custom fields** → (7FE1-7FE4,xxxx) tags

## 📝 Configuration

### Mapping Configuration File
```yaml
# orm-mapping.yaml
field_mappings:
  patient:
    id: "(0010,0020)"           # Patient ID
    name: "(0010,0010)"         # Patient Name
    birth_date: "(0010,0030)"   # Patient Birth Date
    sex: "(0010,0040)"          # Patient Sex
    
  study:
    study_instance_uid: "(0020,000D)"  # Study Instance UID
    study_date: "(0008,0020)"          # Study Date
    study_time: "(0008,0030)"          # Study Time
    modality: "(0008,0060)"            # Modality
    
  series:
    series_instance_uid: "(0020,000E)" # Series Instance UID
    series_number: "(0020,0011)"       # Series Number
    series_description: "(0008,103E)"  # Series Description

type_transformations:
  time.Time:
    date_format: "20060102"           # YYYYMMDD
    time_format: "150405.000000"      # HHMMSS.ffffff
  
  string:
    max_length: 64                    # DICOM string length limits
    encoding: "ISO_IR 100"            # Character set
    
  numeric:
    range_validation: true            # Validate numeric ranges
```

## 🧪 Example Use Cases

### 1. Hospital Information System Integration
```go
// HIS Patient Model
type HISPatient struct {
    MRN          string    `gorm:"primaryKey" dicom:"(0010,0020)"`
    LastName     string    `dicom:"(0010,0010):family"`
    FirstName    string    `dicom:"(0010,0010):given"`
    DOB          time.Time `dicom:"(0010,0030)"`
    Gender       string    `dicom:"(0010,0040)"`
    Insurance    string    `dicom:"(0010,2160)"`
}

// Generated Template
# Generated from HISPatient model
study_templates:
  his-patient-template:
    custom_tags:
      patient:
        "(0010,0020)": "{{ .MRN }}"
        "(0010,0010)": "{{ .LastName }}^{{ .FirstName }}"
        "(0010,0030)": "{{ .DOB.Format \"20060102\" }}"
        "(0010,0040)": "{{ .Gender }}"
        "(0010,2160)": "{{ .Insurance }}"
```

### 2. Research Database Integration
```go
// Research Subject Model
type ResearchSubject struct {
    SubjectID    string `dicom:"(0010,0020)"`
    StudyArm     string `dicom:"(0012,0040)"` # Clinical Trial Subject ID
    ProtocolID   string `dicom:"(0012,0020)"` # Clinical Trial Protocol ID
    TimePoint    string `dicom:"(0012,0050)"` # Clinical Trial Time Point ID
    Modality     string `dicom:"(0008,0060)"`
    BodyPart     string `dicom:"(0018,0015)"`
}
```

### 3. Equipment Database Integration
```go
// Equipment Model
type Equipment struct {
    SerialNumber string `dicom:"(0018,1000)"` # Device Serial Number
    Manufacturer string `dicom:"(0008,0070)"` # Manufacturer
    ModelName    string `dicom:"(0008,1090)"` # Manufacturer's Model Name
    SoftwareVer  string `dicom:"(0018,1020)"` # Software Version(s)
    StationName  string `dicom:"(0008,1010)"` # Station Name
}
```

## 🔧 Implementation Components

### 1. Parser Package (`internal/orm/parser/`)
- **struct_parser.go** - Parse Go struct definitions
- **sql_parser.go** - Parse SQL DDL statements
- **json_parser.go** - Parse JSON schema
- **yaml_parser.go** - Parse YAML model definitions

### 2. Generator Package (`internal/orm/generator/`)
- **template_generator.go** - Generate DICOM templates
- **tag_mapper.go** - Map fields to DICOM tags
- **validator.go** - Validate generated templates

### 3. CLI Command (`internal/cli/orm.go`)
- **ORM Generate Command** - Main CLI interface
- **Configuration Management** - Handle mapping rules
- **Output Formatting** - Generate YAML templates

### 4. Configuration (`internal/orm/config/`)
- **mapping_config.go** - Field mapping configuration
- **validation_rules.go** - Validation and constraints
- **default_mappings.go** - Default DICOM tag mappings

## 🎯 Benefits

### For Developers
- **Rapid Template Creation** - Generate templates from existing models
- **Consistency** - Ensure consistent DICOM tag usage
- **Integration** - Seamless integration with existing systems

### For Researchers
- **Database Integration** - Use existing research databases
- **Protocol Compliance** - Ensure protocol-specific DICOM tags
- **Automation** - Reduce manual template creation

### for Institutions
- **Standardization** - Consistent DICOM metadata across systems
- **Compliance** - Ensure regulatory compliance
- **Efficiency** - Reduce manual configuration work

## 🚀 Future Enhancements

### Advanced Features
- **Database Introspection** - Connect to live databases
- **Multi-Table Relationships** - Handle complex model relationships
- **Custom Validators** - Add business logic validation
- **Template Inheritance** - Support template hierarchies

### Integration Features
- **API Integration** - REST/GraphQL API schema parsing
- **Version Control** - Track template changes
- **Migration Tools** - Update templates when models change
- **Testing Framework** - Validate generated templates

## 📚 Documentation Plan

### User Guides
- **Quick Start Guide** - Basic ORM template generation
- **Advanced Configuration** - Custom mapping rules
- **Framework Integration** - Specific ORM framework guides
- **Best Practices** - Recommended patterns and conventions

### Developer Documentation
- **API Reference** - Parser and generator APIs
- **Extension Guide** - Adding new ORM support
- **Testing Guide** - Testing ORM integrations
- **Contributing** - How to contribute new features

---

*This feature will significantly enhance CRGoDICOM's integration capabilities with existing data systems*

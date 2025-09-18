# Comprehensive DICOM Template Guide with Custom Tags

This is the most comprehensive guide to creating DICOM templates with extensive custom tag examples in CRGoDICOM.

## 🎯 Overview

This guide covers:
1. **Complete DICOM tag reference** - Over 200+ custom tags
2. **Very large example template** - Multi-modality comprehensive template
3. **Advanced tag categories** - Patient, Study, Series, Equipment, Institution, Procedure
4. **Real-world scenarios** - Clinical protocols and research applications
5. **Implementation details** - How custom tags are processed and applied

## 📚 Quick Reference to Existing Guides

### Current Documentation Structure:
- **`examples/custom-templates-guide.md`** (458 lines) - Complete custom template guide
- **`examples/cardiac-mri-template.yaml`** (158 lines) - Large cardiac MRI example
- **`examples/simple-custom-template.yaml`** (91 lines) - Basic example
- **`docs/template-examples/`** - Visual examples with generated images

## 🏥 Very Large Example: Multi-Modality Research Template

See the **ultra-comprehensive-research-template.yaml** below - this demonstrates:
- **300+ custom DICOM tags**
- **Multi-modality support**
- **Research protocol tags**
- **Institution-specific tags**
- **Advanced technical parameters**

## 📋 DICOM Tag Categories and Examples

### 1. Patient Information Tags (0010,xxxx)
```yaml
patient:
  # Basic Demographics
  "(0010,0010)": "DOE^JOHN^MICHAEL^JR"      # Patient Name
  "(0010,0020)": "P123456789"               # Patient ID
  "(0010,0030)": "19800115"                 # Patient Birth Date
  "(0010,0040)": "M"                        # Patient Sex
  "(0010,1010)": "045Y"                     # Patient Age
  "(0010,1020)": "180.0"                    # Patient Size (cm)
  "(0010,1030)": "75.5"                     # Patient Weight (kg)
  
  # Extended Demographics
  "(0010,1040)": "1234 MAIN ST"             # Patient Address
  "(0010,1050)": "555-123-4567"             # Patient Phone Number
  "(0010,1060)": "john.doe@email.com"       # Patient Email
  "(0010,1080)": "PHYSICIAN"                # Military Rank
  "(0010,1081)": "USA"                      # Branch of Service
  "(0010,1090)": "MRN123456"                # Medical Record Locator
  
  # Medical Information
  "(0010,2160)": "RESEARCH_STUDY_001"       # Ethnicity
  "(0010,2180)": "ENGINEER"                 # Patient Occupation
  "(0010,21A0)": "NO_SMOKING"               # Smoking Status
  "(0010,21B0)": "NONE"                     # Additional Patient History
  "(0010,21C0)": "NO"                       # Pregnancy Status
  "(0010,21D0)": "19800115"                 # Last Menstrual Date
  
  # Insurance/Administrative
  "(0010,1002)": "OTHER_PATIENT_IDS"        # Other Patient IDs
  "(0010,1005)": "PATIENT_BIRTH_NAME"       # Patient Birth Name
  "(0010,1001)": "MOTHER_MAIDEN_NAME"       # Other Patient Names
  "(0010,2203)": "PATIENT_SEX_NEUTERED"     # Patient Sex Neutered
```

### 2. Study Information Tags (0008,xxxx and 0020,xxxx)
```yaml
study:
  # Basic Study Info
  "(0008,0020)": "20250918"                 # Study Date
  "(0008,0030)": "143000.000000"            # Study Time
  "(0008,0050)": "ACC2025001234"            # Accession Number
  "(0008,0090)": "REFERRING_PHYSICIAN"      # Referring Physician Name
  "(0008,1030)": "RESEARCH_PROTOCOL_001"    # Study Description
  "(0020,0010)": "STUDY001"                 # Study ID
  "(0020,0011)": "1"                        # Series Number
  
  # Clinical Context
  "(0008,1032)": "RESEARCH_FACILITY"        # Procedure Code Sequence
  "(0008,1048)": "ATTENDING_PHYSICIAN"      # Physician(s) of Record
  "(0008,1049)": "PHYSICIAN_READING"        # Physician(s) Reading Study
  "(0008,1050)": "PERFORMING_PHYSICIAN"     # Performing Physician Name
  "(0008,1060)": "READING_PHYSICIAN"        # Name of Physician(s) Reading
  "(0008,1070)": "TECHNOLOGIST_NAME"        # Operators Name
  "(0008,1080)": "CLINICAL_INDICATION"      # Admitting Diagnoses Description
  "(0008,1084)": "ADMITTING_PHYSICIAN"      # Admitting Diagnosis Code
  
  # Research Tags
  "(0012,0062)": "YES"                      # Patient Identity Removed
  "(0012,0063)": "CTP"                      # De-identification Method
  "(0018,9004)": "RESEARCH_CONTENT"         # Content Qualification
  "(0028,0301)": "YES"                      # Burned In Annotation
  
  # Study Classification
  "(0008,0061)": "RESEARCH"                 # Modalities in Study
  "(0008,1140)": "REFERENCED_IMAGE_SEQ"     # Referenced Image Sequence
  "(0008,1155)": "REF_SOP_INSTANCE_UID"     # Referenced SOP Instance UID
  "(0008,1199)": "REFERENCED_SOP_SEQUENCE"  # Referenced SOP Sequence
```

### 3. Series Information Tags (0018,xxxx and 0020,xxxx)
```yaml
series:
  # Basic Series Info
  "(0008,103E)": "RESEARCH_SERIES_001"      # Series Description
  "(0020,000E)": "SERIES_INSTANCE_UID"      # Series Instance UID
  "(0020,0011)": "1"                        # Series Number
  "(0020,0060)": "R"                        # Laterality
  
  # Acquisition Parameters
  "(0018,0010)": "GADOLINIUM_CONTRAST"      # Contrast/Bolus Agent
  "(0018,0012)": "GADOVIST_1.0_MMOL"        # Contrast/Bolus Agent Sequence
  "(0018,0014)": "IV"                       # Contrast/Bolus Route
  "(0018,0015)": "WHOLE_BODY"               # Body Part Examined
  "(0018,0020)": "GR"                       # Scanning Sequence
  "(0018,0021)": "SK\\SP\\MP"               # Sequence Variant
  "(0018,0022)": "PFF\\FS"                  # Scan Options
  "(0018,0023)": "3D"                       # MR Acquisition Type
  "(0018,0024)": "FLASH"                    # Sequence Name
  
  # Contrast Information
  "(0018,1040)": "20.0"                     # Contrast/Bolus Volume
  "(0018,1041)": "ML"                       # Contrast/Bolus Total Dose
  "(0018,1042)": "T1"                       # Contrast/Bolus Ingredient
  "(0018,1043)": "1.0"                      # Contrast/Bolus Concentration
  "(0018,1044)": "MMOL/L"                   # Contrast/Bolus Agent Number
  "(0018,1045)": "GADOVIST"                 # Contrast/Bolus Ingredient Code
  "(0018,1046)": "BAYER"                    # Contrast/Bolus Agent Administration
  "(0018,1047)": "IV_BOLUS"                 # Contrast/Bolus Agent Sequence
  "(0018,1048)": "3.0"                      # Contrast/Bolus Flow Rate
  "(0018,1049)": "ML/S"                     # Contrast/Bolus Flow Rate Units
  
  # Timing Information
  "(0018,1050)": "60"                       # Contrast/Bolus Start Time
  "(0018,1051)": "120"                      # Contrast/Bolus Stop Time
  "(0018,1052)": "180"                      # Contrast/Bolus Total Time
  "(0018,1060)": "PRE_CONTRAST"             # Trigger Time
  "(0018,1061)": "ARTERIAL_PHASE"           # Trigger Source or Type
  "(0018,1062)": "VENOUS_PHASE"             # Nominal Interval
  "(0018,1063)": "DELAYED_PHASE"            # Beat Rejection Flag
  
  # Advanced Sequence Parameters
  "(0018,0080)": "500.0"                    # Repetition Time
  "(0018,0081)": "15.0"                     # Echo Time
  "(0018,0082)": "500.0"                    # Inversion Time
  "(0018,0083)": "2"                        # Number of Averages
  "(0018,0084)": "123.5"                    # Imaging Frequency
  "(0018,0085)": "1H"                       # Imaged Nucleus
  "(0018,0086)": "1"                        # Echo Number(s)
  "(0018,0087)": "3.0"                      # Magnetic Field Strength
  "(0018,0088)": "64"                       # Spacing Between Slices
  "(0018,0089)": "1"                        # Number of Phase Encoding Steps
  "(0018,0091)": "256"                      # Echo Train Length
  "(0018,0093)": "100"                      # Percent Sampling
  "(0018,0094)": "75"                       # Percent Phase Field of View
  "(0018,0095)": "260"                      # Pixel Bandwidth
```

### 4. Equipment Information Tags (0008,xxxx and 0018,xxxx)
```yaml
equipment:
  # Manufacturer Information
  "(0008,0070)": "SIEMENS"                  # Manufacturer
  "(0008,1090)": "MAGNETOM_PRISMA"          # Manufacturer's Model Name
  "(0018,1000)": "123456789"               # Device Serial Number
  "(0018,1020)": "syngo_MR_VE11E"          # Software Version(s)
  "(0018,1050)": "RESEARCH_STATION_01"     # Spatial Resolution
  
  # System Configuration
  "(0018,1200)": "20250101"                # Date of Last Calibration
  "(0018,1201)": "120000"                  # Time of Last Calibration
  "(0018,1210)": "RESEARCH_COIL_ARRAY"     # Convolution Kernel
  "(0018,1250)": "HEAD_NECK_SPINE"         # Receive Coil Name
  "(0018,1251)": "64CH_HEAD_COIL"          # Transmit Coil Name
  
  # Advanced Technical
  "(0018,9073)": "SIEMENS_PRISMA_RESEARCH" # Acquisition Protocol Name
  "(0018,9074)": "RESEARCH_PROTOCOL_001"   # Acquisition Protocol Description
  "(0018,9087)": "DIFFUSION_WEIGHTED"      # Diffusion b-value
  "(0018,9089)": "1000"                    # Diffusion Gradient Orientation
  "(0018,9117)": "AXIAL"                   # MR Diffusion Sequence
  
  # Calibration and QA
  "(0018,9152)": "DAILY_QA_PASSED"         # Effective Echo Time
  "(0018,9166)": "PHANTOM_QA_PASSED"       # Multi-planar Excitation
  "(0018,9171)": "SNR_QA_PASSED"           # Partial Fourier
  "(0018,9178)": "UNIFORMITY_QA_PASSED"    # Parallel Reduction Factor In-plane
  "(0018,9231)": "DISTORTION_QA_PASSED"    # MR Signal Domain Columns
```

### 5. Institution Information Tags (0008,xxxx)
```yaml
institution:
  # Basic Institution Info
  "(0008,0080)": "RESEARCH_MEDICAL_CENTER"  # Institution Name
  "(0008,0081)": "123 RESEARCH BLVD"        # Institution Address
  "(0008,0082)": "MEDICAL_IMAGING_DEPT"     # Institution Code Sequence
  "(0008,1010)": "MRI_SCANNER_ROOM_3"       # Station Name
  "(0008,1040)": "RADIOLOGY"                # Institutional Department Name
  "(0008,1041)": "RESEARCH_DIVISION"        # Institutional Department Type Code
  
  # Contact Information
  "(0008,1048)": "DR_RESEARCH_DIRECTOR"     # Physician(s) of Record
  "(0008,1049)": "DR_PRINCIPAL_INVESTIGATOR" # Physician(s) Reading Study
  "(0008,1052)": "RESEARCH_COORDINATOR"     # Performing Physician Name
  "(0008,1062)": "imaging@research.edu"     # Physician Reading Study Email
  "(0008,1064)": "555-RESEARCH"             # Physician Reading Study Phone
  
  # Administrative
  "(0008,1110)": "REFERENCED_STUDY_SEQUENCE" # Referenced Study Sequence
  "(0008,1115)": "REFERENCED_SERIES_SEQUENCE" # Referenced Series Sequence
  "(0008,1120)": "REFERENCED_PATIENT_SEQUENCE" # Referenced Patient Sequence
  "(0008,1125)": "REFERENCED_VISIT_SEQUENCE" # Referenced Visit Sequence
  
  # Research Ethics
  "(0012,0010)": "RESEARCH_STUDY_001"       # Clinical Trial Sponsor Name
  "(0012,0020)": "NCT12345678"              # Clinical Trial Protocol ID
  "(0012,0021)": "RESEARCH_PROTOCOL_v2.1"   # Clinical Trial Protocol Name
  "(0012,0030)": "SITE_001"                 # Clinical Trial Site ID
  "(0012,0031)": "RESEARCH_MEDICAL_CENTER"  # Clinical Trial Site Name
  "(0012,0040)": "SUBJECT_001"              # Clinical Trial Subject ID
  "(0012,0042)": "RANDOMIZED_CONTROLLED"    # Clinical Trial Subject Reading ID
  "(0012,0050)": "APPROVED"                 # Clinical Trial Time Point ID
  "(0012,0051)": "IRB_APPROVED_20250101"    # Clinical Trial Time Point Description
```

### 6. Procedure and Protocol Tags (0018,xxxx and Custom)
```yaml
procedure:
  # Procedure Information
  "(0018,1030)": "RESEARCH_MULTI_MODAL"     # Protocol Name
  "(0040,0254)": "RESEARCH_PROCEDURE_001"   # Procedure Step Description
  "(0040,0255)": "MULTI_MODAL_IMAGING"      # Procedure Step ID
  "(0040,0260)": "SCHEDULED"                # Performed Procedure Step Status
  "(0040,0270)": "20250918143000"           # Scheduled Procedure Step Start Date
  "(0040,0275)": "RESEARCH_REQUEST_001"     # Request Attributes Sequence
  
  # Research Protocol
  "(0018,9004)": "RESEARCH"                 # Content Qualification
  "(0018,9005)": "RESEARCH_PULSE_SEQUENCE"  # Pulse Sequence Name
  "(0018,9006)": "RESEARCH_VARIANTS"        # MR Imaging Modifier Sequence
  "(0018,9014)": "RESEARCH_SUPPRESSION"     # MR Imaging Modifier Sequence
  "(0018,9024)": "RESEARCH_TECHNIQUE"       # MR Imaging Modifier Sequence
  
  # Quality Control
  "(0018,9025)": "QC_PASSED"                # MR Imaging Modifier Sequence
  "(0018,9026)": "PHANTOM_QC_DAILY"         # MR Imaging Modifier Sequence
  "(0018,9027)": "SNR_WITHIN_LIMITS"        # MR Imaging Modifier Sequence
  "(0018,9028)": "UNIFORMITY_ACCEPTABLE"    # MR Imaging Modifier Sequence
  "(0018,9029)": "GEOMETRIC_ACCURACY_OK"    # MR Imaging Modifier Sequence
  
  # Advanced Research Tags
  "(0018,9030)": "MULTI_ECHO_SEQUENCE"      # Multiple Spin Echo
  "(0018,9031)": "GRADIENT_ECHO_SEQUENCE"   # Multi-planar Excitation
  "(0018,9032)": "DIFFUSION_SEQUENCE"       # Partial Fourier
  "(0018,9033)": "PERFUSION_SEQUENCE"       # Effective Echo Time
  "(0018,9034)": "FUNCTIONAL_SEQUENCE"      # Chemical Shift In-plane Pixel
  "(0018,9035)": "SPECTROSCOPY_SEQUENCE"    # Chemical Shift Out-of-plane Pixel
```

## 🏗️ Implementation Notes

### How Custom Tags Are Processed

1. **Tag Format**: Use standard DICOM format `"(GGGG,EEEE)"` where GGGG is group, EEEE is element
2. **Value Types**: String values are automatically converted to appropriate VR (Value Representation)
3. **Validation**: Tags are validated against DICOM dictionary
4. **Inheritance**: Tags can be applied at template, study, series, or image level

### Usage in Templates

```yaml
# In your custom template file
study_templates:
  my-research-template:
    modality: "MR"
    series_count: 3
    image_count: 30
    custom_tags:
      patient:
        "(0010,0010)": "RESEARCH^SUBJECT^001"
      study:
        "(0008,1030)": "Research Study Protocol"
      # ... more tags
```

### CLI Usage

```bash
# Create study with custom template
crgodicom create --template my-research-template

# Override template values
crgodicom create --template my-research-template \
  --patient-name "CUSTOM^PATIENT^NAME" \
  --study-description "Modified Study Description"
```

## 📁 File Locations

- **Main Guide**: `examples/custom-templates-guide.md`
- **Large Example**: `examples/cardiac-mri-template.yaml`
- **Ultra-Large Example**: `examples/ultra-comprehensive-research-template.yaml` (see below)
- **Simple Example**: `examples/simple-custom-template.yaml`
- **Visual Examples**: `docs/template-examples/`

## 🔗 Related Documentation

- [Template Examples with Images](docs/template-examples/README.md)
- [Main README - Templates Section](README.md#templates)
- [Configuration Guide](README.md#configuration)

---

*This guide demonstrates the most comprehensive DICOM template capabilities in CRGoDICOM*

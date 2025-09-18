# Ultrasound Abdomen Template Example

This example demonstrates the **ultrasound-abdomen** template, which generates Ultrasound (US) studies with abdominal imaging.

## 🔧 Template Configuration

```yaml
ultrasound-abdomen:
  modality: "US"
  series_count: 1
  image_count: 10
  anatomical_region: "abdomen"
  study_description: "Ultrasound Abdomen"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | US | Ultrasound |
| **Series Count** | 1 | Single series per study |
| **Image Count** | 10 | Ten images per series |
| **Anatomical Region** | abdomen | Abdominal imaging |
| **Study Description** | Ultrasound Abdomen | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create an ultrasound abdomen study with default settings
crgodicom create --template ultrasound-abdomen

# Create with custom patient information
crgodicom create --template ultrasound-abdomen \
  --patient-name "WILSON^SARAH^L" \
  --patient-id "P003456" \
  --accession-number "ACC456789"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file ultrasound-abdomen-report.pdf

# Export with custom output directory
crgodicom export --study-id <study-id> --format png \
  --output-dir /path/to/exports
```

### PACS Transmission
```bash
# Send to PACS server
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --aec DICOM_CLIENT --aet PACS1

# Send with specific AE titles
crgodicom dcmtk --study-id <study-id> \
  --host pacs.example.com --port 104 \
  --aec ULTRASOUND_SCANNER --aet PACS_SERVER
```

## 🖼️ Generated Content

### Sample Images
This example study contains **10 images** in **1 series**:

- **Series 001**: Abdominal Ultrasound
  - Multiple views of abdominal organs
  - Liver, gallbladder, kidneys, pancreas
  - Vascular structures

### Image Features
- **Resolution**: 640×480 pixels (typical for ultrasound)
- **Bit Depth**: 8-bit grayscale
- **Color Space**: Monochrome (grayscale)
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic ultrasound appearance with organ boundaries and acoustic shadows

## 📊 DICOM Metadata

### Patient Information
```
Patient Name: DEMO^PATIENT
Patient ID: P001
Patient Birth Date: 19850101
Patient Sex: F
Patient Age: 040Y
```

### Study Information
```
Study Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783
Study Date: 20250101
Study Time: 140000.000000
Study Description: Ultrasound Abdomen
Accession Number: ACC789123
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117784
Series Number: 1
Modality: US
Series Description: Ultrasound Abdomen
Body Part Examined: ABDOMEN
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.6.1 (Ultrasound Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\ABDOMEN
Rows: 640
Columns: 480
Bits Allocated: 8
Bits Stored: 8
High Bit: 7
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
Ultrasound Color Data Present: 0
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - All ultrasound images with metadata overlays
   - Sequential image display
   - Organ identification

3. **Technical Details**
   - Complete DICOM metadata dump
   - Ultrasound-specific parameters
   - File structure information

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783/
├── series_001/
│   ├── image_001.dcm          # US image 1
│   ├── image_002.dcm          # US image 2
│   ├── image_003.dcm          # US image 3
│   └── ...                    # Additional US images
└── exports/
    ├── series_001/
    │   ├── image_001.png      # PNG export 1
    │   ├── image_002.png      # PNG export 2
    │   └── ...                # Additional PNG exports
    └── study_report.pdf       # Complete PDF report
```

## 🎯 Use Cases

### Clinical Training
- **Ultrasound Education**: Practice reading abdominal ultrasounds
- **Medical Students**: Learn abdominal anatomy on ultrasound
- **Sonographers**: Study image acquisition techniques

### Software Testing
- **DICOM Viewers**: Test ultrasound image display
- **PACS Systems**: Validate ultrasound workflow
- **AI Algorithms**: Train abdominal pathology detection

### Research Applications
- **Image Processing**: Test enhancement algorithms
- **Quantitative Analysis**: Study organ measurements
- **Standardization**: Validate DICOM compliance

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template ultrasound-abdomen \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19900101" \
  --patient-sex "M"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template ultrasound-abdomen \
  --study-description "Complete Abdominal Ultrasound" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify image count
crgodicom create --template ultrasound-abdomen \
  --image-count 15 \
  --series-description "Extended Abdominal Survey"
```

## 🏥 Clinical Context

### Typical Ultrasound Abdomen Protocol
- **Indication**: Abdominal pain, organ assessment, screening
- **Preparation**: Fasting (for gallbladder), full bladder (for pelvis)
- **Scanning Planes**: Transverse, sagittal, coronal
- **Organs Examined**: Liver, gallbladder, kidneys, pancreas, spleen, aorta

### Common Findings
- **Normal Anatomy**: Organ echotexture, vascular flow
- **Pathology**: Gallstones, liver lesions, renal cysts
- **Artifacts**: Acoustic shadowing, reverberation

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

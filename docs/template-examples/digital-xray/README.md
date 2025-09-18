# Digital X-Ray Template Example

This example demonstrates the **digital-xray** template, which generates Digital X-Ray (DX) studies.

## 🔧 Template Configuration

```yaml
digital-xray:
  modality: "DX"
  series_count: 1
  image_count: 1
  anatomical_region: "chest"
  study_description: "Digital X-Ray"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | DX | Digital X-Ray |
| **Series Count** | 1 | Single series per study |
| **Image Count** | 1 | One image per series |
| **Anatomical Region** | chest | Chest/thorax imaging |
| **Study Description** | Digital X-Ray | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create a digital X-ray study with default settings
crgodicom create --template digital-xray

# Create with custom patient information
crgodicom create --template digital-xray \
  --patient-name "BROWN^MICHAEL^R" \
  --patient-id "P005678" \
  --accession-number "ACC678901"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file digital-xray-report.pdf

# Export with custom output file
crgodicom export --study-id <study-id> --format pdf \
  --output-file /path/to/custom-report.pdf
```

### PACS Transmission
```bash
# Send to PACS server
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --aec DICOM_CLIENT --aet PACS1

# Send to multiple PACS servers
crgodicom dcmtk --study-id <study-id> \
  --host pacs1.example.com --port 104
crgodicom dcmtk --study-id <study-id> \
  --host pacs2.example.com --port 104
```

## 🖼️ Generated Content

### Sample Images
This example study contains **1 image** in **1 series**:

- **Series 001**: Digital X-Ray
  - Single high-quality digital radiograph
  - Chest/thorax imaging
  - AP or PA view

### Image Features
- **Resolution**: 2048×2048 pixels (high-resolution digital radiography)
- **Bit Depth**: 16-bit grayscale
- **Color Space**: Monochrome (grayscale)
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic digital X-ray appearance with chest anatomy

## 📊 DICOM Metadata

### Patient Information
```
Patient Name: DEMO^PATIENT
Patient ID: P001
Patient Birth Date: 19800101
Patient Sex: M
Patient Age: 045Y
```

### Study Information
```
Study Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783
Study Date: 20250101
Study Time: 110000.000000
Study Description: Digital X-Ray
Accession Number: ACC789123
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117784
Series Number: 1
Modality: DX
Series Description: Digital X-Ray
Body Part Examined: CHEST
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.1.1.1 (Digital X-Ray Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\CHEST
Rows: 2048
Columns: 2048
Bits Allocated: 16
Bits Stored: 14
High Bit: 13
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
KVP: 120
X-Ray Tube Current: 320
Exposure Time: 20
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - Digital X-ray image with metadata overlay
   - High-resolution display
   - Technical parameters

3. **Technical Details**
   - Complete DICOM metadata dump
   - Digital radiography parameters
   - File structure information

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783/
├── series_001/
│   └── image_001.dcm          # Digital X-ray image
└── exports/
    ├── series_001/
    │   └── image_001.png      # PNG export
    └── study_report.pdf       # Complete PDF report
```

## 🎯 Use Cases

### Clinical Training
- **Radiology Education**: Practice reading digital X-rays
- **Medical Students**: Learn chest anatomy on radiographs
- **Residents**: Study digital radiography techniques

### Software Testing
- **DICOM Viewers**: Test digital X-ray image display
- **PACS Systems**: Validate digital radiography workflow
- **AI Algorithms**: Train pathology detection algorithms

### Research Applications
- **Image Processing**: Test enhancement algorithms
- **Quantitative Analysis**: Study image quality metrics
- **Standardization**: Validate DICOM compliance

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template digital-xray \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19900101" \
  --patient-sex "F"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template digital-xray \
  --study-description "Chest X-Ray PA View" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify anatomical region
crgodicom create --template digital-xray \
  --anatomical-region "spine" \
  --study-description "Lumbar Spine X-Ray"
```

## 🏥 Clinical Context

### Typical Digital X-Ray Protocol
- **Indication**: Routine screening, trauma assessment, follow-up
- **Preparation**: Remove jewelry, change into gown
- **Standard Views**: PA (Posterior-Anterior), Lateral, AP (Anterior-Posterior)
- **Technical Factors**: High kVp, low mAs for reduced dose

### Common Findings
- **Normal Anatomy**: Lungs, heart, ribs, spine
- **Pathology**: Pneumonia, heart failure, fractures, masses
- **Artifacts**: Motion, foreign objects, positioning

### Image Quality Factors
- **Exposure**: Proper kVp and mAs settings
- **Positioning**: Correct patient positioning
- **Motion**: Patient cooperation and breath-holding
- **Processing**: Appropriate window/level settings

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

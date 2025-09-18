# Mammography Template Example

This example demonstrates the **mammography** template, which generates Mammography (MG) studies for breast imaging.

## 🔧 Template Configuration

```yaml
mammography:
  modality: "MG"
  series_count: 2
  image_count: 4
  anatomical_region: "breast"
  study_description: "Mammography"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | MG | Mammography |
| **Series Count** | 2 | Two series per study |
| **Image Count** | 4 | Four images per series |
| **Anatomical Region** | breast | Breast imaging |
| **Study Description** | Mammography | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create a mammography study with default settings
crgodicom create --template mammography

# Create with custom patient information
crgodicom create --template mammography \
  --patient-name "DAVIS^MARY^K" \
  --patient-id "P004567" \
  --accession-number "ACC567890"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file mammography-report.pdf

# Export specific series only
crgodicom export --study-id <study-id> --series-number 1 --format png
```

### PACS Transmission
```bash
# Send to PACS server
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --aec DICOM_CLIENT --aet PACS1

# Send with verbose output for debugging
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --verbose
```

## 🖼️ Generated Content

### Sample Images
This example study contains **4 images** across **2 series**:

- **Series 001**: Right breast imaging
  - CC (Cranio-Caudal) view
  - MLO (Medio-Lateral Oblique) view
- **Series 002**: Left breast imaging
  - CC (Cranio-Caudal) view
  - MLO (Medio-Lateral Oblique) view

### Image Features
- **Resolution**: 2048×2560 pixels (high-resolution mammography)
- **Bit Depth**: 16-bit grayscale
- **Color Space**: Monochrome (grayscale)
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic mammographic appearance with breast tissue patterns

## 📊 DICOM Metadata

### Patient Information
```
Patient Name: DEMO^PATIENT
Patient ID: P001
Patient Birth Date: 19700101
Patient Sex: F
Patient Age: 055Y
```

### Study Information
```
Study Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783
Study Date: 20250101
Study Time: 090000.000000
Study Description: Mammography
Accession Number: ACC789123
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117784
Series Number: 1
Modality: MG
Series Description: Right Breast
Body Part Examined: BREAST
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.1.1.1 (Digital Mammography X-Ray Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\BREAST
Rows: 2048
Columns: 2560
Bits Allocated: 16
Bits Stored: 14
High Bit: 13
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
KVP: 28
X-Ray Tube Current: 100
Exposure Time: 1000
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - All mammographic images with metadata overlays
   - Bilateral breast comparison
   - View labeling (CC, MLO)

3. **Technical Details**
   - Complete DICOM metadata dump
   - Mammography-specific parameters
   - File structure information

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783/
├── series_001/
│   ├── image_001.dcm          # Right CC view
│   ├── image_002.dcm          # Right MLO view
│   └── ...
├── series_002/
│   ├── image_001.dcm          # Left CC view
│   ├── image_002.dcm          # Left MLO view
│   └── ...
└── exports/
    ├── series_001/
    │   ├── image_001.png      # PNG export series 1
    │   └── ...
    ├── series_002/
    │   ├── image_001.png      # PNG export series 2
    │   └── ...
    └── study_report.pdf       # Complete PDF report
```

## 🎯 Use Cases

### Clinical Training
- **Radiology Education**: Practice reading mammograms
- **Medical Students**: Learn breast anatomy and pathology
- **Residents**: Study mammographic imaging techniques

### Software Testing
- **DICOM Viewers**: Test mammography image display
- **PACS Systems**: Validate mammography workflow
- **AI Algorithms**: Train breast cancer detection

### Research Applications
- **Image Processing**: Test enhancement algorithms
- **Computer-Aided Detection**: Develop CAD systems
- **Quantitative Analysis**: Study breast density

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template mammography \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19650101" \
  --patient-sex "F"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template mammography \
  --study-description "Screening Mammography" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify series and image counts
crgodicom create --template mammography \
  --series-count 1 \
  --image-count 2 \
  --series-description "Bilateral Screening"
```

## 🏥 Clinical Context

### Typical Mammography Protocol
- **Indication**: Breast cancer screening, diagnostic evaluation
- **Preparation**: No special preparation required
- **Standard Views**: CC (Cranio-Caudal), MLO (Medio-Lateral Oblique)
- **Additional Views**: Spot compression, magnification, lateral

### Common Findings
- **Normal Anatomy**: Breast tissue, Cooper's ligaments, nipple
- **Benign Changes**: Cysts, fibroadenomas, calcifications
- **Malignant Signs**: Masses, architectural distortion, suspicious calcifications

### BI-RADS Categories
- **BI-RADS 1**: Negative
- **BI-RADS 2**: Benign findings
- **BI-RADS 3**: Probably benign
- **BI-RADS 4**: Suspicious abnormality
- **BI-RADS 5**: Highly suggestive of malignancy

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

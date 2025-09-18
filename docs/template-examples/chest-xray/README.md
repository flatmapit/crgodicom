# Chest X-Ray Template Example

This example demonstrates the **chest-xray** template, which generates Computed Radiography (CR) studies with chest X-ray images.

## 🔧 Template Configuration

```yaml
chest-xray:
  modality: "CR"
  series_count: 1
  image_count: 2
  anatomical_region: "chest"
  study_description: "Chest X-Ray"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | CR | Computed Radiography |
| **Series Count** | 1 | Single series per study |
| **Image Count** | 2 | Two images per series |
| **Anatomical Region** | chest | Chest/thorax imaging |
| **Study Description** | Chest X-Ray | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create a chest X-ray study with default settings
crgodicom create --template chest-xray

# Create with custom patient information
crgodicom create --template chest-xray \
  --patient-name "SMITH^JANE^M" \
  --patient-id "P001234" \
  --accession-number "ACC789456"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file chest-xray-report.pdf

# Export to both formats
crgodicom export --study-id <study-id> --format png pdf
```

### PACS Transmission
```bash
# Send to PACS server
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --aec DICOM_CLIENT --aet PACS1

# Check DCMTK availability first
crgodicom check-dcmtk
```

## 🖼️ Generated Content

### Sample Images
This example study contains **2 images** in **1 series**:

- **Series 001**: Chest X-Ray CR images
  - `image_001.png` - Anterior-Posterior view
  - `image_002.png` - Lateral view

### Image Features
- **Resolution**: 512×512 pixels (typical for CR)
- **Bit Depth**: 16-bit grayscale
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic chest X-ray appearance with ribs, lungs, and heart shadow

## 📊 DICOM Metadata

### Patient Information
```
Patient Name: DEMO^CHEST
Patient ID: CX001
Patient Birth Date: 19800101
Patient Sex: M
Patient Age: 045Y
```

### Study Information
```
Study Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783
Study Date: 20250101
Study Time: 120000.000000
Study Description: Chest X-Ray
Accession Number: ACC123456
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117784
Series Number: 1
Modality: CR
Series Description: Chest X-Ray
Body Part Examined: CHEST
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.1 (Computed Radiography Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\CHEST
Rows: 512
Columns: 512
Bits Allocated: 16
Bits Stored: 12
High Bit: 11
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - All images with metadata overlays
   - High-resolution PNG exports
   - Series and instance information

3. **Technical Details**
   - Complete DICOM metadata dump
   - File structure information
   - Export timestamps

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783/
├── series_001/
│   ├── image_001.dcm          # DICOM file 1
│   └── image_002.dcm          # DICOM file 2
└── exports/
    ├── series_001/
    │   ├── image_001.png      # PNG export 1
    │   └── image_002.png      # PNG export 2
    └── study_report.pdf       # Complete PDF report
```

## 🎯 Use Cases

### Clinical Training
- **Radiology Education**: Practice reading chest X-rays
- **Medical Students**: Learn normal chest anatomy
- **Residents**: Study CR imaging characteristics

### Software Testing
- **DICOM Viewers**: Test CR image display
- **PACS Systems**: Validate CR workflow
- **AI Algorithms**: Train chest pathology detection

### Research Applications
- **Image Processing**: Test enhancement algorithms
- **Compression Studies**: Evaluate CR compression
- **Standardization**: Validate DICOM compliance

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template chest-xray \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19900101" \
  --patient-sex "F"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template chest-xray \
  --study-description "Routine Chest X-Ray" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify image characteristics
crgodicom create --template chest-xray \
  --image-count 3 \
  --series-description "AP and Lateral Views"
```

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

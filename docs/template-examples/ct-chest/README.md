# CT Chest Template Example

This example demonstrates the **ct-chest** template, which generates Computed Tomography (CT) studies with chest imaging series.

## 🔧 Template Configuration

```yaml
ct-chest:
  modality: "CT"
  series_count: 2
  image_count: 50
  anatomical_region: "chest"
  study_description: "CT Chest"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | CT | Computed Tomography |
| **Series Count** | 2 | Two series per study |
| **Image Count** | 50 | Fifty images per series |
| **Anatomical Region** | chest | Chest/thorax imaging |
| **Study Description** | CT Chest | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create a CT chest study with default settings
crgodicom create --template ct-chest

# Create with custom patient information
crgodicom create --template ct-chest \
  --patient-name "JOHNSON^ROBERT^A" \
  --patient-id "P002345" \
  --accession-number "ACC789123"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file ct-chest-report.pdf

# Export specific series only
crgodicom export --study-id <study-id> --series-number 1 --format png
```

### PACS Transmission
```bash
# Send to PACS server
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --aec DICOM_CLIENT --aet PACS1

# Send with verbose output
crgodicom dcmtk --study-id <study-id> \
  --host localhost --port 4242 \
  --verbose
```

## 🖼️ Generated Content

### Sample Images
This example study contains **multiple images** across **3 series**:

- **Series 001**: Axial CT images (chest)
- **Series 002**: Coronal CT images (chest)  
- **Series 003**: Sagittal CT images (chest)

### Image Features
- **Resolution**: 512×512 pixels (standard CT resolution)
- **Bit Depth**: 16-bit grayscale with CT windowing
- **Slice Thickness**: Variable (typically 1-5mm)
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic chest CT appearance with lungs, heart, and mediastinum

## 📊 DICOM Metadata

### Patient Information
```
Patient Name: DOE^JOHN^M
Patient ID: P123456
Patient Birth Date: 19800101
Patient Sex: M
Patient Age: 045Y
```

### Study Information
```
Study Instance UID: 1.2.840.10008.5.1.4.1.1.1758084608.4462841482011563283
Study Date: 20250101
Study Time: 120000.000000
Study Description: CT Chest
Accession Number: ACC123456
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758084608.4462841482011563284
Series Number: 1
Modality: CT
Series Description: CT Chest
Body Part Examined: CHEST
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.2 (CT Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\AXIAL
Rows: 512
Columns: 512
Bits Allocated: 16
Bits Stored: 12
High Bit: 11
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
Slice Thickness: 5.0
KVP: 120
X-Ray Tube Current: 200
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - All images with metadata overlays
   - Series organization
   - Instance numbering

3. **Technical Details**
   - Complete DICOM metadata dump
   - CT-specific parameters
   - File structure information

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758084608.4462841482011563283/
├── series_001/
│   ├── image_001.dcm          # Axial slice 1
│   ├── image_002.dcm          # Axial slice 2
│   └── ...                    # Additional axial slices
├── series_002/
│   ├── image_001.dcm          # Coronal slice 1
│   ├── image_002.dcm          # Coronal slice 2
│   └── ...                    # Additional coronal slices
├── series_003/
│   ├── image_001.dcm          # Sagittal slice 1
│   ├── image_002.dcm          # Sagittal slice 2
│   └── ...                    # Additional sagittal slices
└── exports/
    ├── series_001/
    │   ├── image_001.png      # PNG export series 1
    │   └── ...
    ├── series_002/
    │   ├── image_001.png      # PNG export series 2
    │   └── ...
    ├── series_003/
    │   ├── image_001.png      # PNG export series 3
    │   └── ...
    └── study_report.pdf       # Complete PDF report
```

## 🎯 Use Cases

### Clinical Training
- **Radiology Education**: Practice reading CT chest scans
- **Medical Students**: Learn cross-sectional anatomy
- **Residents**: Study CT imaging protocols

### Software Testing
- **DICOM Viewers**: Test CT image display and windowing
- **PACS Systems**: Validate CT workflow and storage
- **3D Reconstruction**: Test volume rendering algorithms

### Research Applications
- **AI Development**: Train lung nodule detection
- **Image Processing**: Test enhancement algorithms
- **Quantitative Analysis**: Study CT density measurements

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template ct-chest \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19750101" \
  --patient-sex "F"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template ct-chest \
  --study-description "CT Chest with Contrast" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify series and image counts
crgodicom create --template ct-chest \
  --series-count 1 \
  --image-count 100 \
  --series-description "High Resolution CT"
```

## 🏥 Clinical Context

### Typical CT Chest Protocol
- **Indication**: Suspected pulmonary pathology, trauma assessment
- **Contrast**: May include contrast-enhanced series
- **Reconstruction**: Multiple reconstruction algorithms
- **Window Settings**: Lung window (W:1500, L:-600), Mediastinal window (W:400, L:40)

### Common Findings
- **Normal Anatomy**: Lungs, heart, mediastinum, ribs
- **Pathology**: Nodules, masses, effusions, pneumothorax
- **Artifacts**: Motion, beam hardening, metallic

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

# MRI Brain Template Example

This example demonstrates the **mri-brain** template, which generates Magnetic Resonance Imaging (MR) studies with brain imaging.

## 🔧 Template Configuration

```yaml
mri-brain:
  modality: "MR"
  series_count: 3
  image_count: 30
  anatomical_region: "brain"
  study_description: "MRI Brain"
```

## 📋 Template Parameters

| Parameter | Value | Description |
|-----------|-------|-------------|
| **Modality** | MR | Magnetic Resonance Imaging |
| **Series Count** | 3 | Three series per study |
| **Image Count** | 30 | Thirty images per series |
| **Anatomical Region** | brain | Brain imaging |
| **Study Description** | MRI Brain | Human-readable study name |

## 🚀 Usage Examples

### Basic Generation
```bash
# Create an MRI brain study with default settings
crgodicom create --template mri-brain

# Create with custom patient information
crgodicom create --template mri-brain \
  --patient-name "TAYLOR^JENNIFER^A" \
  --patient-id "P006789" \
  --accession-number "ACC789012"
```

### Export Options
```bash
# Export to PNG images with metadata overlay
crgodicom export --study-id <study-id> --format png

# Export to PDF report
crgodicom export --study-id <study-id> --format pdf \
  --output-file mri-brain-report.pdf

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
This example study contains **30 images** across **3 series**:

- **Series 001**: T1-weighted axial images
  - 30 axial slices through the brain
  - High contrast for anatomical detail
- **Series 002**: T2-weighted axial images
  - 30 axial slices with fluid-sensitive contrast
  - Pathological contrast enhancement
- **Series 003**: FLAIR axial images
  - 30 axial slices with fluid attenuation
  - White matter lesion detection

### Image Features
- **Resolution**: 256×256 pixels (standard MRI resolution)
- **Bit Depth**: 16-bit grayscale
- **Color Space**: Monochrome (grayscale)
- **Metadata Overlay**: Patient info, study details, and technical parameters
- **Anatomical Content**: Synthetic brain MRI appearance with gray matter, white matter, and CSF

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
Study Time: 130000.000000
Study Description: MRI Brain
Accession Number: ACC789123
Study ID: S001
```

### Series Information
```
Series Instance UID: 1.2.840.10008.5.1.4.1.1.1758083884.615900210705117784
Series Number: 1
Modality: MR
Series Description: T1 Axial
Body Part Examined: BRAIN
```

### Technical Parameters
```
SOP Class UID: 1.2.840.10008.5.1.4.1.1.4 (MR Image Storage)
Transfer Syntax UID: 1.2.840.10008.1.2 (Implicit VR Little Endian)
Image Type: ORIGINAL\PRIMARY\BRAIN
Rows: 256
Columns: 256
Bits Allocated: 16
Bits Stored: 12
High Bit: 11
Pixel Representation: 0
Photometric Interpretation: MONOCHROME2
Slice Thickness: 5.0
Repetition Time: 500.0
Echo Time: 15.0
Magnetic Field Strength: 1.5
```

## 📄 PDF Report

The generated PDF report includes:

1. **Cover Page**
   - Study summary and patient information
   - Study and series UIDs
   - Technical specifications

2. **Image Gallery**
   - All MRI images with metadata overlays
   - Multi-series display
   - Sequence identification

3. **Technical Details**
   - Complete DICOM metadata dump
   - MRI-specific parameters
   - File structure information

## 🔍 File Structure

```
studies/1.2.840.10008.5.1.4.1.1.1758083884.615900210705117783/
├── series_001/
│   ├── image_001.dcm          # T1 slice 1
│   ├── image_002.dcm          # T1 slice 2
│   └── ...                    # Additional T1 slices
├── series_002/
│   ├── image_001.dcm          # T2 slice 1
│   ├── image_002.dcm          # T2 slice 2
│   └── ...                    # Additional T2 slices
├── series_003/
│   ├── image_001.dcm          # FLAIR slice 1
│   ├── image_002.dcm          # FLAIR slice 2
│   └── ...                    # Additional FLAIR slices
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
- **Radiology Education**: Practice reading brain MRIs
- **Medical Students**: Learn neuroanatomy
- **Residents**: Study MRI sequences and protocols

### Software Testing
- **DICOM Viewers**: Test MRI image display and windowing
- **PACS Systems**: Validate MRI workflow and storage
- **AI Algorithms**: Train brain pathology detection

### Research Applications
- **Image Processing**: Test enhancement algorithms
- **3D Reconstruction**: Test volume rendering
- **Quantitative Analysis**: Study brain volumetry

## 🔧 Customization Options

### Patient Data
```bash
# Custom patient demographics
crgodicom create --template mri-brain \
  --patient-name "LAST^FIRST^MIDDLE" \
  --patient-id "CUSTOM_ID" \
  --patient-birth-date "19750101" \
  --patient-sex "F"
```

### Study Parameters
```bash
# Custom study information
crgodicom create --template mri-brain \
  --study-description "MRI Brain with Contrast" \
  --accession-number "ACC2025001" \
  --study-id "S2025001"
```

### Technical Settings
```bash
# Modify series and image counts
crgodicom create --template mri-brain \
  --series-count 2 \
  --image-count 50 \
  --series-description "High Resolution T1"
```

## 🏥 Clinical Context

### Typical MRI Brain Protocol
- **Indication**: Headache, neurological symptoms, follow-up
- **Preparation**: Remove metal objects, contrast if needed
- **Standard Sequences**: T1, T2, FLAIR, DWI
- **Additional Sequences**: T1 post-contrast, MRA, DTI

### Common Findings
- **Normal Anatomy**: Gray matter, white matter, CSF spaces
- **Pathology**: Tumors, strokes, demyelination, atrophy
- **Artifacts**: Motion, susceptibility, chemical shift

### MRI Sequences
- **T1-weighted**: Anatomical detail, post-contrast enhancement
- **T2-weighted**: Fluid-sensitive, pathological contrast
- **FLAIR**: Fluid attenuation, white matter lesions
- **DWI**: Diffusion-weighted, acute ischemia

## 📞 Support

- **Documentation**: [Template Examples](../README.md)
- **GitHub**: [CRGoDICOM Repository](https://github.com/flatmapit/crgodicom)
- **Issues**: [Report Problems](https://github.com/flatmapit/crgodicom/issues)

---

*This example was generated using CRGoDICOM v8f67d2c*

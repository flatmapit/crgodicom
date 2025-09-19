package dicom

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

// Writer handles writing DICOM files to disk
type Writer struct {
	config *config.Config
}

// NewWriter creates a new DICOM writer
func NewWriter(cfg *config.Config) *Writer {
	return &Writer{
		config: cfg,
	}
}

// WriteStudy writes a complete study to disk
func (w *Writer) WriteStudy(study *types.Study, outputDir string) error {
	// Create study directory
	studyDir := filepath.Join(outputDir, study.StudyInstanceUID)
	if err := os.MkdirAll(studyDir, 0755); err != nil {
		return fmt.Errorf("failed to create study directory: %w", err)
	}

	logrus.Infof("Writing study %s to %s", study.StudyInstanceUID, studyDir)

	// Write study metadata
	if err := w.writeStudyMetadata(study, studyDir); err != nil {
		return fmt.Errorf("failed to write study metadata: %w", err)
	}

	// Write series
	for i, series := range study.Series {
		seriesDir := filepath.Join(studyDir, fmt.Sprintf("series_%03d", i+1))
		if err := os.MkdirAll(seriesDir, 0755); err != nil {
			return fmt.Errorf("failed to create series directory: %w", err)
		}

		if err := w.writeSeries(study, &series, seriesDir); err != nil {
			return fmt.Errorf("failed to write series %d: %w", i+1, err)
		}
	}

	logrus.Infof("Successfully wrote study with %d series", len(study.Series))
	return nil
}

// writeStudyMetadata writes study metadata to JSON file
func (w *Writer) writeStudyMetadata(study *types.Study, studyDir string) error {
	// TODO: Implement JSON metadata writing
	// For now, just log the metadata
	logrus.Infof("Study metadata: UID=%s, Patient=%s (%s), Description=%s",
		study.StudyInstanceUID, study.PatientName, study.PatientID, study.StudyDescription)
	return nil
}

// writeSeries writes a series to disk
func (w *Writer) writeSeries(study *types.Study, series *types.Series, seriesDir string) error {
	logrus.Infof("Writing series %s with %d images", series.SeriesInstanceUID, len(series.Images))

	for i, image := range series.Images {
		imageFile := filepath.Join(seriesDir, fmt.Sprintf("image_%03d.dcm", i+1))
		if err := w.writeImage(study, series, &image, imageFile); err != nil {
			return fmt.Errorf("failed to write image %d: %w", i+1, err)
		}
	}

	return nil
}

// writeImage writes a single DICOM image to disk
func (w *Writer) writeImage(study *types.Study, series *types.Series, image *types.Image, filePath string) error {
	// Create DICOM dataset
	dataset := dicom.Dataset{
		Elements: make([]*dicom.Element, 0),
	}

	// Add elements in correct ascending tag order
	w.addElementsInOrder(&dataset, study, series, image)

	// Create DICOM file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create DICOM file: %w", err)
	}
	defer file.Close()

	// Write DICOM file
	logrus.Infof("Writing DICOM file with %d elements", len(dataset.Elements))
	for i, elem := range dataset.Elements {
		logrus.Infof("Element %d: Tag=%s", i, elem.Tag)
	}

	if err := dicom.Write(file, dataset); err != nil {
		logrus.Errorf("DICOM write failed: %v", err)
		return fmt.Errorf("failed to write DICOM file: %w", err)
	}

	return nil
}

// addElementsInOrder adds DICOM elements in correct ascending tag order
func (w *Writer) addElementsInOrder(dataset *dicom.Dataset, study *types.Study, series *types.Series, image *types.Image) {
	// File Meta Information Group (0002,xxxx)
	w.addMandatoryElements(dataset, image)

	// Image Presentation Group (0008,xxxx)
	w.addImagePresentationElements(dataset, study, series, image)

	// Patient Group (0010,xxxx)
	w.addPatientElements(dataset, study)

	// Study Group (0020,xxxx) - Study Instance UID
	if elem, err := dicom.NewElement(tag.StudyInstanceUID, []string{study.StudyInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Group (0020,xxxx) - Series Instance UID
	if elem, err := dicom.NewElement(tag.SeriesInstanceUID, []string{series.SeriesInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Number (0020,0011) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.SeriesNumber, []int{series.SeriesNumber}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Instance Number (0020,0013) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.InstanceNumber, []int{image.InstanceNumber}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Image Pixel Group (0028,xxxx)
	w.addImagePixelElements(dataset, image)

	// Pixel Data (7FE0,0010) - TEMPORARILY COMMENTED OUT TO TEST
	// w.addPixelDataElements(dataset, image)
}

// addImagePresentationElements adds 0008 group elements in order
func (w *Writer) addImagePresentationElements(dataset *dicom.Dataset, study *types.Study, series *types.Series, image *types.Image) {
	// SOP Class UID (0008,0016)
	if elem, err := dicom.NewElement(tag.SOPClassUID, []string{image.SOPClassUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// SOP Instance UID (0008,0018)
	if elem, err := dicom.NewElement(tag.SOPInstanceUID, []string{image.SOPInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Date (0008,0020)
	if elem, err := dicom.NewElement(tag.StudyDate, []string{study.StudyDate}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Time (0008,0030)
	if elem, err := dicom.NewElement(tag.StudyTime, []string{study.StudyTime}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Accession Number (0008,0050)
	if elem, err := dicom.NewElement(tag.AccessionNumber, []string{study.AccessionNumber}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Modality (0008,0060)
	if elem, err := dicom.NewElement(tag.Modality, []string{series.Modality}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Description (0008,1030)
	if elem, err := dicom.NewElement(tag.StudyDescription, []string{study.StudyDescription}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Description (0008,103E)
	if elem, err := dicom.NewElement(tag.SeriesDescription, []string{series.SeriesDescription}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addImagePixelElements adds 0028 group elements in order
func (w *Writer) addImagePixelElements(dataset *dicom.Dataset, image *types.Image) {
	// Samples per Pixel (0028,0002) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.SamplesPerPixel, []int{1}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Photometric Interpretation (0028,0004) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Planar Configuration (0028,0006) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.PlanarConfiguration, []int{0}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Rows (0028,0010) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.Rows, []int{int(image.Height)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Columns (0028,0011) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.Columns, []int{int(image.Width)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Bits Allocated (0028,0100) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.BitsAllocated, []int{int(image.BitsPerPixel)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Bits Stored (0028,0101) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.BitsStored, []int{int(image.BitsPerPixel)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// High Bit (0028,0102) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.HighBit, []int{int(image.BitsPerPixel - 1)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Pixel Representation (0028,0103) - TEMPORARILY COMMENTED OUT TO TEST
	// if elem, err := dicom.NewElement(tag.PixelRepresentation, []int{0}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }
}

// addPatientElements adds patient-related DICOM elements
func (w *Writer) addPatientElements(dataset *dicom.Dataset, study *types.Study) {
	// Patient Name (0010,0010)
	if elem, err := dicom.NewElement(tag.PatientName, []string{study.PatientName}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient ID (0010,0020)
	if elem, err := dicom.NewElement(tag.PatientID, []string{study.PatientID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient Birth Date (0010,0030)
	if elem, err := dicom.NewElement(tag.PatientBirthDate, []string{study.PatientBirthDate}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient Sex (0010,0040) - Default to "O" (Other)
	if elem, err := dicom.NewElement(tag.PatientSex, []string{"O"}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addStudyElements adds study-related DICOM elements
func (w *Writer) addStudyElements(dataset *dicom.Dataset, study *types.Study) {
	// Study Instance UID (0020,000D)
	if elem, err := dicom.NewElement(tag.StudyInstanceUID, []string{study.StudyInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Date (0008,0020)
	if elem, err := dicom.NewElement(tag.StudyDate, []string{study.StudyDate}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Time (0008,0030)
	if elem, err := dicom.NewElement(tag.StudyTime, []string{study.StudyTime}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Description (0008,1030)
	if elem, err := dicom.NewElement(tag.StudyDescription, []string{study.StudyDescription}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Accession Number (0008,0050)
	if elem, err := dicom.NewElement(tag.AccessionNumber, []string{study.AccessionNumber}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addSeriesElements adds series-related DICOM elements
func (w *Writer) addSeriesElements(dataset *dicom.Dataset, series *types.Series) {
	// Series Instance UID (0020,000E)
	if elem, err := dicom.NewElement(tag.SeriesInstanceUID, []string{series.SeriesInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Number (0020,0011) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.SeriesNumber, []int{series.SeriesNumber}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Modality (0008,0060) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.Modality, []string{series.Modality}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Series Description (0008,103E) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.SeriesDescription, []string{series.SeriesDescription}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }
}

// addImageElements adds image-related DICOM elements
func (w *Writer) addImageElements(dataset *dicom.Dataset, image *types.Image) {
	// SOP Instance UID (0008,0018)
	if elem, err := dicom.NewElement(tag.SOPInstanceUID, []string{image.SOPInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// SOP Class UID (0008,0016)
	if elem, err := dicom.NewElement(tag.SOPClassUID, []string{image.SOPClassUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Instance Number (0020,0013) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.InstanceNumber, []int{image.InstanceNumber}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Image dimensions
	w.addImageDimensionElements(dataset, image)

	// TEMPORARILY SKIP PIXEL DATA TO TEST
	// Pixel data
	// w.addPixelDataElements(dataset, image)
}

// addImageDimensionElements adds image dimension elements
func (w *Writer) addImageDimensionElements(dataset *dicom.Dataset, image *types.Image) {
	// Rows (0028,0010) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.Rows, []int{int(image.Height)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Columns (0028,0011) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.Columns, []int{int(image.Width)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Bits Allocated (0028,0100) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.BitsAllocated, []int{int(image.BitsPerPixel)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Bits Stored (0028,0101) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.BitsStored, []int{int(image.BitsPerPixel)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// High Bit (0028,0102) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.HighBit, []int{int(image.BitsPerPixel - 1)}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Pixel Representation (0028,0103) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.PixelRepresentation, []int{0}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Samples per Pixel (0028,0002) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.SamplesPerPixel, []int{1}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Photometric Interpretation (0028,0004) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.PhotometricInterpretation, []string{"MONOCHROME2"}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }

	// Planar Configuration (0028,0006) - TEMPORARILY COMMENTED OUT
	// if elem, err := dicom.NewElement(tag.PlanarConfiguration, []int{0}); err == nil {
	//	dataset.Elements = append(dataset.Elements, elem)
	// }
}

// addPixelDataElements adds pixel data elements
func (w *Writer) addPixelDataElements(dataset *dicom.Dataset, image *types.Image) {
	// Pixel Data (7FE0,0010) - pixel data should be []byte
	logrus.Infof("Creating Pixel Data element with %d bytes", len(image.PixelData))
	if elem, err := dicom.NewElement(tag.PixelData, image.PixelData); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		logrus.Infof("✅ Pixel Data element created successfully")
	} else {
		logrus.Errorf("❌ Pixel Data element failed: %v", err)
	}
}

// addMandatoryElements adds mandatory DICOM metadata elements
func (w *Writer) addMandatoryElements(dataset *dicom.Dataset, image *types.Image) {
	// File Meta Information Group Length (0002,0000)
	if elem, err := dicom.NewElement(tag.FileMetaInformationGroupLength, []int{0}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Media Storage SOP Class UID (0002,0002)
	if elem, err := dicom.NewElement(tag.MediaStorageSOPClassUID, []string{image.SOPClassUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Media Storage SOP Instance UID (0002,0003)
	if elem, err := dicom.NewElement(tag.MediaStorageSOPInstanceUID, []string{image.SOPInstanceUID}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Transfer Syntax UID (0002,0010)
	if elem, err := dicom.NewElement(tag.TransferSyntaxUID, []string{"1.2.840.10008.1.2"}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

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

	// Add patient information
	w.addPatientElements(&dataset, study)

	// Add study information
	w.addStudyElements(&dataset, study)

	// Add series information
	w.addSeriesElements(&dataset, series)

	// Add image information
	w.addImageElements(&dataset, image)

	// Create DICOM file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create DICOM file: %w", err)
	}
	defer file.Close()

	// Write DICOM file
	if err := dicom.Write(file, dataset); err != nil {
		return fmt.Errorf("failed to write DICOM file: %w", err)
	}

	return nil
}

// addPatientElements adds patient-related DICOM elements
func (w *Writer) addPatientElements(dataset *dicom.Dataset, study *types.Study) {
	// Patient Name (0010,0010)
	if elem, err := dicom.NewElement(tag.PatientName, study.PatientName); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient ID (0010,0020)
	if elem, err := dicom.NewElement(tag.PatientID, study.PatientID); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient Birth Date (0010,0030)
	if elem, err := dicom.NewElement(tag.PatientBirthDate, study.PatientBirthDate); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Patient Sex (0010,0040) - Default to "O" (Other)
	if elem, err := dicom.NewElement(tag.PatientSex, "O"); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addStudyElements adds study-related DICOM elements
func (w *Writer) addStudyElements(dataset *dicom.Dataset, study *types.Study) {
	// Study Instance UID (0020,000D)
	if elem, err := dicom.NewElement(tag.StudyInstanceUID, study.StudyInstanceUID); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Date (0008,0020)
	if elem, err := dicom.NewElement(tag.StudyDate, study.StudyDate); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Time (0008,0030)
	if elem, err := dicom.NewElement(tag.StudyTime, study.StudyTime); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Study Description (0008,1030)
	if elem, err := dicom.NewElement(tag.StudyDescription, study.StudyDescription); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Accession Number (0008,0050)
	if elem, err := dicom.NewElement(tag.AccessionNumber, study.AccessionNumber); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addSeriesElements adds series-related DICOM elements
func (w *Writer) addSeriesElements(dataset *dicom.Dataset, series *types.Series) {
	// Series Instance UID (0020,000E)
	if elem, err := dicom.NewElement(tag.SeriesInstanceUID, series.SeriesInstanceUID); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Number (0020,0011)
	if elem, err := dicom.NewElement(tag.SeriesNumber, fmt.Sprintf("%d", series.SeriesNumber)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Modality (0008,0060)
	if elem, err := dicom.NewElement(tag.Modality, series.Modality); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Series Description (0008,103E)
	if elem, err := dicom.NewElement(tag.SeriesDescription, series.SeriesDescription); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addImageElements adds image-related DICOM elements
func (w *Writer) addImageElements(dataset *dicom.Dataset, image *types.Image) {
	// SOP Instance UID (0008,0018)
	if elem, err := dicom.NewElement(tag.SOPInstanceUID, image.SOPInstanceUID); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// SOP Class UID (0008,0016)
	if elem, err := dicom.NewElement(tag.SOPClassUID, image.SOPClassUID); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Instance Number (0020,0013)
	if elem, err := dicom.NewElement(tag.InstanceNumber, fmt.Sprintf("%d", image.InstanceNumber)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Image dimensions
	w.addImageDimensionElements(dataset, image)

	// Pixel data
	w.addPixelDataElements(dataset, image)
}

// addImageDimensionElements adds image dimension elements
func (w *Writer) addImageDimensionElements(dataset *dicom.Dataset, image *types.Image) {
	// Rows (0028,0010)
	if elem, err := dicom.NewElement(tag.Rows, uint16(image.Height)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Columns (0028,0011)
	if elem, err := dicom.NewElement(tag.Columns, uint16(image.Width)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Bits Allocated (0028,0100)
	if elem, err := dicom.NewElement(tag.BitsAllocated, uint16(image.BitsPerPixel)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Bits Stored (0028,0101)
	if elem, err := dicom.NewElement(tag.BitsStored, uint16(image.BitsPerPixel)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// High Bit (0028,0102)
	if elem, err := dicom.NewElement(tag.HighBit, uint16(image.BitsPerPixel-1)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Pixel Representation (0028,0103)
	if elem, err := dicom.NewElement(tag.PixelRepresentation, uint16(0)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Samples per Pixel (0028,0002)
	if elem, err := dicom.NewElement(tag.SamplesPerPixel, uint16(1)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Photometric Interpretation (0028,0004)
	if elem, err := dicom.NewElement(tag.PhotometricInterpretation, "MONOCHROME2"); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}

	// Planar Configuration (0028,0006)
	if elem, err := dicom.NewElement(tag.PlanarConfiguration, uint16(0)); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}

// addPixelDataElements adds pixel data elements
func (w *Writer) addPixelDataElements(dataset *dicom.Dataset, image *types.Image) {
	// Pixel Data (7FE0,0010)
	if elem, err := dicom.NewElement(tag.PixelData, image.PixelData); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
	}
}
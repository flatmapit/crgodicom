package dicom

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
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

// writeImage writes a single DICOM image to disk using DCMTK
func (w *Writer) writeImage(study *types.Study, series *types.Series, image *types.Image, filePath string) error {
	logrus.Infof("Writing DICOM file using DCMTK: %s", filePath)
	logrus.Infof("Image dimensions: %dx%d, %d bits per pixel, %d bytes pixel data", 
		image.Width, image.Height, image.BitsPerPixel, len(image.PixelData))

	// Create DICOM dataset
	dataset := dicom.NewDataset()

	// Add study-level elements
	w.addStudyElements(dataset, study)

	// Add series-level elements
	w.addSeriesElements(dataset, series)

	// Add image-level elements
	w.addImageElements(dataset, image)

	// Add pixel data elements
	w.addPixelDataElements(dataset, image)

	// Write DICOM file
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create DICOM file: %w", err)
	}
	defer file.Close()

	// Write DICOM dataset to file
	err = dicom.Write(file, dataset)
	if err != nil {
		return fmt.Errorf("failed to write DICOM dataset: %w", err)
	}

	logrus.Info("Successfully wrote DICOM file")
	return nil
}

// addStudyElements adds study-level DICOM elements
func (w *Writer) addStudyElements(dataset *dicom.Dataset, study *types.Study) {
	dataset.AddElement(dicom.NewElement(tag.StudyInstanceUID, study.StudyInstanceUID))
	dataset.AddElement(dicom.NewElement(tag.PatientName, study.PatientName))
	dataset.AddElement(dicom.NewElement(tag.PatientID, study.PatientID))
	dataset.AddElement(dicom.NewElement(tag.StudyDate, study.StudyDate))
	dataset.AddElement(dicom.NewElement(tag.StudyTime, study.StudyTime))
	dataset.AddElement(dicom.NewElement(tag.StudyDescription, study.StudyDescription))
	dataset.AddElement(dicom.NewElement(tag.AccessionNumber, study.AccessionNumber))
	dataset.AddElement(dicom.NewElement(tag.PatientBirthDate, study.PatientBirthDate))
}

// addSeriesElements adds series-level DICOM elements
func (w *Writer) addSeriesElements(dataset *dicom.Dataset, series *types.Series) {
	dataset.AddElement(dicom.NewElement(tag.SeriesInstanceUID, series.SeriesInstanceUID))
	dataset.AddElement(dicom.NewElement(tag.SeriesNumber, series.SeriesNumber))
	dataset.AddElement(dicom.NewElement(tag.Modality, series.Modality))
	dataset.AddElement(dicom.NewElement(tag.SeriesDescription, series.SeriesDescription))
}

// addImageElements adds image-level DICOM elements
func (w *Writer) addImageElements(dataset *dicom.Dataset, image *types.Image) {
	dataset.AddElement(dicom.NewElement(tag.SOPInstanceUID, image.SOPInstanceUID))
	dataset.AddElement(dicom.NewElement(tag.SOPClassUID, image.SOPClassUID))
	dataset.AddElement(dicom.NewElement(tag.InstanceNumber, image.InstanceNumber))
	dataset.AddElement(dicom.NewElement(tag.Rows, image.Height))
	dataset.AddElement(dicom.NewElement(tag.Columns, image.Width))
	dataset.AddElement(dicom.NewElement(tag.BitsAllocated, image.BitsPerPixel))
	dataset.AddElement(dicom.NewElement(tag.BitsStored, image.BitsPerPixel))
	dataset.AddElement(dicom.NewElement(tag.HighBit, image.BitsPerPixel-1))
	dataset.AddElement(dicom.NewElement(tag.SamplesPerPixel, 1))
	dataset.AddElement(dicom.NewElement(tag.PhotometricInterpretation, "MONOCHROME2"))
	dataset.AddElement(dicom.NewElement(tag.BurnedInAnnotation, "YES"))
	
	// Add windowing parameters
	dataset.AddElement(dicom.NewElement(tag.WindowCenter, "2048"))
	dataset.AddElement(dicom.NewElement(tag.WindowWidth, "4096"))
	dataset.AddElement(dicom.NewElement(tag.RescaleIntercept, "0"))
	dataset.AddElement(dicom.NewElement(tag.RescaleSlope, "1"))
}

// addPixelDataElements adds pixel data elements
func (w *Writer) addPixelDataElements(dataset *dicom.Dataset, image *types.Image) {
	logrus.Infof("Adding pixel data: %d bytes", len(image.PixelData))
	
	// Create pixel data info
	pixelDataInfo := dicom.PixelDataInfo{
		UnprocessedValueData: image.PixelData,
		IntentionallyUnprocessed: true,
	}
	
	dataset.AddElement(dicom.NewElement(tag.PixelData, pixelDataInfo))
}

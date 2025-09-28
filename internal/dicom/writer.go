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

	// Use DCMTK to write the DICOM file
	err := dcmtk.WriteDicomFile(
		filePath,
		study.PatientName,
		study.PatientID,
		study.StudyInstanceUID,
		series.SeriesInstanceUID,
		image.SOPInstanceUID,
		series.Modality,
		image.Width,
		image.Height,
		image.BitsPerPixel,
		image.BitsPerPixel,
		image.BitsPerPixel-1,
		1,
		"MONOCHROME2",
		image.PixelData,
	)

	if err != nil {
		return fmt.Errorf("failed to write DICOM file using DCMTK: %w", err)
	}

	logrus.Info("Successfully wrote DICOM file using DCMTK")
	return nil
}

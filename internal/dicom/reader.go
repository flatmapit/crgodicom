package dicom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
)

// Reader handles DICOM file reading
type Reader struct {
	config *config.Config
}

// NewReader creates a new DICOM reader
func NewReader(cfg *config.Config) *Reader {
	return &Reader{
		config: cfg,
	}
}

// DetailedStudyMetadata represents detailed study metadata
type DetailedStudyMetadata struct {
	StudyUID         string
	StudyInstanceUID string
	PatientName      string
	PatientID        string
	StudyDate        string
	StudyTime        string
	AccessionNumber  string
	PatientBirthDate string
	SeriesCount      int
	ImageCount       int
	Modality         string
	StudyDescription string
	SeriesDetails    []SeriesDetail
}

// SeriesDetail represents series metadata
type SeriesDetail struct {
	SeriesInstanceUID string
	SeriesNumber      int
	Modality          string
	SeriesDescription string
	ImageCount        int
	SeriesMetadata    SeriesMetadata
	ImageDetails      []ImageDetail
}

// SeriesMetadata represents series metadata fields
type SeriesMetadata struct {
	SeriesUID         string
	SeriesNumber      string
	Modality          string
	SeriesDescription string
}

// ImageDetail represents image metadata
type ImageDetail struct {
	SOPInstanceUID    string
	SOPClassUID       string
	InstanceNumber    string
	Width             int
	Height            int
	BitsPerPixel      int
	Modality          string
	PixelData         []byte
	PatientName       string
	PatientID         string
	StudyInstanceUID  string
	SeriesInstanceUID string
	StudyDate         string
	StudyTime         string
	StudyDescription  string
	SeriesDescription string
	AccessionNumber   string
	PatientBirthDate  string
}

// ReadStudy reads a DICOM study from disk
func (r *Reader) ReadStudy(studyUID string) (*types.Study, error) {
	// For now, return nil to indicate study not found
	// This will force the export process to use the in-memory study data
	return nil, fmt.Errorf("study not found: %s", studyUID)
}

// GetStudyMetadata gets basic study metadata
func (r *Reader) GetStudyMetadata(studyUID string) (*DetailedStudyMetadata, error) {
	// Placeholder implementation - will be implemented as needed
	return &DetailedStudyMetadata{}, nil
}

// ReadDetailedStudyMetadata reads detailed study metadata
func (r *Reader) ReadDetailedStudyMetadata(studyDir string) (*DetailedStudyMetadata, error) {
	logrus.Infof("=== ReadDetailedStudyMetadata called for: %s ===", studyDir)

	// Check if study directory exists
	if _, err := os.Stat(studyDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("study directory not found: %s", studyDir)
	}

	metadata := &DetailedStudyMetadata{
		SeriesDetails: make([]SeriesDetail, 0),
	}

	// Read series directories
	entries, err := os.ReadDir(studyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read study directory: %w", err)
	}

	logrus.Infof("Found %d entries in study directory", len(entries))
	for _, entry := range entries {
		logrus.Infof("Processing entry: %s (isDir: %v)", entry.Name(), entry.IsDir())
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "series_") {
			logrus.Infof("Skipping entry %s (not a series directory)", entry.Name())
			continue
		}

		seriesDir := filepath.Join(studyDir, entry.Name())
		logrus.Infof("Reading series metadata from: %s", seriesDir)
		seriesDetail, err := r.readSeriesMetadataWithDCMTK(seriesDir)
		if err != nil {
			logrus.Warnf("Failed to read series %s: %v", entry.Name(), err)
			continue
		}

		metadata.SeriesDetails = append(metadata.SeriesDetails, *seriesDetail)
		metadata.SeriesCount++
		metadata.ImageCount += seriesDetail.ImageCount
	}

	// If we found series, read study metadata from the first DICOM file
	if len(metadata.SeriesDetails) > 0 && len(metadata.SeriesDetails[0].ImageDetails) > 0 {
		firstImagePath := filepath.Join(studyDir, "series_001", "image_001.dcm")
		if err := r.readStudyMetadataFromFile(firstImagePath, metadata); err != nil {
			logrus.Warnf("Failed to read study metadata from file: %v", err)
		}
	}

	logrus.Infof("Read study metadata: %d series, %d images", metadata.SeriesCount, metadata.ImageCount)
	return metadata, nil
}

// ReadStudyMetadata reads study metadata
func (r *Reader) ReadStudyMetadata(studyPath string) (*DetailedStudyMetadata, error) {
	// Use ReadDetailedStudyMetadata to read actual DICOM files
	return r.ReadDetailedStudyMetadata(studyPath)
}

// readSeriesMetadata reads metadata for a single series
func (r *Reader) readSeriesMetadata(seriesDir string) (*SeriesDetail, error) {
	logrus.Debugf("Reading series metadata from %s", seriesDir)

	seriesDetail := &SeriesDetail{
		ImageDetails: make([]ImageDetail, 0),
	}

	// Read DICOM files in series directory
	entries, err := os.ReadDir(seriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read series directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".dcm") {
			continue
		}

		imagePath := filepath.Join(seriesDir, entry.Name())
		imageDetail, err := r.readImageMetadataWithDCMTK(imagePath)
		if err != nil {
			logrus.Warnf("Failed to read image %s: %v", entry.Name(), err)
			continue
		}

		seriesDetail.ImageDetails = append(seriesDetail.ImageDetails, *imageDetail)
		seriesDetail.ImageCount++

		// Use first image to populate series metadata
		if seriesDetail.ImageCount == 1 {
			seriesDetail.SeriesMetadata = SeriesMetadata{
				SeriesUID:         imageDetail.SOPInstanceUID, // Use instance UID as series UID for now
				SeriesNumber:      "1",                        // Default series number
				Modality:          imageDetail.Modality,
				SeriesDescription: fmt.Sprintf("%s Series", imageDetail.Modality),
			}
		}
	}

	logrus.Debugf("Read series metadata: %d images", seriesDetail.ImageCount)
	return seriesDetail, nil
}

// readSeriesMetadataWithDCMTK reads metadata for a single series using DCMTK
func (r *Reader) readSeriesMetadataWithDCMTK(seriesDir string) (*SeriesDetail, error) {
	logrus.Debugf("Reading series metadata from %s using DCMTK", seriesDir)

	seriesDetail := &SeriesDetail{
		ImageDetails: make([]ImageDetail, 0),
	}

	// Read DICOM files in series directory
	entries, err := os.ReadDir(seriesDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read series directory: %w", err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".dcm") {
			continue
		}

		imagePath := filepath.Join(seriesDir, entry.Name())
		imageDetail, err := r.readImageMetadataWithDCMTK(imagePath)
		if err != nil {
			logrus.Warnf("Failed to read image %s: %v", entry.Name(), err)
			continue
		}

		seriesDetail.ImageDetails = append(seriesDetail.ImageDetails, *imageDetail)
		seriesDetail.ImageCount++

		// Use first image to populate series metadata
		if seriesDetail.ImageCount == 1 {
			seriesDetail.SeriesMetadata = SeriesMetadata{
				SeriesUID:         imageDetail.SeriesInstanceUID,
				SeriesNumber:      "1", // Default series number
				Modality:          imageDetail.Modality,
				SeriesDescription: fmt.Sprintf("%s Series", imageDetail.Modality),
			}
		}
	}

	logrus.Debugf("Read series metadata using DCMTK: %d images", seriesDetail.ImageCount)
	return seriesDetail, nil
}

// readImageMetadataWithDCMTK reads metadata from a single DICOM file using DCMTK
func (r *Reader) readImageMetadataWithDCMTK(filePath string) (*ImageDetail, error) {
	logrus.Infof("Reading image metadata from %s using DCMTK", filePath)

	// Use DCMTK to read the DICOM file
	dcmtkMetadata, err := dcmtk.ReadDicomFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read DICOM file with DCMTK: %w", err)
	}

	// Convert DCMTK metadata to our internal format
	imageDetail := &ImageDetail{
		SOPInstanceUID:    dcmtkMetadata.InstanceUID,
		SOPClassUID:       dcmtkMetadata.SOPClassUID,
		InstanceNumber:    "1", // Default instance number
		Modality:          dcmtkMetadata.Modality,
		Width:             dcmtkMetadata.Width,
		Height:            dcmtkMetadata.Height,
		BitsPerPixel:      dcmtkMetadata.BitsPerPixel,
		PixelData:         dcmtkMetadata.PixelData,
		PatientName:       dcmtkMetadata.PatientName,
		PatientID:         dcmtkMetadata.PatientID,
		StudyInstanceUID:  dcmtkMetadata.StudyUID,
		SeriesInstanceUID: dcmtkMetadata.SeriesUID,
		StudyDate:         dcmtkMetadata.StudyDate,
		StudyTime:         dcmtkMetadata.StudyTime,
		StudyDescription:  dcmtkMetadata.StudyDescription,
		SeriesDescription: dcmtkMetadata.SeriesDescription,
		AccessionNumber:   "", // Not extracted in DCMTK reader yet
		PatientBirthDate:  "", // Not extracted in DCMTK reader yet
	}

	// Set defaults if not found
	if imageDetail.Width == 0 {
		imageDetail.Width = 512 // Default width
	}
	if imageDetail.Height == 0 {
		imageDetail.Height = 512 // Default height
	}
	if imageDetail.BitsPerPixel == 0 {
		imageDetail.BitsPerPixel = 16 // Default bits per pixel
	}
	if imageDetail.Modality == "" {
		imageDetail.Modality = "CR" // Default modality
	}

	logrus.Debugf("Read image metadata using DCMTK: %dx%d, %d bits, %d bytes pixel data",
		imageDetail.Width, imageDetail.Height, imageDetail.BitsPerPixel, len(imageDetail.PixelData))

	return imageDetail, nil
}

// readStudyMetadataFromFile reads study-level metadata from a DICOM file using DCMTK
func (r *Reader) readStudyMetadataFromFile(filePath string, metadata *DetailedStudyMetadata) error {
	logrus.Debugf("Reading study metadata from %s using DCMTK", filePath)

	// Use DCMTK to read the DICOM file
	dcmtkMetadata, err := dcmtk.ReadDicomFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read DICOM file with DCMTK: %w", err)
	}

	// Extract study-level metadata from DCMTK metadata
	metadata.StudyInstanceUID = dcmtkMetadata.StudyUID
	metadata.StudyUID = dcmtkMetadata.StudyUID
	metadata.PatientName = dcmtkMetadata.PatientName
	metadata.PatientID = dcmtkMetadata.PatientID
	metadata.StudyDate = dcmtkMetadata.StudyDate
	metadata.StudyTime = dcmtkMetadata.StudyTime
	metadata.AccessionNumber = ""  // Not extracted in DCMTK reader yet
	metadata.PatientBirthDate = "" // Not extracted in DCMTK reader yet
	metadata.StudyDescription = dcmtkMetadata.StudyDescription

	return nil
}

// ExtractSOPInstanceUID extracts SOP Instance UID from DICOM file using DCMTK
func (r *Reader) ExtractSOPInstanceUID(filePath string) (string, error) {
	// Use DCMTK to read the DICOM file
	dcmtkMetadata, err := dcmtk.ReadDicomFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read DICOM file with DCMTK: %w", err)
	}

	if dcmtkMetadata.InstanceUID == "" {
		return "", fmt.Errorf("SOP Instance UID not found in DICOM file")
	}

	return dcmtkMetadata.InstanceUID, nil
}

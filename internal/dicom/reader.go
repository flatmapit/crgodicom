package dicom

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
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
	SOPInstanceUID string
	SOPClassUID    string
	InstanceNumber string
	Width          int
	Height         int
	BitsPerPixel   int
	Modality       string
	PixelData      []byte
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
	logrus.Infof("Reading detailed study metadata from %s", studyDir)

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

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasPrefix(entry.Name(), "series_") {
			continue
		}

		seriesDir := filepath.Join(studyDir, entry.Name())
		seriesDetail, err := r.readSeriesMetadata(seriesDir)
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
func (r *Reader) ReadStudyMetadata(studyUID string) (*DetailedStudyMetadata, error) {
	// Placeholder implementation - will be implemented as needed
	return &DetailedStudyMetadata{}, nil
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
		imageDetail, err := r.readImageMetadata(imagePath)
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

// readImageMetadata reads metadata from a single DICOM file
func (r *Reader) readImageMetadata(filePath string) (*ImageDetail, error) {
	logrus.Debugf("Reading image metadata from %s", filePath)

	// Open DICOM file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open DICOM file: %w", err)
	}
	defer file.Close()

	// Get file size for parsing
	fileInfo, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Parse DICOM file
	dataset, err := dicom.Parse(file, fileInfo.Size(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DICOM file: %w", err)
	}

	imageDetail := &ImageDetail{}

	// Extract metadata from DICOM elements
	for _, elem := range dataset.Elements {
		switch elem.Tag {
		case tag.SOPInstanceUID:
			if value, ok := elem.Value.GetValue().(string); ok {
				imageDetail.SOPInstanceUID = value
			}
		case tag.SOPClassUID:
			if value, ok := elem.Value.GetValue().(string); ok {
				imageDetail.SOPClassUID = value
			}
		case tag.InstanceNumber:
			if value, ok := elem.Value.GetValue().(string); ok {
				imageDetail.InstanceNumber = value
			}
		case tag.Modality:
			if value, ok := elem.Value.GetValue().(string); ok {
				imageDetail.Modality = value
			}
		case tag.Rows:
			if value, ok := elem.Value.GetValue().(int); ok {
				imageDetail.Height = value
			}
		case tag.Columns:
			if value, ok := elem.Value.GetValue().(int); ok {
				imageDetail.Width = value
			}
		case tag.BitsAllocated:
			if value, ok := elem.Value.GetValue().(int); ok {
				imageDetail.BitsPerPixel = value
			}
		case tag.PixelData:
			if value, ok := elem.Value.GetValue().([]byte); ok {
				imageDetail.PixelData = value
			}
		}
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

	logrus.Debugf("Read image metadata: %dx%d, %d bits, %d bytes pixel data",
		imageDetail.Width, imageDetail.Height, imageDetail.BitsPerPixel, len(imageDetail.PixelData))

	return imageDetail, nil
}

// readStudyMetadataFromFile reads study-level metadata from a DICOM file
func (r *Reader) readStudyMetadataFromFile(filePath string, metadata *DetailedStudyMetadata) error {
	logrus.Debugf("Reading study metadata from %s", filePath)

	// Open DICOM file
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open DICOM file: %w", err)
	}
	defer file.Close()

	// Get file size for parsing
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	// Parse DICOM file
	dataset, err := dicom.Parse(file, fileInfo.Size(), nil)
	if err != nil {
		return fmt.Errorf("failed to parse DICOM file: %w", err)
	}

	// Extract study-level metadata
	for _, elem := range dataset.Elements {
		switch elem.Tag {
		case tag.StudyInstanceUID:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.StudyInstanceUID = value
				metadata.StudyUID = value
			}
		case tag.PatientName:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.PatientName = value
			}
		case tag.PatientID:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.PatientID = value
			}
		case tag.StudyDate:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.StudyDate = value
			}
		case tag.StudyTime:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.StudyTime = value
			}
		case tag.AccessionNumber:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.AccessionNumber = value
			}
		case tag.PatientBirthDate:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.PatientBirthDate = value
			}
		case tag.StudyDescription:
			if value, ok := elem.Value.GetValue().(string); ok {
				metadata.StudyDescription = value
			}
		}
	}

	return nil
}

// ExtractSOPInstanceUID extracts SOP Instance UID from DICOM file
func (r *Reader) ExtractSOPInstanceUID(filePath string) (string, error) {
	// Open DICOM file
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to open DICOM file: %w", err)
	}
	defer file.Close()

	// Get file size for parsing
	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("failed to get file info: %w", err)
	}

	// Parse DICOM file
	dataset, err := dicom.Parse(file, fileInfo.Size(), nil)
	if err != nil {
		return "", fmt.Errorf("failed to parse DICOM file: %w", err)
	}

	// Extract SOP Instance UID
	for _, elem := range dataset.Elements {
		if elem.Tag == tag.SOPInstanceUID {
			if value, ok := elem.Value.GetValue().(string); ok {
				return value, nil
			}
		}
	}

	return "", fmt.Errorf("SOP Instance UID not found in DICOM file")
}

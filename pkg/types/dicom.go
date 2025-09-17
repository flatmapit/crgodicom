package types

import (
	"time"
)

// DICOMTypes defines common DICOM-related types and constants

// SOPClassUIDs defines the standard SOP Class UIDs for different modalities
var SOPClassUIDs = map[string]string{
	"CR": "1.2.840.10008.5.1.4.1.1.1",     // Computed Radiography Image Storage
	"CT": "1.2.840.10008.5.1.4.1.1.2",     // CT Image Storage
	"MR": "1.2.840.10008.5.1.4.1.1.4",     // MR Image Storage
	"US": "1.2.840.10008.5.1.4.1.1.6",     // Ultrasound Image Storage
	"DX": "1.2.840.10008.5.1.4.1.1.1.1",   // Digital X-Ray Image Storage
	"MG": "1.2.840.10008.5.1.4.1.1.1.2",   // Digital Mammography X-Ray Image Storage
}

// TransferSyntaxUIDs defines common transfer syntax UIDs
var TransferSyntaxUIDs = map[string]string{
	"ImplicitVRLittleEndian": "1.2.840.10008.1.2",
	"ExplicitVRLittleEndian": "1.2.840.10008.1.2.1",
}

// ImageDimensions defines standard image dimensions by modality
var ImageDimensions = map[string]ImageSize{
	"CR": {Width: 2048, Height: 2048, BitsPerPixel: 16},
	"CT": {Width: 512, Height: 512, BitsPerPixel: 16},
	"MR": {Width: 256, Height: 256, BitsPerPixel: 16},
	"US": {Width: 640, Height: 480, BitsPerPixel: 8},
	"DX": {Width: 2048, Height: 2048, BitsPerPixel: 16},
	"MG": {Width: 4096, Height: 3328, BitsPerPixel: 16},
}

// ImageSize represents image dimensions and bit depth
type ImageSize struct {
	Width        int
	Height       int
	BitsPerPixel int
}

// Study represents a DICOM study
type Study struct {
	StudyInstanceUID string
	StudyDate        string
	StudyTime        string
	AccessionNumber  string
	StudyDescription string
	PatientName      string
	PatientID        string
	PatientBirthDate string
	Series           []Series
}

// Series represents a DICOM series
type Series struct {
	SeriesInstanceUID string
	SeriesNumber      int
	Modality          string
	SeriesDescription string
	Images            []Image
}

// Image represents a DICOM image
type Image struct {
	SOPInstanceUID string
	SOPClassUID    string
	InstanceNumber int
	PixelData      []byte
	Width          int
	Height         int
	BitsPerPixel   int
	Modality       string
}

// PatientInfo represents patient information
type PatientInfo struct {
	Name      string
	ID        string
	BirthDate time.Time
}

// StudyInfo represents study information
type StudyInfo struct {
	Description   string
	AccessionNumber string
	Date         time.Time
	Time         time.Time
}

// SeriesInfo represents series information
type SeriesInfo struct {
	Number      int
	Description string
	Modality    string
}

// ImageInfo represents image information
type ImageInfo struct {
	Number    int
	Width     int
	Height    int
	BitsPerPixel int
}

// DICOMField represents a DICOM data element
type DICOMField struct {
	Tag   string
	VR    string // Value Representation
	Value interface{}
}

// UIDGenerator interface for generating DICOM UIDs
type UIDGenerator interface {
	GenerateStudyUID() string
	GenerateSeriesUID() string
	GenerateInstanceUID() string
}

// ImageGenerator interface for generating synthetic images
type ImageGenerator interface {
	GenerateImage(modality string, width, height, bitsPerPixel int) ([]byte, error)
}

// StudyGenerator interface for generating complete studies
type StudyGenerator interface {
	GenerateStudy(params StudyParams) (*Study, error)
}

// StudyParams represents parameters for study generation
type StudyParams struct {
	StudyCount       int
	SeriesCount      int
	ImageCount       int
	Modality         string
	AnatomicalRegion string
	PatientName      string
	PatientID        string
	AccessionNumber  string
	StudyDescription string
	OutputDir        string
	Template         interface{} // Template configuration
}

// ValidationError represents a DICOM validation error
type ValidationError struct {
	Field string
	Value interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

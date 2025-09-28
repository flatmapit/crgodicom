package dicom

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/sirupsen/logrus"
)

// EnhancedUIDGenerator generates cryptographically secure DICOM UIDs
type EnhancedUIDGenerator struct {
	orgRoot string
}

// NewEnhancedUIDGenerator creates a new enhanced UID generator
func NewEnhancedUIDGenerator(orgRoot string) *EnhancedUIDGenerator {
	return &EnhancedUIDGenerator{
		orgRoot: orgRoot,
	}
}

// GenerateStudyUID generates a cryptographically secure study UID
func (g *EnhancedUIDGenerator) GenerateStudyUID() string {
	return g.generateSecureUID(g.orgRoot)
}

// GenerateSeriesUID generates a cryptographically secure series UID
func (g *EnhancedUIDGenerator) GenerateSeriesUID() string {
	return g.generateSecureUID(g.orgRoot)
}

// GenerateInstanceUID generates a cryptographically secure instance UID
func (g *EnhancedUIDGenerator) GenerateInstanceUID() string {
	return g.generateSecureUID(g.orgRoot)
}

// generateSecureUID generates a cryptographically secure UID
func (g *EnhancedUIDGenerator) generateSecureUID(orgRoot string) string {
	// Generate a large random number
	max := big.NewInt(1 << 62) // 2^62
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		logrus.Warnf("Failed to generate secure random number, falling back to timestamp: %v", err)
		return fmt.Sprintf("%s.%d", orgRoot, time.Now().UnixNano())
	}

	return fmt.Sprintf("%s.%d", orgRoot, n.Int64())
}

// MetadataGenerator generates comprehensive DICOM metadata
type MetadataGenerator struct {
	orgRoot string
}

// NewMetadataGenerator creates a new metadata generator
func NewMetadataGenerator(orgRoot string) *MetadataGenerator {
	return &MetadataGenerator{
		orgRoot: orgRoot,
	}
}

// PatientModule generates patient metadata
func (m *MetadataGenerator) PatientModule(patientName, patientID string) map[string]interface{} {
	return map[string]interface{}{
		"PatientName":      patientName,
		"PatientID":        patientID,
		"PatientBirthDate": "19500101", // Default birth date
	}
}

// StudyModule generates study metadata
func (m *MetadataGenerator) StudyModule(studyDescription, accessionNumber string) map[string]interface{} {
	now := time.Now()
	return map[string]interface{}{
		"StudyInstanceUID": m.generateSecureUID(m.orgRoot),
		"StudyDate":        now.Format("20060102"),
		"StudyTime":        now.Format("150405"),
		"AccessionNumber":  accessionNumber,
		"StudyDescription": studyDescription,
	}
}

// SeriesModule generates series metadata
func (m *MetadataGenerator) SeriesModule(modality string, seriesNumber int, anatomicalRegion string) map[string]interface{} {
	return map[string]interface{}{
		"SeriesInstanceUID": m.generateSecureUID(m.orgRoot),
		"SeriesNumber":      seriesNumber,
		"Modality":          modality,
		"SeriesDescription": fmt.Sprintf("%s %s", modality, anatomicalRegion),
	}
}

// ImageModule generates image metadata
func (m *MetadataGenerator) ImageModule(modality string, instanceNumber int, sopClassUID string) map[string]interface{} {
	return map[string]interface{}{
		"SOPInstanceUID": m.generateSecureUID(m.orgRoot),
		"SOPClassUID":    sopClassUID,
		"InstanceNumber": instanceNumber,
	}
}

// ImagePixelModule generates image pixel metadata
func (m *MetadataGenerator) ImagePixelModule(width, height, bitsPerPixel int, modality string) map[string]interface{} {
	return map[string]interface{}{
		"Columns":                   width,
		"Rows":                      height,
		"BitsAllocated":             bitsPerPixel,
		"BitsStored":                bitsPerPixel,
		"HighBit":                   bitsPerPixel - 1,
		"PixelRepresentation":       0, // Unsigned
		"PhotometricInterpretation": "MONOCHROME2",
		"SamplesPerPixel":           1,
	}
}

// generateSecureUID generates a cryptographically secure UID
func (m *MetadataGenerator) generateSecureUID(orgRoot string) string {
	// Generate a large random number
	max := big.NewInt(1 << 62) // 2^62
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		logrus.Warnf("Failed to generate secure random number, falling back to timestamp: %v", err)
		return fmt.Sprintf("%s.%d", orgRoot, time.Now().UnixNano())
	}

	return fmt.Sprintf("%s.%d", orgRoot, n.Int64())
}

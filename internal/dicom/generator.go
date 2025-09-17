package dicom

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/pkg/types"
)

// Generator handles DICOM data generation
type Generator struct {
	config      *config.Config
	uidGen      *UIDGenerator
	imageGen    *ImageGenerator
}

// NewGenerator creates a new DICOM generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config: cfg,
		uidGen: NewUIDGenerator(cfg.DICOM.OrgRoot),
		imageGen: NewImageGenerator(),
	}
}

// UIDGenerator generates DICOM UIDs
type UIDGenerator struct {
	orgRoot string
	rand    *rand.Rand
}

// ImageGenerator generates synthetic images
type ImageGenerator struct {
	rand *rand.Rand
}

// NewUIDGenerator creates a new UID generator
func NewUIDGenerator(orgRoot string) *UIDGenerator {
	return &UIDGenerator{
		orgRoot: orgRoot,
		rand:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// NewImageGenerator creates a new image generator
func NewImageGenerator() *ImageGenerator {
	return &ImageGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateStudy generates a complete DICOM study
func (g *Generator) GenerateStudy(params types.StudyParams) (*types.Study, error) {
	// Generate patient information
	patientInfo := g.generatePatientInfo(params.PatientName, params.PatientID)
	
	// Generate study information
	studyInfo := g.generateStudyInfo(params.StudyDescription, params.AccessionNumber)
	
	// Generate study UID
	studyUID := g.uidGen.GenerateStudyUID()
	
	// Create study
	study := &types.Study{
		StudyInstanceUID: studyUID,
		StudyDate:        studyInfo.Date.Format("20060102"),
		StudyTime:        studyInfo.Date.Format("150405"),
		AccessionNumber:  studyInfo.AccessionNumber,
		StudyDescription: studyInfo.Description,
		PatientName:      patientInfo.Name,
		PatientID:        patientInfo.ID,
		PatientBirthDate: patientInfo.BirthDate.Format("20060102"),
		Series:           make([]types.Series, 0, params.SeriesCount),
	}
	
	// Generate series
	for i := 0; i < params.SeriesCount; i++ {
		series, err := g.generateSeries(studyUID, params.Modality, i+1, params.ImageCount)
		if err != nil {
			return nil, fmt.Errorf("failed to generate series %d: %w", i+1, err)
		}
		study.Series = append(study.Series, *series)
	}
	
	return study, nil
}

// generateSeries generates a DICOM series
func (g *Generator) generateSeries(studyUID, modality string, seriesNumber, imageCount int) (*types.Series, error) {
	seriesUID := g.uidGen.GenerateSeriesUID()
	
	series := &types.Series{
		SeriesInstanceUID: seriesUID,
		SeriesNumber:      seriesNumber,
		Modality:          modality,
		SeriesDescription: fmt.Sprintf("%s Series %d", modality, seriesNumber),
		Images:            make([]types.Image, 0, imageCount),
	}
	
	// Generate images
	for i := 0; i < imageCount; i++ {
		image, err := g.generateImage(studyUID, seriesUID, modality, i+1)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image %d: %w", i+1, err)
		}
		series.Images = append(series.Images, *image)
	}
	
	return series, nil
}

// generateImage generates a DICOM image
func (g *Generator) generateImage(studyUID, seriesUID, modality string, instanceNumber int) (*types.Image, error) {
	instanceUID := g.uidGen.GenerateInstanceUID()
	
	// Get SOP class UID for modality
	sopClassUID, exists := types.SOPClassUIDs[modality]
	if !exists {
		return nil, fmt.Errorf("unsupported modality: %s", modality)
	}
	
	// Get image dimensions
	imageSize, exists := types.ImageDimensions[modality]
	if !exists {
		return nil, fmt.Errorf("unsupported modality: %s", modality)
	}
	
	// Generate pixel data
	pixelData, err := g.imageGen.GenerateImage(modality, imageSize.Width, imageSize.Height, imageSize.BitsPerPixel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pixel data: %w", err)
	}
	
	image := &types.Image{
		SOPInstanceUID: instanceUID,
		SOPClassUID:    sopClassUID,
		InstanceNumber: instanceNumber,
		PixelData:      pixelData,
		Width:          imageSize.Width,
		Height:         imageSize.Height,
		BitsPerPixel:   imageSize.BitsPerPixel,
		Modality:       modality,
	}
	
	return image, nil
}

// generatePatientInfo generates patient information
func (g *Generator) generatePatientInfo(patientName, patientID string) types.PatientInfo {
	// Use provided values or generate defaults
	if patientName == "" {
		patientName = "DOE^JOHN^M"
	}
	if patientID == "" {
		patientID = g.generateRandomPatientID()
	}
	
	// Generate random birth date (18-80 years old)
	now := time.Now()
	minAge := 18
	maxAge := 80
	age := g.uidGen.rand.Intn(maxAge-minAge+1) + minAge
	birthYear := now.Year() - age
	birthDate := time.Date(birthYear, time.Month(g.uidGen.rand.Intn(12)+1), g.uidGen.rand.Intn(28)+1, 0, 0, 0, 0, time.UTC)
	
	return types.PatientInfo{
		Name:      patientName,
		ID:        patientID,
		BirthDate: birthDate,
	}
}

// generateStudyInfo generates study information
func (g *Generator) generateStudyInfo(studyDescription, accessionNumber string) types.StudyInfo {
	// Use provided values or generate defaults
	if studyDescription == "" {
		studyDescription = "Generated Study"
	}
	if accessionNumber == "" {
		accessionNumber = g.generateAccessionNumber()
	}
	
	now := time.Now()
	
	return types.StudyInfo{
		Description:     studyDescription,
		AccessionNumber: accessionNumber,
		Date:           now,
		Time:           now,
	}
}

// generateRandomPatientID generates a random patient ID
func (g *Generator) generateRandomPatientID() string {
	const chars = "0123456789ABCDEF"
	patientID := make([]byte, 8)
	for i := range patientID {
		patientID[i] = chars[g.uidGen.rand.Intn(len(chars))]
	}
	return string(patientID)
}

// generateAccessionNumber generates an accession number
func (g *Generator) generateAccessionNumber() string {
	now := time.Now()
	return fmt.Sprintf("%s-%04d", now.Format("20060102"), g.uidGen.rand.Intn(10000))
}

// GenerateStudyUID generates a study instance UID
func (u *UIDGenerator) GenerateStudyUID() string {
	timestamp := time.Now().Unix()
	random := u.rand.Int63()
	return fmt.Sprintf("%s.%d.%d", u.orgRoot, timestamp, random)
}

// GenerateSeriesUID generates a series instance UID
func (u *UIDGenerator) GenerateSeriesUID() string {
	timestamp := time.Now().Unix()
	random := u.rand.Int63()
	return fmt.Sprintf("%s.%d.%d", u.orgRoot, timestamp, random)
}

// GenerateInstanceUID generates a SOP instance UID
func (u *UIDGenerator) GenerateInstanceUID() string {
	timestamp := time.Now().Unix()
	random := u.rand.Int63()
	return fmt.Sprintf("%s.%d.%d", u.orgRoot, timestamp, random)
}

// GenerateImage generates synthetic image data
func (i *ImageGenerator) GenerateImage(modality string, width, height, bitsPerPixel int) ([]byte, error) {
	// Calculate bytes per pixel
	bytesPerPixel := bitsPerPixel / 8
	if bitsPerPixel%8 != 0 {
		bytesPerPixel++
	}
	
	// Create pixel data buffer
	pixelData := make([]byte, width*height*bytesPerPixel)
	
	// Generate noise pattern based on modality
	switch modality {
	case "CR", "DX":
		// X-ray: high contrast, more structured noise
		i.generateXRayPattern(pixelData, width, height, bytesPerPixel)
	case "CT":
		// CT: moderate contrast, slice-like patterns
		i.generateCTPattern(pixelData, width, height, bytesPerPixel)
	case "MR":
		// MRI: high contrast, more uniform noise
		i.generateMRPattern(pixelData, width, height, bytesPerPixel)
	case "US":
		// Ultrasound: low contrast, speckle noise
		i.generateUSPattern(pixelData, width, height, bytesPerPixel)
	case "MG":
		// Mammography: high resolution, subtle patterns
		i.generateMGPattern(pixelData, width, height, bytesPerPixel)
	default:
		// Default: simple noise pattern
		i.generateDefaultPattern(pixelData, width, height, bytesPerPixel)
	}
	
	return pixelData, nil
}

// generateXRayPattern generates X-ray-like noise pattern
func (i *ImageGenerator) generateXRayPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// X-ray characteristics: high contrast, some anatomical-like structures
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			
			// Base noise
			noise := i.rand.Intn(256)
			
			// Add some structured patterns (simulate ribs, etc.)
			if (x+y)%50 < 10 {
				noise += 30
			}
			if (x-y)%80 < 15 {
				noise -= 20
			}
			
			// Ensure values are in valid range
			if noise < 0 {
				noise = 0
			}
			if noise > 255 {
				noise = 255
			}
			
			// Store pixel value
			if bytesPerPixel == 2 {
				// 16-bit
				value := uint16(noise) * 256 // Scale to 16-bit range
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				// 8-bit
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateCTPattern generates CT-like noise pattern
func (i *ImageGenerator) generateCTPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// CT characteristics: moderate contrast, slice-like patterns
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			
			// Base noise
			noise := i.rand.Intn(256)
			
			// Add circular patterns (simulate cross-sections)
			centerX, centerY := width/2, height/2
			dist := (x-centerX)*(x-centerX) + (y-centerY)*(y-centerY)
			if dist < (width*height)/8 {
				noise += 40
			}
			
			// Ensure values are in valid range
			if noise < 0 {
				noise = 0
			}
			if noise > 255 {
				noise = 255
			}
			
			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(noise) * 256
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateMRPattern generates MRI-like noise pattern
func (i *ImageGenerator) generateMRPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// MRI characteristics: high contrast, more uniform noise
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			
			// Base noise with higher contrast
			noise := i.rand.Intn(256)
			
			// Add some regional variations
			region := (x/64 + y/64) % 4
			switch region {
			case 0:
				noise += 20
			case 1:
				noise -= 20
			case 2:
				noise += 40
			case 3:
				noise -= 40
			}
			
			// Ensure values are in valid range
			if noise < 0 {
				noise = 0
			}
			if noise > 255 {
				noise = 255
			}
			
			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(noise) * 256
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateUSPattern generates ultrasound-like noise pattern
func (i *ImageGenerator) generateUSPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// Ultrasound characteristics: low contrast, speckle noise
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			
			// Base noise with lower contrast
			noise := i.rand.Intn(128) + 64 // Range 64-191
			
			// Add speckle noise (characteristic of ultrasound)
			if i.rand.Intn(10) < 2 {
				noise += 50
			}
			
			// Ensure values are in valid range
			if noise < 0 {
				noise = 0
			}
			if noise > 255 {
				noise = 255
			}
			
			// Store pixel value
			pixelData[idx] = byte(noise)
		}
	}
}

// generateMGPattern generates mammography-like noise pattern
func (i *ImageGenerator) generateMGPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// Mammography characteristics: high resolution, subtle patterns
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			
			// Base noise with subtle variations
			noise := i.rand.Intn(256)
			
			// Add very subtle patterns (simulate breast tissue)
			if (x+y)%100 < 5 {
				noise += 10
			}
			if (x-y)%150 < 8 {
				noise -= 10
			}
			
			// Ensure values are in valid range
			if noise < 0 {
				noise = 0
			}
			if noise > 255 {
				noise = 255
			}
			
			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(noise) * 256
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateDefaultPattern generates default noise pattern
func (i *ImageGenerator) generateDefaultPattern(pixelData []byte, width, height, bytesPerPixel int) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel
			noise := i.rand.Intn(256)
			
			if bytesPerPixel == 2 {
				value := uint16(noise) * 256
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

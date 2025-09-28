package dicom

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
)

// Generator handles DICOM data generation
type Generator struct {
	config   *config.Config
	uidGen   *UIDGenerator
	imageGen *ImageGenerator
}

// NewGenerator creates a new DICOM generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config:   cfg,
		uidGen:   NewUIDGenerator(cfg.DICOM.OrgRoot),
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
	fmt.Printf("DEBUG: generateImage - modality='%s', imageSize=%+v\n", modality, imageSize)
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
		Date:            now,
		Time:            now,
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
	totalBytes := width * height * bytesPerPixel
	fmt.Printf("DEBUG: GenerateImage - width=%d, height=%d, bitsPerPixel=%d, bytesPerPixel=%d, totalBytes=%d\n",
		width, height, bitsPerPixel, bytesPerPixel, totalBytes)
	pixelData := make([]byte, totalBytes)

	// Generate noise pattern based on modality
	switch modality {
	case "CR", "DX":
		// X-ray: high contrast, more structured noise
		logrus.Info("Generating X-ray pattern for modality: " + modality)
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

// generateXRayPattern generates a greyscale spiral pattern with modality text
func (i *ImageGenerator) generateXRayPattern(pixelData []byte, width, height, bytesPerPixel int) {
	logrus.Info("Starting spiral pattern generation")

	// Calculate center point
	centerX := width / 2
	centerY := height / 2

	// Generate spiral pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Check if this pixel should be part of modality text
			if i.isModalityTextPixel(x, y, width, height) {
				// Bright white text
				if bytesPerPixel == 2 {
					// 16-bit maximum value (65535)
					pixelData[idx] = 0xFF
					pixelData[idx+1] = 0xFF
				} else {
					// 8-bit maximum value
					pixelData[idx] = 0xFF
				}
			} else {
				// Generate spiral pattern
				// Calculate distance from center
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				distance := math.Sqrt(dx*dx + dy*dy)

				// Calculate angle
				angle := math.Atan2(dy, dx)

				// Create spiral pattern: distance increases with angle
				spiralValue := (angle + math.Pi) / (2 * math.Pi) // Normalize to 0-1
				spiralValue = spiralValue*10 + distance/50       // Scale and add distance component

				// Convert to grayscale value (0-255)
				spiralValue = math.Mod(spiralValue, 1.0) // Keep in 0-1 range
				grayValue := int(spiralValue * 255)

				// Ensure values are in valid range
				if grayValue < 0 {
					grayValue = 0
				}
				if grayValue > 255 {
					grayValue = 255
				}

				// Store pixel value
				if bytesPerPixel == 2 {
					// 16-bit - scale to full range
					value := uint16(grayValue) * 257 // Scale 0-255 to 0-65535
					pixelData[idx] = byte(value & 0xFF)
					pixelData[idx+1] = byte((value >> 8) & 0xFF)
				} else {
					// 8-bit
					pixelData[idx] = byte(grayValue)
				}
			}
		}
	}

	logrus.Info("Finished spiral pattern generation with modality text")
}

// isModalityTextPixel checks if a pixel should be part of the modality text (centered)
func (i *ImageGenerator) isModalityTextPixel(x, y, width, height int) bool {
	// Calculate center point
	centerX := width / 2
	centerY := height / 2

	// Large "CR" text pattern centered in the image
	// Each character is approximately 60x80 pixels

	// C - Left vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX-80 && x < centerX-60 {
			return true
		}
	}
	// C - Top horizontal line
	if y >= centerY-40 && y < centerY-20 {
		if x >= centerX-80 && x < centerX-20 {
			return true
		}
	}
	// C - Bottom horizontal line
	if y >= centerY+20 && y < centerY+40 {
		if x >= centerX-80 && x < centerX-20 {
			return true
		}
	}

	// R - Left vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX-10 && x < centerX+10 {
			return true
		}
	}
	// R - Top horizontal line
	if y >= centerY-40 && y < centerY-20 {
		if x >= centerX-10 && x < centerX+50 {
			return true
		}
	}
	// R - Middle horizontal line
	if y >= centerY-10 && y < centerY+10 {
		if x >= centerX-10 && x < centerX+40 {
			return true
		}
	}
	// R - Right diagonal line
	if y >= centerY+10 && y < centerY+40 {
		if x >= centerX+30 && x < centerX+50 {
			// Simple diagonal approximation
			if (x - centerX - 30) == (y - centerY - 10) {
				return true
			}
		}
	}

	return false
}

// generateCTPattern generates CT-like noise pattern
func (i *ImageGenerator) generateCTPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// CT characteristics: create a clear circular pattern with high contrast
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Calculate distance from center
			centerX, centerY := width/2, height/2
			dist := (x-centerX)*(x-centerX) + (y-centerY)*(y-centerY)
			maxRadius := (width * width) / 4

			var baseValue int
			if dist < maxRadius {
				// Inside circle - bright (bone/tissue)
				baseValue = 3000 + i.rand.Intn(500) // High values for visibility
			} else {
				// Outside circle - dark (air/lung)
				baseValue = 100 + i.rand.Intn(200) // Low values for contrast
			}

			// Ensure values are in valid range for 16-bit
			if baseValue < 0 {
				baseValue = 0
			}
			if baseValue > 65535 {
				baseValue = 65535
			}

			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(baseValue)
				pixelData[idx] = byte(value & 0xFF)
				pixelData[idx+1] = byte((value >> 8) & 0xFF)
			} else {
				pixelData[idx] = byte(baseValue & 0xFF)
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

// addBurnedInText adds burnt-in text directly to the pixel data for testing
func (i *ImageGenerator) addBurnedInText(pixelData []byte, width, height, bytesPerPixel int) {
	logrus.Info("Adding burnt-in text to pixel data")

	// Create a very obvious test pattern - a large white rectangle in the top-left corner
	// This should be clearly visible against the noise pattern
	testWidth := 200
	testHeight := 100

	for y := 0; y < testHeight && y < height; y++ {
		for x := 0; x < testWidth && x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Set pixel to maximum brightness for high contrast
			if bytesPerPixel == 2 {
				// 16-bit maximum value
				pixelData[idx] = 0xFF
				pixelData[idx+1] = 0xFF
			} else {
				// 8-bit maximum value
				pixelData[idx] = 0xFF
			}
		}
	}

	// Add a smaller rectangle below for "TEST"
	testY := testHeight + 20
	for y := testY; y < testY+50 && y < height; y++ {
		for x := 0; x < 150 && x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			if bytesPerPixel == 2 {
				pixelData[idx] = 0xFF
				pixelData[idx+1] = 0xFF
			} else {
				pixelData[idx] = 0xFF
			}
		}
	}

	logrus.Info("Finished adding burnt-in text")
}

// drawTextBlock draws a simple rectangular block for text
func (i *ImageGenerator) drawTextBlock(pixelData []byte, width, height, bytesPerPixel int, startX, startY, blockWidth, blockHeight int) {
	for y := startY; y < startY+blockHeight && y < height; y++ {
		for x := startX; x < startX+blockWidth && x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Set pixel to maximum brightness for high contrast
			if bytesPerPixel == 2 {
				// 16-bit maximum value
				pixelData[idx] = 0xFF
				pixelData[idx+1] = 0xFF
			} else {
				// 8-bit maximum value
				pixelData[idx] = 0xFF
			}
		}
	}
}

// drawCharacter draws a simple character using basic pixel patterns
func (i *ImageGenerator) drawCharacter(pixelData []byte, width, height, bytesPerPixel int, char rune, startX, startY, fontWidth, fontHeight int) {
	// Simple character patterns (8x8 pixels each)
	charPatterns := map[rune][8]uint8{
		'T': {0xFF, 0xFF, 0x18, 0x18, 0x18, 0x18, 0x18, 0x18},
		'E': {0xFF, 0xFF, 0xC0, 0xFC, 0xFC, 0xC0, 0xFF, 0xFF},
		'S': {0x7E, 0xFF, 0xC0, 0x7E, 0x03, 0xFF, 0x7E, 0x00},
		' ': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		'P': {0xFF, 0xFF, 0xC3, 0xFF, 0xFF, 0xC0, 0xC0, 0xC0},
		'a': {0x00, 0x7E, 0x03, 0x7F, 0xC3, 0xC3, 0x7F, 0x00},
		't': {0x30, 0x30, 0xFC, 0x30, 0x30, 0x30, 0x1C, 0x00},
		'i': {0x18, 0x00, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00},
		'e': {0x00, 0x7E, 0xC3, 0xFF, 0xC0, 0xC3, 0x7E, 0x00},
		'n': {0x00, 0xDE, 0xF3, 0xC3, 0xC3, 0xC3, 0xC3, 0x00},
		':': {0x00, 0x18, 0x18, 0x00, 0x00, 0x18, 0x18, 0x00},
		'D': {0xFC, 0xFE, 0xC3, 0xC3, 0xC3, 0xC3, 0xFE, 0xFC},
		'O': {0x7E, 0xFF, 0xC3, 0xC3, 0xC3, 0xC3, 0xFF, 0x7E},
		'^': {0x18, 0x3C, 0x66, 0xC3, 0x00, 0x00, 0x00, 0x00},
		'J': {0x0F, 0x0F, 0x0F, 0x0F, 0x0F, 0xCF, 0xFE, 0x7C},
		'H': {0xC3, 0xC3, 0xC3, 0xFF, 0xFF, 0xC3, 0xC3, 0xC3},
		'N': {0xC3, 0xE3, 0xF3, 0xDB, 0xCF, 0xC7, 0xC3, 0xC3},
		'M': {0xC3, 0xE7, 0xFF, 0xDB, 0xC3, 0xC3, 0xC3, 0xC3},
		'I': {0xFF, 0x18, 0x18, 0x18, 0x18, 0x18, 0xFF, 0x00},
		'0': {0x7E, 0xFF, 0xC3, 0xC3, 0xC3, 0xC3, 0xFF, 0x7E},
		'1': {0x18, 0x38, 0x18, 0x18, 0x18, 0x18, 0x3C, 0x00},
		'2': {0x7E, 0xFF, 0x03, 0x7E, 0xC0, 0xC0, 0xFF, 0xFF},
		'3': {0x7E, 0xFF, 0x03, 0x3E, 0x03, 0x03, 0xFF, 0x7E},
		'4': {0x06, 0x0E, 0x1E, 0x36, 0x66, 0xFF, 0x06, 0x06},
		'5': {0xFF, 0xFF, 0xC0, 0xFE, 0x03, 0x03, 0xFF, 0x7E},
		'6': {0x7E, 0xFF, 0xC0, 0xFE, 0xC3, 0xC3, 0xFF, 0x7E},
		'7': {0xFF, 0xFF, 0x03, 0x06, 0x0C, 0x18, 0x30, 0x60},
		'8': {0x7E, 0xFF, 0xC3, 0x7E, 0xC3, 0xC3, 0xFF, 0x7E},
		'9': {0x7E, 0xFF, 0xC3, 0xFF, 0x03, 0x03, 0xFF, 0x7E},
		'-': {0x00, 0x00, 0x00, 0xFF, 0xFF, 0x00, 0x00, 0x00},
		'C': {0x7E, 0xFF, 0xC0, 0xC0, 0xC0, 0xC0, 0xFF, 0x7E},
		'R': {0xFF, 0xFF, 0xC3, 0xFF, 0xFC, 0xC6, 0xC3, 0xC3},
	}

	// Get character pattern or use default for unknown characters
	pattern, exists := charPatterns[char]
	if !exists {
		// Use a simple pattern for unknown characters
		pattern = [8]uint8{0x7E, 0xFF, 0xC3, 0xC3, 0xC3, 0xC3, 0xFF, 0x7E}
	}

	// Draw the character pattern
	for row := 0; row < fontHeight; row++ {
		if startY+row >= height {
			break
		}

		rowData := pattern[row]
		for col := 0; col < fontWidth; col++ {
			if startX+col >= width {
				break
			}

			// Check if pixel should be set (bit is 1)
			if (rowData>>(7-col))&1 == 1 {
				idx := ((startY+row)*width + (startX + col)) * bytesPerPixel

				// Set pixel to high contrast value for visibility
				// Use a very bright value that will stand out against the noise
				if bytesPerPixel == 2 {
					// 16-bit bright value (near maximum)
					pixelData[idx] = 0xFF
					pixelData[idx+1] = 0xFF
				} else {
					// 8-bit bright value
					pixelData[idx] = 0xFF
				}
			}
		}
	}
}

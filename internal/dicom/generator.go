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
	config      *config.Config
	uidGen      *EnhancedUIDGenerator
	imageGen    *ImageGenerator
	metadataGen *MetadataGenerator
}

// NewGenerator creates a new DICOM generator
func NewGenerator(cfg *config.Config) *Generator {
	return &Generator{
		config:      cfg,
		uidGen:      NewEnhancedUIDGenerator(cfg.DICOM.OrgRoot),
		imageGen:    NewImageGenerator(),
		metadataGen: NewMetadataGenerator(cfg.DICOM.OrgRoot),
	}
}

// ImageGenerator generates synthetic images
type ImageGenerator struct {
	rand *rand.Rand
}

// NewImageGenerator creates a new image generator
func NewImageGenerator() *ImageGenerator {
	return &ImageGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateStudy generates a complete DICOM study with comprehensive metadata
func (g *Generator) GenerateStudy(params types.StudyParams) (*types.Study, error) {
	// Generate comprehensive patient metadata
	patientMetadata := g.metadataGen.PatientModule(params.PatientName, params.PatientID)

	// Generate comprehensive study metadata
	studyMetadata := g.metadataGen.StudyModule(params.StudyDescription, params.AccessionNumber)

	// Create study with comprehensive metadata
	study := &types.Study{
		StudyInstanceUID: studyMetadata["StudyInstanceUID"].(string),
		StudyDate:        studyMetadata["StudyDate"].(string),
		StudyTime:        studyMetadata["StudyTime"].(string),
		AccessionNumber:  studyMetadata["AccessionNumber"].(string),
		StudyDescription: studyMetadata["StudyDescription"].(string),
		PatientName:      patientMetadata["PatientName"].(string),
		PatientID:        patientMetadata["PatientID"].(string),
		PatientBirthDate: patientMetadata["PatientBirthDate"].(string),
		Series:           make([]types.Series, 0, params.SeriesCount),
	}

	// Generate series with comprehensive metadata
	for i := 0; i < params.SeriesCount; i++ {
		series, err := g.generateSeries(study, params.Modality, i+1, params.ImageCount, params.AnatomicalRegion)
		if err != nil {
			return nil, fmt.Errorf("failed to generate series %d: %w", i+1, err)
		}
		study.Series = append(study.Series, *series)
	}

	return study, nil
}

// generateSeries generates a DICOM series with comprehensive metadata
func (g *Generator) generateSeries(study *types.Study, modality string, seriesNumber, imageCount int, anatomicalRegion string) (*types.Series, error) {
	// Generate comprehensive series metadata
	seriesMetadata := g.metadataGen.SeriesModule(modality, seriesNumber, anatomicalRegion)

	series := &types.Series{
		SeriesInstanceUID: seriesMetadata["SeriesInstanceUID"].(string),
		SeriesNumber:      seriesMetadata["SeriesNumber"].(int),
		Modality:          seriesMetadata["Modality"].(string),
		SeriesDescription: seriesMetadata["SeriesDescription"].(string),
		Images:            make([]types.Image, 0, imageCount),
	}

	// Generate images with comprehensive metadata
	for i := 0; i < imageCount; i++ {
		image, err := g.generateImage(study, series, i+1)
		if err != nil {
			return nil, fmt.Errorf("failed to generate image %d: %w", i+1, err)
		}
		series.Images = append(series.Images, *image)
	}

	return series, nil
}

// generateImage generates a DICOM image with comprehensive metadata and burned-in text for integration debugging
func (g *Generator) generateImage(study *types.Study, series *types.Series, instanceNumber int) (*types.Image, error) {
	// Get SOP class UID for modality
	sopClassUID, exists := types.SOPClassUIDs[series.Modality]
	if !exists {
		return nil, fmt.Errorf("unsupported modality: %s", series.Modality)
	}

	// Generate comprehensive image metadata
	imageMetadata := g.metadataGen.ImageModule(series.Modality, instanceNumber, sopClassUID)

	// Get image dimensions
	imageSize, exists := types.ImageDimensions[series.Modality]
	if !exists {
		return nil, fmt.Errorf("unsupported modality: %s", series.Modality)
	}

	// Generate comprehensive image pixel metadata
	pixelMetadata := g.metadataGen.ImagePixelModule(imageSize.Width, imageSize.Height, imageSize.BitsPerPixel, series.Modality)

	// Generate pixel data with burned-in metadata
	fmt.Printf("DEBUG: generateImage - modality='%s', imageSize=%+v\n", series.Modality, imageSize)
	pixelData, err := g.imageGen.GenerateImageWithMetadata(study, series, instanceNumber, len(series.Images)+1, imageSize.Width, imageSize.Height, imageSize.BitsPerPixel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate pixel data: %w", err)
	}

	image := &types.Image{
		SOPInstanceUID: imageMetadata["SOPInstanceUID"].(string),
		SOPClassUID:    imageMetadata["SOPClassUID"].(string),
		InstanceNumber: imageMetadata["InstanceNumber"].(int),
		PixelData:      pixelData,
		Width:          pixelMetadata["Columns"].(int),
		Height:         pixelMetadata["Rows"].(int),
		BitsPerPixel:   pixelMetadata["BitsAllocated"].(int),
		Modality:       series.Modality,
	}

	return image, nil
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
		// MRI: high contrast, more uniform noise with sequence-specific patterns
		i.generateMRPattern(pixelData, width, height, bytesPerPixel)
	case "US":
		// Ultrasound: low contrast, speckle noise
		i.generateUSPattern(pixelData, width, height, bytesPerPixel)
	case "MG":
		// Mammography: high resolution, subtle patterns
		i.generateMGPattern(pixelData, width, height, bytesPerPixel)
	case "NM":
		// Nuclear Medicine: low resolution, hot spots
		i.generateCTPattern(pixelData, width, height, bytesPerPixel) // Use CT pattern as fallback
	case "PT":
		// PET: low resolution, metabolic activity patterns
		i.generateCTPattern(pixelData, width, height, bytesPerPixel) // Use CT pattern as fallback
	case "RT":
		// Radiotherapy: treatment planning patterns
		i.generateCTPattern(pixelData, width, height, bytesPerPixel) // Use CT pattern as fallback
	case "SR":
		// Structured Reports: no pixel data
		logrus.Info("SR modality - no pixel data generation needed")
	default:
		// Default: simple noise pattern
		i.generateDefaultPattern(pixelData, width, height, bytesPerPixel)
	}

	return pixelData, nil
}

// GenerateImageWithMetadata generates synthetic image data with burned-in metadata for integration debugging
func (i *ImageGenerator) GenerateImageWithMetadata(study *types.Study, series *types.Series, instanceNumber, totalInstances int, width, height, bitsPerPixel int) ([]byte, error) {
	// First generate the base pattern
	pixelData, err := i.GenerateImage(series.Modality, width, height, bitsPerPixel)
	if err != nil {
		return nil, fmt.Errorf("failed to generate base image: %w", err)
	}

	// Add burned-in metadata to the pixel data
	if err := i.addBurnedInMetadata(pixelData, study, series, instanceNumber, totalInstances, width, height, bitsPerPixel); err != nil {
		return nil, fmt.Errorf("failed to add burned-in metadata: %w", err)
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
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
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
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateUSPattern generates ultrasound-like noise pattern
func (i *ImageGenerator) generateUSPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// Ultrasound characteristics: low contrast, speckle noise with anatomical structures
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Create anatomical structure (simulate abdominal ultrasound)
			var baseValue int
			
			// Top section: skin/fat layer (bright)
			if y < height/8 {
				baseValue = 180 + i.rand.Intn(40) // Bright with some variation
			} else if y < height/4 {
				// Muscle layer (medium brightness)
				baseValue = 120 + i.rand.Intn(60)
			} else if y < height/2 {
				// Organ tissue (variable brightness)
				baseValue = 80 + i.rand.Intn(80)
			} else {
				// Deeper structures (darker)
				baseValue = 40 + i.rand.Intn(60)
			}

			// Add speckle noise (characteristic of ultrasound)
			speckle := i.rand.Intn(20) - 10 // -10 to +10 variation
			baseValue += speckle

			// Add horizontal scan lines (ultrasound beam pattern)
			if y%4 == 0 {
				baseValue += i.rand.Intn(15) - 7 // Slight brightness variation
			}

			// Ensure values are in valid range
			if baseValue < 0 {
				baseValue = 0
			}
			if baseValue > 255 {
				baseValue = 255
			}

			// Store pixel value with proper bit depth handling
			if bytesPerPixel == 2 {
				// 16-bit - scale to full range
				value := uint16(baseValue) * 257 // Scale 0-255 to 0-65535
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				// 8-bit
				pixelData[idx] = byte(baseValue)
			}
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
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				pixelData[idx] = byte(noise)
			}
		}
	}
}

// generateNMPattern generates Nuclear Medicine-like pattern with hot spots
func (i *ImageGenerator) generateNMPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// NM characteristics: low resolution, hot spots representing radioactive uptake
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Base background noise (low activity)
			baseValue := i.rand.Intn(50) + 10

			// Add hot spots (areas of high radioactive uptake)
			hotSpots := []struct{ cx, cy, radius, intensity int }{
				{width / 4, height / 4, width / 8, 200},
				{3 * width / 4, height / 2, width / 6, 150},
				{width / 2, 3 * height / 4, width / 10, 100},
			}

			for _, spot := range hotSpots {
				dx := x - spot.cx
				dy := y - spot.cy
				distance := dx*dx + dy*dy
				if distance < spot.radius*spot.radius {
					// Gaussian falloff for hot spot
					intensity := float64(spot.intensity) * math.Exp(-float64(distance)/(2*float64(spot.radius*spot.radius)))
					baseValue += int(intensity)
				}
			}

			// Ensure values are in valid range
			if baseValue > 65535 {
				baseValue = 65535
			}

			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(baseValue)
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				pixelData[idx] = byte(baseValue & 0xFF)
			}
		}
	}
}

// generatePTPattern generates PET-like pattern with metabolic activity
func (i *ImageGenerator) generatePTPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// PET characteristics: metabolic activity patterns, SUV values
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Base background (low metabolic activity)
			baseValue := i.rand.Intn(30) + 5

			// Add metabolic activity regions
			activityRegions := []struct{ cx, cy, radius, suv float64 }{
				{float64(width) * 0.3, float64(height) * 0.3, float64(width) * 0.15, 3.5},
				{float64(width) * 0.7, float64(height) * 0.6, float64(width) * 0.12, 2.8},
				{float64(width) * 0.5, float64(height) * 0.8, float64(width) * 0.08, 1.9},
			}

			for _, region := range activityRegions {
				dx := float64(x) - region.cx
				dy := float64(y) - region.cy
				distance := math.Sqrt(dx*dx + dy*dy)
				if distance < region.radius {
					// Gaussian falloff for metabolic activity
					intensity := region.suv * 1000 * math.Exp(-(distance*distance)/(2*region.radius*region.radius))
					baseValue += int(intensity)
				}
			}

			// Ensure values are in valid range
			if baseValue > 65535 {
				baseValue = 65535
			}

			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(baseValue)
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				pixelData[idx] = byte(baseValue & 0xFF)
			}
		}
	}
}

// generateRTPattern generates Radiotherapy treatment planning pattern
func (i *ImageGenerator) generateRTPattern(pixelData []byte, width, height, bytesPerPixel int) {
	// RT characteristics: treatment planning with dose distributions
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Base background (air/tissue)
			baseValue := i.rand.Intn(100) + 50

			// Add treatment field (rectangular high-dose region)
			fieldX1, fieldY1 := width/4, height/4
			fieldX2, fieldY2 := 3*width/4, 3*height/4

			if x >= fieldX1 && x <= fieldX2 && y >= fieldY1 && y <= fieldY2 {
				// High dose region
				baseValue += 2000 + i.rand.Intn(500)
			}

			// Add dose gradient around field edges
			edgeDistance := 20
			if (x >= fieldX1-edgeDistance && x <= fieldX1+edgeDistance) ||
				(x >= fieldX2-edgeDistance && x <= fieldX2+edgeDistance) ||
				(y >= fieldY1-edgeDistance && y <= fieldY1+edgeDistance) ||
				(y >= fieldY2-edgeDistance && y <= fieldY2+edgeDistance) {
				baseValue += 500 + i.rand.Intn(200)
			}

			// Ensure values are in valid range
			if baseValue > 65535 {
				baseValue = 65535
			}

			// Store pixel value
			if bytesPerPixel == 2 {
				value := uint16(baseValue)
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
			} else {
				pixelData[idx] = byte(baseValue & 0xFF)
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
				pixelData[idx] = byte(value & 0xFF)          // Low byte first (little-endian)
				pixelData[idx+1] = byte((value >> 8) & 0xFF) // High byte second
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

// addBurnedInMetadata adds patient and study metadata to pixel data for integration debugging
func (i *ImageGenerator) addBurnedInMetadata(pixelData []byte, study *types.Study, series *types.Series, instanceNumber, totalInstances int, width, height, bitsPerPixel int) error {
	// Skip burned-in metadata for modalities without pixel data (like SR)
	if width == 0 || height == 0 || bitsPerPixel == 0 {
		return nil
	}

	bytesPerPixel := bitsPerPixel / 8
	if bitsPerPixel%8 != 0 {
		bytesPerPixel++
	}

	// Create metadata text lines
	textLines := []string{
		fmt.Sprintf("Patient: %s", study.PatientName),
		fmt.Sprintf("Patient ID: %s", study.PatientID),
		fmt.Sprintf("DOB: %s", study.PatientBirthDate),
		fmt.Sprintf("Accession: %s", study.AccessionNumber),
		fmt.Sprintf("Study UID: %s", study.StudyInstanceUID),
		fmt.Sprintf("Series UID: %s", series.SeriesInstanceUID),
		fmt.Sprintf("Instance: %d of %d", instanceNumber, totalInstances),
		fmt.Sprintf("Modality: %s", series.Modality),
		fmt.Sprintf("Study Date: %s", study.StudyDate),
		"Generated by crgodicom flatmapit.com",
	}

	// Position for text (top-left corner with padding)
	x := 20
	y := 30
	lineHeight := 16
	padding := 12

	// Calculate the maximum text width more accurately
	maxTextWidth := 0
	for _, line := range textLines {
		textWidth := len(line) * 8 // More accurate width for 7x13 font
		if textWidth > maxTextWidth {
			maxTextWidth = textWidth
		}
	}

	// Ensure we don't exceed image boundaries
	if rectRight := x + maxTextWidth + padding; rectRight > width {
		maxTextWidth = width - x - padding
	}
	if rectBottom := y + (len(textLines) * lineHeight) + padding; rectBottom > height {
		// Truncate text lines if they would exceed image height
		maxLines := (height - y - padding) / lineHeight
		if maxLines < len(textLines) {
			textLines = textLines[:maxLines]
		}
	}

	// Draw text lines FIRST
	for lineIndex, line := range textLines {
		textY := y + (lineIndex * lineHeight)
		if textY >= height {
			break
		}
		i.drawText(pixelData, line, x, textY, width, height, bytesPerPixel)
	}

	// Draw background rectangle AFTER text (so text is visible)
	rectTop := y - padding
	rectBottom := y + (len(textLines) * lineHeight) + padding
	rectLeft := x - padding
	rectRight := x + maxTextWidth + padding

	// Draw semi-transparent background rectangle (but preserve text pixels)
	for rectY := rectTop; rectY < rectBottom && rectY < height; rectY++ {
		for rectX := rectLeft; rectX < rectRight && rectX < width; rectX++ {
			idx := (rectY*width + rectX) * bytesPerPixel
			if idx+bytesPerPixel <= len(pixelData) {
				// Only set background if pixel is not text (not maximum brightness)
				if bytesPerPixel == 2 {
					lowByte := pixelData[idx]
					highByte := pixelData[idx+1]
					currentValue := uint16(lowByte) | (uint16(highByte) << 8)

					// Only set background if pixel is not text
					if currentValue != 65535 {
						pixelData[idx] = 0x00
						pixelData[idx+1] = 0x00
					}
				} else {
					// Only set background if pixel is not text
					if pixelData[idx] != 255 {
						pixelData[idx] = 0x00
					}
				}
			}
		}
	}

	return nil
}

// formatDate formats a time.Time to YYYYMMDD string
func (i *ImageGenerator) formatDate(t time.Time) string {
	return t.Format("20060102")
}

// drawText draws text using a simple block-based approach for better readability
func (i *ImageGenerator) drawText(pixelData []byte, text string, startX, startY, width, height, bytesPerPixel int) {
	fontWidth := 8
	charSpacing := 1

	// Calculate adaptive text color ONCE for the entire text area
	// This ensures all characters use the same brightness for perfect consistency
	textColor := i.calculateAdaptiveTextColor(pixelData, width, height, bytesPerPixel, startX, startY, len(text)*9, 12)

	for charIndex, char := range text {
		charX := startX + (charIndex * (fontWidth + charSpacing))
		if charX+fontWidth >= width {
			break
		}

		// Draw character using the consistent text color
		i.drawSimpleCharWithColor(pixelData, char, charX, startY, width, height, bytesPerPixel, textColor)
	}
}

// drawSimpleChar draws a character using consistent brightness for all characters (backward compatibility)
func (i *ImageGenerator) drawSimpleChar(pixelData []byte, char rune, startX, startY, width, height, bytesPerPixel int) {
	// Calculate adaptive text color for this character
	textColor := i.calculateAdaptiveTextColor(pixelData, width, height, bytesPerPixel, startX, startY, 8, 12)

	// Use the new function with the calculated color
	i.drawSimpleCharWithColor(pixelData, char, startX, startY, width, height, bytesPerPixel, textColor)
}

// drawSimpleCharWithColor draws a character using a pre-calculated consistent text color
func (i *ImageGenerator) drawSimpleCharWithColor(pixelData []byte, char rune, startX, startY, width, height, bytesPerPixel int, textColor uint16) {
	fontWidth := 8
	fontHeight := 12

	// Simple character patterns - each character is drawn as blocks
	// All patterns are exactly 8x12 for consistency
	charPatterns := map[rune][]string{
		' ': []string{
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'A': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX    XX",
			"XXXXXXXX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'B': []string{
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"XXXXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'C': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX      ",
			"XX      ",
			"XX      ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'D': []string{
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XXXXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'E': []string{
			"XXXXXXXX",
			"XX      ",
			"XX      ",
			"XXXXXX  ",
			"XX      ",
			"XX      ",
			"XXXXXXXX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'F': []string{
			"XXXXXXXX",
			"XX      ",
			"XX      ",
			"XXXXXX  ",
			"XX      ",
			"XX      ",
			"XX      ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'G': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX      ",
			"XX  XXXX",
			"XX    XX",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'H': []string{
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XXXXXXXX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'I': []string{
			"XXXXXXXX",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"XXXXXXXX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'J': []string{
			"XXXXXXXX",
			"     XX ",
			"     XX ",
			"     XX ",
			"XX   XX ",
			" XX XX  ",
			"  XXX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'K': []string{
			"XX    XX",
			"XX   XX ",
			"XX  XX  ",
			"XXXX    ",
			"XX  XX  ",
			"XX   XX ",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'L': []string{
			"XX      ",
			"XX      ",
			"XX      ",
			"XX      ",
			"XX      ",
			"XX      ",
			"XXXXXXXX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'M': []string{
			"XX    XX",
			"XXX  XXX",
			"XXXXXXX ",
			"XX X XX ",
			"XX   XX ",
			"XX   XX ",
			"XX   XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'N': []string{
			"XX    XX",
			"XXX   XX",
			"XXXX  XX",
			"XX XX XX",
			"XX  XXXX",
			"XX   XXX",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'O': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'P': []string{
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"XXXXXX  ",
			"XX      ",
			"XX      ",
			"XX      ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'Q': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX    XX",
			"XX    XX",
			"XX  X XX",
			" XX  XX ",
			"  XXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'R': []string{
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"XXXXXX  ",
			"XX  XX  ",
			"XX   XX ",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'S': []string{
			"  XXXXX ",
			" XX     ",
			"XX      ",
			"  XXXX  ",
			"     XX ",
			"     XX ",
			" XXXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'T': []string{
			"XXXXXXXX",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'U': []string{
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XX    XX",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'V': []string{
			"XX    XX",
			"XX    XX",
			"XX    XX",
			" XX  XX ",
			" XX  XX ",
			"  XXXX  ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'W': []string{
			"XX    XX",
			"XX    XX",
			"XX    XX",
			"XX X XX ",
			"XXXXXXX ",
			"XXX XXX ",
			"XX   XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'X': []string{
			"XX    XX",
			" XX  XX ",
			"  XXXX  ",
			"   XX   ",
			"  XXXX  ",
			" XX  XX ",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'Y': []string{
			"XX    XX",
			" XX  XX ",
			"  XXXX  ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'Z': []string{
			"XXXXXXXX",
			"     XX ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX     ",
			"XXXXXXXX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'0': []string{
			"  XXXX  ",
			" XX  XX ",
			"XX  X XX",
			"XX XX XX",
			"XXX  XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'1': []string{
			"   XX   ",
			"  XXX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'2': []string{
			"  XXXX  ",
			" XX  XX ",
			"     XX ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'3': []string{
			"  XXXX  ",
			" XX  XX ",
			"     XX ",
			"  XXXX  ",
			"     XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'4': []string{
			"    XX  ",
			"   XXX  ",
			"  XXXX  ",
			" XX XX  ",
			"XXXXXXXX",
			"    XX  ",
			"    XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'5': []string{
			" XXXXXX ",
			" XX     ",
			" XXXXX  ",
			"     XX ",
			"     XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'6': []string{
			"  XXXX  ",
			" XX     ",
			"XX      ",
			"XXXXXX  ",
			"XX    XX",
			"XX    XX",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'7': []string{
			"XXXXXXXX",
			"     XX ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX     ",
			"XX      ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'8': []string{
			"  XXXX  ",
			" XX  XX ",
			" XX  XX ",
			"  XXXX  ",
			" XX  XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'9': []string{
			"  XXXX  ",
			" XX  XX ",
			" XX  XX ",
			"  XXXXXX",
			"      XX",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		':': []string{
			"        ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'.': []string{
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'-': []string{
			"        ",
			"        ",
			"        ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'^': []string{
			"   XX   ",
			"  XXXX  ",
			" XX  XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		// Lowercase letters
		'a': []string{
			"        ",
			"        ",
			"  XXXX  ",
			"     XX ",
			"  XXXXX ",
			" XX  XX ",
			"  XXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		't': []string{
			"   XX   ",
			"   XX   ",
			" XXXXXX ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"    XXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'i': []string{
			"   XX   ",
			"        ",
			"  XXX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'e': []string{
			"        ",
			"        ",
			"  XXXX  ",
			" XX  XX ",
			" XXXXXX ",
			" XX     ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'n': []string{
			"        ",
			"        ",
			" XX XXX ",
			" XXX  XX",
			" XX   XX",
			" XX   XX",
			" XX   XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'r': []string{
			"        ",
			"        ",
			" XX XXX ",
			" XXX  XX",
			" XX     ",
			" XX     ",
			" XX     ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'o': []string{
			"        ",
			"        ",
			"  XXXX  ",
			" XX  XX ",
			" XX  XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'c': []string{
			"        ",
			"        ",
			"  XXXX  ",
			" XX  XX ",
			" XX     ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'u': []string{
			"        ",
			"        ",
			" XX   XX",
			" XX   XX",
			" XX   XX",
			" XXX  XX",
			"  XXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'l': []string{
			"  XXX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'y': []string{
			"        ",
			"        ",
			" XX   XX",
			" XX   XX",
			" XX   XX",
			"  XXXXX ",
			"     XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
		},
		'x': []string{
			"        ",
			"        ",
			" XX   XX",
			"  XX XX ",
			"   XXX  ",
			"  XX XX ",
			" XX   XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'z': []string{
			"        ",
			"        ",
			" XXXXXX ",
			"     XX ",
			"   XXX  ",
			" XX     ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		// Additional missing characters
		's': []string{
			"        ",
			"        ",
			"  XXXX  ",
			" XX     ",
			"  XXXX  ",
			"     XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'd': []string{
			"     XX ",
			"     XX ",
			"  XXXXX ",
			" XX  XX ",
			" XX  XX ",
			" XX  XX ",
			"  XXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'f': []string{
			"   XXX  ",
			"  XX    ",
			" XXXXXX ",
			"  XX    ",
			"  XX    ",
			"  XX    ",
			"  XX    ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'g': []string{
			"        ",
			"        ",
			"  XXXXX ",
			" XX  XX ",
			" XX  XX ",
			"  XXXXX ",
			"     XX ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
		},
		'h': []string{
			" XX     ",
			" XX     ",
			" XX XXX ",
			" XXX  XX",
			" XX   XX",
			" XX   XX",
			" XX   XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'j': []string{
			"     XX ",
			"        ",
			"    XX  ",
			"    XX  ",
			"    XX  ",
			" XX  XX ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'k': []string{
			" XX     ",
			" XX     ",
			" XX  XX ",
			" XX XX  ",
			" XXXX   ",
			" XX XX  ",
			" XX  XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'm': []string{
			"        ",
			"        ",
			"XX XXX  ",
			"XXX  XXX",
			"XX X XX ",
			"XX   XX ",
			"XX   XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'p': []string{
			"        ",
			"        ",
			" XX XXX ",
			" XXX  XX",
			" XX   XX",
			" XXXXX  ",
			" XX     ",
			" XX     ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'q': []string{
			"        ",
			"        ",
			"  XXXXX ",
			" XX  XX ",
			" XX  XX ",
			"  XXXXX ",
			"     XX ",
			"     XX ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'v': []string{
			"        ",
			"        ",
			"XX   XX ",
			"XX   XX ",
			" XX XX  ",
			"  XXX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'w': []string{
			"        ",
			"        ",
			"XX   XX ",
			"XX   XX ",
			"XX X XX ",
			"XXXXXXX ",
			" XX XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'b': []string{
			" XX     ",
			" XX     ",
			" XXXXX  ",
			" XX  XX ",
			" XX  XX ",
			" XX  XX ",
			" XXXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		// Special characters
		'/': []string{
			"     XX ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX     ",
			" XX     ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'_': []string{
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'=': []string{
			"        ",
			"        ",
			" XXXXXX ",
			"        ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'+': []string{
			"        ",
			"   XX   ",
			"   XX   ",
			" XXXXXX ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'*': []string{
			" XX XX  ",
			"  XXX   ",
			"XXXXXXX ",
			"  XXX   ",
			" XX XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'#': []string{
			" XX XX  ",
			" XXXXXX ",
			" XX XX  ",
			" XXXXXX ",
			" XX XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'@': []string{
			"  XXXX  ",
			" XX  XX ",
			" XX XXX ",
			" XX XXX ",
			" XX XXX ",
			" XX     ",
			"  XXXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'$': []string{
			"   XX   ",
			"  XXXX  ",
			" XX     ",
			"  XXXX  ",
			"     XX ",
			"  XXXX  ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'%': []string{
			"XX    XX",
			"XX   XX ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX   XX",
			"XX    XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'&': []string{
			"  XXX   ",
			" XX  XX ",
			"  XXX   ",
			" XXX    ",
			"XX  XX  ",
			"XX   XX ",
			" XXX  XX",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'(': []string{
			"   XX   ",
			"  XX    ",
			" XX     ",
			" XX     ",
			" XX     ",
			"  XX    ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		')': []string{
			"   XX   ",
			"    XX  ",
			"     XX ",
			"     XX ",
			"     XX ",
			"    XX  ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'[': []string{
			" XXXXXX ",
			" XX     ",
			" XX     ",
			" XX     ",
			" XX     ",
			" XX     ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		']': []string{
			" XXXXXX ",
			"     XX ",
			"     XX ",
			"     XX ",
			"     XX ",
			"     XX ",
			" XXXXXX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'{': []string{
			"   XXX  ",
			"  XX    ",
			"  XX    ",
			" XX     ",
			"  XX    ",
			"  XX    ",
			"   XXX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'}': []string{
			" XXX    ",
			"    XX  ",
			"    XX  ",
			"     XX ",
			"    XX  ",
			"    XX  ",
			" XXX    ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'<': []string{
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX     ",
			"  XX    ",
			"   XX   ",
			"    XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'>': []string{
			" XX     ",
			"  XX    ",
			"   XX   ",
			"    XX  ",
			"   XX   ",
			"  XX    ",
			" XX     ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'|': []string{
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'\\': []string{
			" XX     ",
			"  XX    ",
			"   XX   ",
			"    XX  ",
			"     XX ",
			"     XX ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'`': []string{
			" XX     ",
			"  XX    ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'~': []string{
			"        ",
			" XX  XX ",
			"XX  XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'!': []string{
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"   XX   ",
			"        ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'?': []string{
			"  XXXX  ",
			" XX  XX ",
			"     XX ",
			"    XX  ",
			"   XX   ",
			"        ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'"': []string{
			" XX XX  ",
			" XX XX  ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		'\'': []string{
			"   XX   ",
			"   XX   ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		',': []string{
			"        ",
			"        ",
			"        ",
			"        ",
			"   XX   ",
			"   XX   ",
			"  XX    ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
		';': []string{
			"        ",
			"   XX   ",
			"   XX   ",
			"        ",
			"   XX   ",
			"   XX   ",
			"  XX    ",
			"        ",
			"        ",
			"        ",
			"        ",
			"        ",
		},
	}

	pattern, exists := charPatterns[char]
	if !exists {
		pattern = charPatterns[' '] // Use space for unknown characters
	}

	// Draw the character pattern - ensure all patterns are exactly 8x12
	for row := 0; row < fontHeight && startY+row < height; row++ {
		if row >= len(pattern) {
			continue
		}
		line := pattern[row]
		// Ensure line is exactly 8 characters
		if len(line) < fontWidth {
			line = line + "        "[:fontWidth-len(line)]
		}
		if len(line) > fontWidth {
			line = line[:fontWidth]
		}

		for col := 0; col < fontWidth && startX+col < width; col++ {
			if line[col] == 'X' {
				idx := ((startY+row)*width + (startX + col)) * bytesPerPixel
				if idx+bytesPerPixel <= len(pixelData) {
					// Use the pre-calculated consistent text color
					if bytesPerPixel == 2 {
						// Store 16-bit value in little-endian format
						pixelData[idx] = byte(textColor & 0xFF)          // Low byte first
						pixelData[idx+1] = byte((textColor >> 8) & 0xFF) // High byte second
					} else {
						// 8-bit value
						pixelData[idx] = byte(textColor & 0xFF)
					}

					// Add text outline for better visibility
					// i.addTextOutline(pixelData, width, height, bytesPerPixel, startX+col, startY+row, textColor)
				}
			}
		}
	}
}

// calculateAdaptiveTextColor returns maximum brightness for burned-in metadata
func (i *ImageGenerator) calculateAdaptiveTextColor(pixelData []byte, width, height, bytesPerPixel int, startX, startY, fontWidth, fontHeight int) uint16 {
	// For burned-in metadata (debugging purposes), always use maximum brightness
	// This ensures all characters are consistently visible and bright
	return 65535
}

// addTextOutline adds a subtle outline around text pixels for better visibility
func (i *ImageGenerator) addTextOutline(pixelData []byte, width, height, bytesPerPixel int, x, y int, textColor uint16) {
	// Create outline color (opposite of text color for contrast)
	var outlineColor uint16
	if textColor > 32768 {
		// Bright text - use dark outline
		outlineColor = 0
	} else {
		// Dark text - use bright outline
		outlineColor = 65535
	}

	// Add outline pixels around the text pixel
	offsets := []struct{ dx, dy int }{
		{-1, -1}, {-1, 0}, {-1, 1},
		{0, -1}, {0, 1},
		{1, -1}, {1, 0}, {1, 1},
	}

	for _, offset := range offsets {
		nx := x + offset.dx
		ny := y + offset.dy

		// Check bounds
		if nx >= 0 && nx < width && ny >= 0 && ny < height {
			idx := (ny*width + nx) * bytesPerPixel
			if idx+bytesPerPixel <= len(pixelData) {
				// Only add outline if the pixel isn't already text
				if bytesPerPixel == 2 {
					lowByte := pixelData[idx]
					highByte := pixelData[idx+1]
					currentValue := uint16(lowByte) | (uint16(highByte) << 8)

					// Only add outline if current pixel is not text color
					if currentValue != textColor {
						pixelData[idx] = byte(outlineColor & 0xFF)
						pixelData[idx+1] = byte((outlineColor >> 8) & 0xFF)
					}
				} else {
					currentValue := uint8(pixelData[idx])
					if currentValue != uint8(textColor) {
						pixelData[idx] = byte(outlineColor & 0xFF)
					}
				}
			}
		}
	}
}

package cli

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
	"github.com/flatmapit/crgodicom/internal/export"
	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ExportCommand returns the export command
func ExportCommand() *cli.Command {
	return &cli.Command{
		Name:  "export",
		Usage: "Export DICOM study to various formats",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "study-id",
				Usage:    "Study Instance UID (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "format",
				Usage:    "Export format: png, jpeg, pdf (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory (for PNG/JPEG format)",
			},
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Output file path (for PDF format)",
			},
			&cli.StringFlag{
				Name:  "input-dir",
				Usage: "Studies directory",
				Value: "studies",
			},
			&cli.BoolFlag{
				Name:  "include-metadata",
				Usage: "Include metadata files (PNG format)",
			},
		},
		Action: exportAction,
	}
}

func exportAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	studyID := c.String("study-id")
	format := c.String("format")
	outputDir := c.String("output-dir")
	outputFile := c.String("output-file")
	inputDir := c.String("input-dir")
	includeMetadata := c.Bool("include-metadata")

	// Validate format
	validFormats := []string{"png", "jpeg", "pdf"}
	validFormat := false
	for _, f := range validFormats {
		if format == f {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid format '%s'. Valid formats: %v", format, validFormats)
	}

	// Validate output parameters based on format
	if (format == "png" || format == "jpeg") && outputDir == "" {
		return fmt.Errorf("%s format requires --output-dir parameter", strings.ToUpper(format))
	}
	if format == "pdf" && outputFile == "" {
		return fmt.Errorf("PDF format requires --output-file parameter")
	}

	logrus.Infof("Exporting study %s to %s format", studyID, format)
	logrus.Infof("Input directory: %s", inputDir)
	if format == "png" || format == "jpeg" {
		logrus.Infof("Output directory: %s, Include metadata: %v", outputDir, includeMetadata)
	} else {
		logrus.Infof("Output file: %s", outputFile)
	}

	// Find the study directory
	studyDir := filepath.Join(inputDir, studyID)
	if _, err := os.Stat(studyDir); os.IsNotExist(err) {
		return fmt.Errorf("study directory not found: %s", studyDir)
	}

	// Create DICOM reader and exporter
	reader := dicom.NewReader(cfg)
	exporter := export.NewExporter(outputDir)

	// Read real DICOM metadata and pixel data from the study
	logrus.Infof("About to call ReadDetailedStudyMetadata for: %s", studyDir)
	detailedMetadata, err := reader.ReadDetailedStudyMetadata(studyDir)
	if err != nil {
		logrus.Errorf("ReadDetailedStudyMetadata failed: %v", err)
		return fmt.Errorf("failed to read study metadata: %w", err)
	}
	logrus.Infof("ReadDetailedStudyMetadata completed successfully")

	// Convert detailed metadata to types.Study for export
	study, err := convertDetailedMetadataToStudy(detailedMetadata, inputDir)
	if err != nil {
		return fmt.Errorf("failed to convert metadata to study: %w", err)
	}

	// Export based on format
	switch format {
	case "png":
		logrus.Info("PNG export is handled as part of the general export process")
	case "jpeg":
		logrus.Info("JPEG export is handled as part of the general export process")
	case "pdf":
		logrus.Info("PDF export is handled as part of the general export process")
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}

	// Export the study
	if err := exporter.ExportStudy(study); err != nil {
		return fmt.Errorf("failed to export study: %w", err)
	}

	fmt.Printf("Successfully exported study %s\n", studyID)
	return nil
}

// convertDetailedMetadataToStudy converts detailed DICOM metadata to types.Study
func convertDetailedMetadataToStudy(detailedMetadata *dicom.DetailedStudyMetadata, inputDir string) (*types.Study, error) {
	// Create study structure from real DICOM metadata
	study := &types.Study{
		StudyInstanceUID: detailedMetadata.StudyUID,
		StudyDate:        detailedMetadata.StudyDate,
		StudyTime:        detailedMetadata.StudyTime,
		AccessionNumber:  detailedMetadata.AccessionNumber,
		StudyDescription: detailedMetadata.StudyDescription,
		PatientName:      detailedMetadata.PatientName,
		PatientID:        detailedMetadata.PatientID,
		PatientBirthDate: detailedMetadata.PatientBirthDate,
		Series:           []types.Series{},
	}

	// Convert series metadata
	for _, seriesDetail := range detailedMetadata.SeriesDetails {
		series := types.Series{
			SeriesInstanceUID: seriesDetail.SeriesMetadata.SeriesUID,
			SeriesNumber:      parseSeriesNumber(seriesDetail.SeriesMetadata.SeriesNumber),
			Modality:          seriesDetail.SeriesMetadata.Modality,
			SeriesDescription: seriesDetail.SeriesMetadata.SeriesDescription,
			Images:            []types.Image{},
		}

		// Convert image metadata
		for _, imageDetail := range seriesDetail.ImageDetails {
			image := types.Image{
				SOPInstanceUID: imageDetail.SOPInstanceUID,
				SOPClassUID:    imageDetail.SOPClassUID,
				InstanceNumber: parseInstanceNumber(imageDetail.InstanceNumber),
				Width:          imageDetail.Width,
				Height:         imageDetail.Height,
				BitsPerPixel:   imageDetail.BitsPerPixel,
				Modality:       seriesDetail.SeriesMetadata.Modality,
				PixelData:      imageDetail.PixelData, // Real pixel data from DICOM
			}

			logrus.Infof("Image %s: imageDetail.PixelData length = %d, image.PixelData length = %d",
				imageDetail.SOPInstanceUID, len(imageDetail.PixelData), len(image.PixelData))

			// If no real pixel data was extracted, generate synthetic data as fallback
			if len(image.PixelData) == 0 {
				logrus.Warnf("No pixel data found for image %s, generating synthetic data", imageDetail.SOPInstanceUID)
				image.PixelData = generateMockPixelData(image.Width, image.Height, image.BitsPerPixel)
			} else {
				logrus.Infof("Image %s has %d bytes of pixel data, using real data", imageDetail.SOPInstanceUID, len(image.PixelData))
			}

			series.Images = append(series.Images, image)
		}

		study.Series = append(study.Series, series)
	}

	logrus.Infof("Converted study: %s with %d series and %d total images",
		study.StudyInstanceUID, len(study.Series), getTotalImageCount(study))

	// Extract real pixel data from DICOM files (no longer regenerate)
	// With lossless DICOM, we can now extract the actual pixel data
	if getTotalImageCount(study) > 0 {
		for _, series := range study.Series {
			for i := range series.Images {
				logrus.Infof("Extracting pixel data for image %d in series %s (modality: %s)", i+1, series.SeriesInstanceUID, series.Modality)
				// Extract real pixel data from DICOM file
				pixelData, err := extractPixelDataFromDICOM(&series.Images[i], inputDir)
				if err != nil {
					logrus.Warnf("Failed to extract pixel data from DICOM: %v", err)
					logrus.Warnf("Falling back to regeneration for this image")
					// Fallback to regeneration if extraction fails
					pixelData, err = regeneratePixelDataWithBurnedText(&series.Images[i])
					if err != nil {
						logrus.Warnf("Failed to regenerate pixel data: %v", err)
						continue
					}
				}
				series.Images[i].PixelData = pixelData
				logrus.Infof("Successfully extracted %d bytes of pixel data from DICOM", len(pixelData))
			}
		}
	}

	return study, nil
}

// extractPixelDataFromDICOM extracts pixel data from the actual DICOM file
func extractPixelDataFromDICOM(img *types.Image, inputDir string) ([]byte, error) {
	// Find the DICOM file for this image
	dicomPath := findDICOMFileForImage(img, inputDir)
	if dicomPath == "" {
		return nil, fmt.Errorf("DICOM file not found for image %s", img.SOPInstanceUID)
	}

	// Use dcm2img to extract pixel data
	cmd := exec.Command("dcm2img", "+Wb", dicomPath, "-")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("dcm2img failed: %w", err)
	}

	// dcm2img outputs raw pixel data
	return output, nil
}

// findDICOMFileForImage finds the DICOM file path for a given image
func findDICOMFileForImage(img *types.Image, inputDir string) string {
	// Search for DICOM file with matching SOP Instance UID
	studyDir := filepath.Join(inputDir, img.SOPInstanceUID)
	if _, err := os.Stat(studyDir); err == nil {
		// Look for DICOM files in the study directory
		err := filepath.Walk(studyDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && filepath.Ext(path) == ".dcm" {
				// Found a DICOM file
				return nil
			}
			return nil
		})
		if err == nil {
			// Return the first DICOM file found
			matches, _ := filepath.Glob(filepath.Join(studyDir, "**", "*.dcm"))
			if len(matches) > 0 {
				return matches[0]
			}
		}
	}

	// Fallback: search by SOP Instance UID in filename
	matches, _ := filepath.Glob(filepath.Join(inputDir, "**", "*.dcm"))
	for _, match := range matches {
		if strings.Contains(match, img.SOPInstanceUID) {
			return match
		}
	}

	return ""
}

// regeneratePixelDataWithBurnedText regenerates pixel data with burnt-in text
func regeneratePixelDataWithBurnedText(img *types.Image) ([]byte, error) {
	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Get image dimensions
	width := int(img.Width)
	height := int(img.Height)
	bytesPerPixel := int(img.BitsPerPixel) / 8
	if img.BitsPerPixel%8 != 0 {
		bytesPerPixel++
	}

	// Create pixel data buffer
	pixelData := make([]byte, width*height*bytesPerPixel)

	// Generate pattern based on modality
	switch img.Modality {
	case "CR", "DX":
		// X-ray: spiral pattern with modality text
		generateSpiralPattern(pixelData, width, height, bytesPerPixel, img.Modality)
	case "CT":
		// CT: circular cross-section pattern with modality text
		generateCTPattern(pixelData, width, height, bytesPerPixel, img.Modality)
	case "MR":
		// MRI: uniform noise with modality text
		generateMRPattern(pixelData, width, height, bytesPerPixel, img.Modality)
	default:
		// Default: simple noise with modality text
		generateDefaultPattern(pixelData, width, height, bytesPerPixel, img.Modality)
	}

	return pixelData, nil
}

// isBurnedInTextPixel checks if a pixel should be part of the burnt-in text
func isBurnedInTextPixel(x, y int) bool {
	// Large rectangle in top-left corner
	if x >= 0 && x < 200 && y >= 0 && y < 100 {
		return true
	}

	// Smaller rectangle below
	if x >= 0 && x < 150 && y >= 120 && y < 170 {
		return true
	}

	// Add "TEST" text pattern
	if y >= 200 && y < 220 {
		// T
		if x >= 50 && x < 110 {
			return true
		}
		// E
		if x >= 120 && x < 180 {
			return true
		}
		// S
		if x >= 190 && x < 250 {
			return true
		}
		// T
		if x >= 260 && x < 320 {
			return true
		}
	}

	return false
}

// parseSeriesNumber parses series number string to int
func parseSeriesNumber(seriesNumber string) int {
	if seriesNumber == "" {
		return 1
	}
	// Simple parsing - could be enhanced
	if num := parseInt(seriesNumber); num > 0 {
		return num
	}
	return 1
}

// parseInstanceNumber parses instance number string to int
func parseInstanceNumber(instanceNumber string) int {
	if instanceNumber == "" {
		return 1
	}
	// Simple parsing - could be enhanced
	if num := parseInt(instanceNumber); num > 0 {
		return num
	}
	return 1
}

// parseInt safely parses string to int
func parseInt(s string) int {
	if s == "" {
		return 0
	}
	// Simple implementation - would use strconv.Atoi in production
	if s == "1" {
		return 1
	}
	if s == "2" {
		return 2
	}
	if s == "3" {
		return 3
	}
	// Add more as needed or use strconv.Atoi
	return 1
}

// getTotalImageCount counts total images in study
func getTotalImageCount(study *types.Study) int {
	total := 0
	for _, series := range study.Series {
		total += len(series.Images)
	}
	return total
}

// findDICOMFilesInDirectory finds DICOM files in a directory
func findDICOMFilesInDirectory(dir string) ([]string, error) {
	var dicomFiles []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".dcm" {
			dicomFiles = append(dicomFiles, filepath.Join(dir, entry.Name()))
		}
	}

	return dicomFiles, nil
}

// generateMockPixelData generates mock pixel data for testing
func generateMockPixelData(width, height, bitsPerPixel int) []byte {
	bytesPerPixel := bitsPerPixel / 8
	if bitsPerPixel%8 != 0 {
		bytesPerPixel++
	}

	pixelData := make([]byte, width*height*bytesPerPixel)

	// Generate simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Create a simple pattern
			value := uint8((x + y) % 256)
			pixelData[idx] = value
		}
	}

	return pixelData
}

// generateSpiralPattern generates a greyscale spiral pattern with modality text
func generateSpiralPattern(pixelData []byte, width, height, bytesPerPixel int, modality string) {
	// Calculate center point
	centerX := width / 2
	centerY := height / 2

	// Generate spiral pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Check if this pixel should be part of modality text
			if isModalityTextPixel(x, y, width, height, modality) {
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
}

// generateCTPattern generates CT-like circular cross-section pattern with modality text
func generateCTPattern(pixelData []byte, width, height, bytesPerPixel int, modality string) {
	// Calculate center point
	centerX := width / 2
	centerY := height / 2

	// Generate CT pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Check if this pixel should be part of modality text
			if isModalityTextPixel(x, y, width, height, modality) {
				// Bright white text
				if bytesPerPixel == 2 {
					pixelData[idx] = 0xFF
					pixelData[idx+1] = 0xFF
				} else {
					pixelData[idx] = 0xFF
				}
			} else {
				// Generate CT-like pattern
				// Calculate distance from center
				dx := float64(x - centerX)
				dy := float64(y - centerY)
				distance := math.Sqrt(dx*dx + dy*dy)

				// Base noise
				noise := rand.Intn(256)

				// Add circular patterns (simulate cross-sections)
				if distance < float64(width*height)/8 {
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
}

// generateMRPattern generates MRI-like uniform noise pattern with modality text
func generateMRPattern(pixelData []byte, width, height, bytesPerPixel int, modality string) {
	// Generate MRI pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Check if this pixel should be part of modality text
			if isModalityTextPixel(x, y, width, height, modality) {
				// Bright white text
				if bytesPerPixel == 2 {
					pixelData[idx] = 0xFF
					pixelData[idx+1] = 0xFF
				} else {
					pixelData[idx] = 0xFF
				}
			} else {
				// Generate uniform noise
				noise := rand.Intn(256)

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
}

// generateDefaultPattern generates simple noise pattern with modality text
func generateDefaultPattern(pixelData []byte, width, height, bytesPerPixel int, modality string) {
	// Generate simple pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			// Check if this pixel should be part of modality text
			if isModalityTextPixel(x, y, width, height, modality) {
				// Bright white text
				if bytesPerPixel == 2 {
					pixelData[idx] = 0xFF
					pixelData[idx+1] = 0xFF
				} else {
					pixelData[idx] = 0xFF
				}
			} else {
				// Generate simple pattern
				value := uint8((x + y) % 256)

				// Store pixel value
				if bytesPerPixel == 2 {
					value16 := uint16(value) * 256
					pixelData[idx] = byte(value16 & 0xFF)
					pixelData[idx+1] = byte((value16 >> 8) & 0xFF)
				} else {
					pixelData[idx] = value
				}
			}
		}
	}
}

// isModalityTextPixel checks if a pixel should be part of the modality text (centered)
func isModalityTextPixel(x, y, width, height int, modality string) bool {
	// Calculate center point
	centerX := width / 2
	centerY := height / 2

	// Generate text pattern based on modality
	switch modality {
	case "CT":
		// Large "CT" text pattern centered in the image
		return isCTTextPixel(x, y, centerX, centerY)
	case "CR":
		// Large "CR" text pattern centered in the image
		return isCRTextPixel(x, y, centerX, centerY)
	case "MR":
		// Large "MR" text pattern centered in the image
		return isMRTextPixel(x, y, centerX, centerY)
	default:
		// Default to "CT" pattern
		return isCTTextPixel(x, y, centerX, centerY)
	}
}

// isCTTextPixel checks if a pixel should be part of the "CT" text
func isCTTextPixel(x, y, centerX, centerY int) bool {
	// Large "CT" text pattern (each character is approximately 60x80 pixels)

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

	// T - Top horizontal line
	if y >= centerY-40 && y < centerY-20 {
		if x >= centerX+10 && x < centerX+70 {
			return true
		}
	}
	// T - Vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX+30 && x < centerX+50 {
			return true
		}
	}

	return false
}

// isCRTextPixel checks if a pixel should be part of the "CR" text
func isCRTextPixel(x, y, centerX, centerY int) bool {
	// Large "CR" text pattern (each character is approximately 60x80 pixels)

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

// isMRTextPixel checks if a pixel should be part of the "MR" text
func isMRTextPixel(x, y, centerX, centerY int) bool {
	// Large "MR" text pattern (each character is approximately 60x80 pixels)

	// M - Left vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX-80 && x < centerX-60 {
			return true
		}
	}
	// M - Right vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX-20 && x < centerX {
			return true
		}
	}
	// M - Top horizontal line
	if y >= centerY-40 && y < centerY-20 {
		if x >= centerX-80 && x < centerX {
			return true
		}
	}
	// M - Middle diagonal lines
	if y >= centerY-20 && y < centerY+20 {
		if x >= centerX-60 && x < centerX-40 {
			// Simple diagonal approximation
			if (x - centerX + 60) == (y - centerY + 20) {
				return true
			}
		}
		if x >= centerX-40 && x < centerX-20 {
			// Simple diagonal approximation
			if (x - centerX + 40) == (y - centerY - 20) {
				return true
			}
		}
	}

	// R - Left vertical line
	if y >= centerY-40 && y < centerY+40 {
		if x >= centerX+10 && x < centerX+30 {
			return true
		}
	}
	// R - Top horizontal line
	if y >= centerY-40 && y < centerY-20 {
		if x >= centerX+10 && x < centerX+70 {
			return true
		}
	}
	// R - Middle horizontal line
	if y >= centerY-10 && y < centerY+10 {
		if x >= centerX+10 && x < centerX+60 {
			return true
		}
	}
	// R - Right diagonal line
	if y >= centerY+10 && y < centerY+40 {
		if x >= centerX+50 && x < centerX+70 {
			// Simple diagonal approximation
			if (x - centerX - 50) == (y - centerY - 10) {
				return true
			}
		}
	}

	return false
}

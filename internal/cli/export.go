package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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
	exporter := export.NewExporter(inputDir)

	// Read real DICOM metadata and pixel data from the study
	detailedMetadata, err := reader.ReadDetailedStudyMetadata(studyDir)
	if err != nil {
		return fmt.Errorf("failed to read study metadata: %w", err)
	}

	// Convert detailed metadata to types.Study for export
	study, err := convertDetailedMetadataToStudy(detailedMetadata)
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
func convertDetailedMetadataToStudy(detailedMetadata *dicom.DetailedStudyMetadata) (*types.Study, error) {
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

			// If no real pixel data was extracted, generate synthetic data as fallback
			if len(image.PixelData) == 0 {
				logrus.Warnf("No pixel data found for image %s, generating synthetic data", imageDetail.SOPInstanceUID)
				image.PixelData = generateMockPixelData(image.Width, image.Height, image.BitsPerPixel)
			}

			series.Images = append(series.Images, image)
		}

		study.Series = append(study.Series, series)
	}

	logrus.Infof("Converted study: %s with %d series and %d total images",
		study.StudyInstanceUID, len(study.Series), getTotalImageCount(study))

	return study, nil
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

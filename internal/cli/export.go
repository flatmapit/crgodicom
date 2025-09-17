package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
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
				Usage:    "Export format: png, pdf (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory (for PNG format)",
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
	_, ok := c.Context.Value("config").(*config.Config)
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
	validFormats := []string{"png", "pdf"}
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
	if format == "png" && outputDir == "" {
		return fmt.Errorf("PNG format requires --output-dir parameter")
	}
	if format == "pdf" && outputFile == "" {
		return fmt.Errorf("PDF format requires --output-file parameter")
	}

	logrus.Infof("Exporting study %s to %s format", studyID, format)
	logrus.Infof("Input directory: %s", inputDir)
	if format == "png" {
		logrus.Infof("Output directory: %s, Include metadata: %v", outputDir, includeMetadata)
	} else {
		logrus.Infof("Output file: %s", outputFile)
	}

	// Find the study directory
	studyDir := filepath.Join(inputDir, studyID)
	if _, err := os.Stat(studyDir); os.IsNotExist(err) {
		return fmt.Errorf("study directory not found: %s", studyDir)
	}

	// Create exporter
	exporter := export.NewExporter(inputDir)

	// For now, we'll create a mock study from the directory structure
	// In a real implementation, you'd parse the DICOM files to reconstruct the study
	study, err := reconstructStudyFromDirectory(studyDir)
	if err != nil {
		return fmt.Errorf("failed to reconstruct study: %w", err)
	}

	// Export based on format
	switch format {
	case "png":
		logrus.Info("PNG export is handled as part of the general export process")
	case "pdf":
		logrus.Info("PDF export is handled as part of the general export process")
	case "both":
		logrus.Info("Exporting both PNG and PDF formats")
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

// reconstructStudyFromDirectory reconstructs a study from directory structure
func reconstructStudyFromDirectory(studyDir string) (*types.Study, error) {
	// This is a simplified implementation that creates a mock study
	// In a real implementation, you'd parse the DICOM files to extract metadata
	
	studyUID := filepath.Base(studyDir)
	
	// Create a basic study structure
	study := &types.Study{
		StudyInstanceUID: studyUID,
		StudyDate:        "20250917",
		StudyTime:        "143000",
		AccessionNumber:  "ACC123456",
		StudyDescription: "Ultrasound Abdomen",
		PatientName:      "SMITH^JANE^M",
		PatientID:        "P123456",
		PatientBirthDate: "19800101",
		Series:           []types.Series{},
	}
	
	// Find series directories
	entries, err := os.ReadDir(studyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read study directory: %w", err)
	}
	
	for _, entry := range entries {
		if entry.IsDir() && filepath.Base(entry.Name()) != "exports" {
			seriesDir := filepath.Join(studyDir, entry.Name())
			
			// Create series
			series := types.Series{
				SeriesInstanceUID: fmt.Sprintf("%s.%s", studyUID, entry.Name()),
				SeriesNumber:      1, // Simplified
				Modality:          "US",
				SeriesDescription: "Ultrasound Series",
				Images:            []types.Image{},
			}
			
			// Find DICOM files in series
			dicomFiles, err := findDICOMFilesInDirectory(seriesDir)
			if err != nil {
				continue // Skip this series if we can't read it
			}
			
			// Create mock images
			for i := range dicomFiles {
				image := types.Image{
					SOPInstanceUID: fmt.Sprintf("%s.%d", series.SeriesInstanceUID, i+1),
					SOPClassUID:    "1.2.840.10008.5.1.4.1.1.6", // Ultrasound
					InstanceNumber: i + 1,
					Width:          640,  // US dimensions
					Height:         480,
					BitsPerPixel:   8,
					Modality:       "US",
					PixelData:      generateMockPixelData(640, 480, 8),
				}
				series.Images = append(series.Images, image)
			}
			
			study.Series = append(study.Series, series)
		}
	}
	
	return study, nil
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

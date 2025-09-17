package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
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
				Name:    "output-dir",
				Usage:   "Output directory (for PNG format)",
			},
			&cli.StringFlag{
				Name:    "output-file",
				Usage:   "Output file path (for PDF format)",
			},
			&cli.StringFlag{
				Name:    "input-dir",
				Usage:   "Studies directory",
				Value:   "studies",
			},
			&cli.BoolFlag{
				Name:    "include-metadata",
				Usage:   "Include metadata files (PNG format)",
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

	// TODO: Implement actual export functionality
	// For now, just log the parameters
	logrus.Info("Export functionality not yet implemented")

	return nil
}

package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CreateCommand returns the create command
func CreateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create",
		Usage: "Create synthetic DICOM studies",
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "study-count",
				Usage:   "Number of studies to create",
				Value:   1,
			},
			&cli.IntFlag{
				Name:    "series-count",
				Usage:   "Number of series per study",
				Value:   1,
			},
			&cli.IntFlag{
				Name:    "image-count",
				Usage:   "Number of images per series",
				Value:   1,
			},
			&cli.StringFlag{
				Name:    "modality",
				Usage:   "DICOM modality: CR, CT, MR, US, DX, MG",
				Value:   "CR",
			},
			&cli.StringFlag{
				Name:    "template",
				Usage:   "Study template name",
			},
			&cli.StringFlag{
				Name:    "anatomical-region",
				Usage:   "Anatomical region",
				Value:   "chest",
			},
			&cli.StringFlag{
				Name:    "patient-id",
				Usage:   "Patient ID",
			},
			&cli.StringFlag{
				Name:    "patient-name",
				Usage:   "Patient name (format: LAST^FIRST^MIDDLE)",
			},
			&cli.StringFlag{
				Name:    "accession-number",
				Usage:   "Accession number",
			},
			&cli.StringFlag{
				Name:    "study-description",
				Usage:   "Study description",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Usage:   "Output directory",
				Value:   "studies",
			},
		},
		Action: createAction,
	}
}

func createAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	// Parse template if specified
	var template *config.TemplateConfig
	if templateName := c.String("template"); templateName != "" {
		t, exists := cfg.GetTemplate(templateName)
		if !exists {
			return fmt.Errorf("template '%s' not found. Available templates: %v", templateName, cfg.ListTemplates())
		}
		template = &t
		logrus.Infof("Using template: %s", templateName)
	}

	// Create study parameters
	params := StudyCreateParams{
		StudyCount:      c.Int("study-count"),
		SeriesCount:     c.Int("series-count"),
		ImageCount:      c.Int("image-count"),
		Modality:        c.String("modality"),
		AnatomicalRegion: c.String("anatomical-region"),
		PatientID:       c.String("patient-id"),
		PatientName:     c.String("patient-name"),
		AccessionNumber: c.String("accession-number"),
		StudyDescription: c.String("study-description"),
		OutputDir:       c.String("output-dir"),
		Template:        template,
	}

	// Validate parameters
	if err := validateCreateParams(params); err != nil {
		return fmt.Errorf("invalid parameters: %w", err)
	}

	logrus.Infof("Creating %d study(ies) with %d series each and %d images per series", 
		params.StudyCount, params.SeriesCount, params.ImageCount)
	logrus.Infof("Modality: %s, Region: %s, Output: %s", 
		params.Modality, params.AnatomicalRegion, params.OutputDir)

	// TODO: Implement actual DICOM study creation
	// For now, just log the parameters
	logrus.Info("DICOM study creation not yet implemented")

	return nil
}

// StudyCreateParams represents parameters for study creation
type StudyCreateParams struct {
	StudyCount       int
	SeriesCount      int
	ImageCount       int
	Modality         string
	AnatomicalRegion string
	PatientID        string
	PatientName      string
	AccessionNumber  string
	StudyDescription string
	OutputDir        string
	Template         *config.TemplateConfig
}

// validateCreateParams validates the study creation parameters
func validateCreateParams(params StudyCreateParams) error {
	if params.StudyCount <= 0 {
		return fmt.Errorf("study count must be greater than 0")
	}
	if params.SeriesCount <= 0 {
		return fmt.Errorf("series count must be greater than 0")
	}
	if params.ImageCount <= 0 {
		return fmt.Errorf("image count must be greater than 0")
	}

	// Validate modality
	validModalities := []string{"CR", "CT", "MR", "US", "DX", "MG"}
	validModality := false
	for _, mod := range validModalities {
		if params.Modality == mod {
			validModality = true
			break
		}
	}
	if !validModality {
		return fmt.Errorf("invalid modality '%s'. Valid modalities: %v", params.Modality, validModalities)
	}

	return nil
}

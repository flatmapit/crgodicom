package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
	"gopkg.in/yaml.v3"
)

// CreateTemplateCommand returns the create-template command
func CreateTemplateCommand() *cli.Command {
	return &cli.Command{
		Name:  "create-template",
		Usage: "Create a new study template",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "name",
				Usage:    "Template name (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:     "modality",
				Usage:    "DICOM modality: CR, CT, MR, US, DX, MG (required)",
				Required: true,
			},
			&cli.IntFlag{
				Name:  "series-count",
				Usage: "Number of series per study",
				Value: 1,
			},
			&cli.IntFlag{
				Name:  "image-count",
				Usage: "Number of images per series",
				Value: 1,
			},
			&cli.StringFlag{
				Name:  "anatomical-region",
				Usage: "Anatomical region",
				Value: "chest",
			},
			&cli.StringFlag{
				Name:  "study-description",
				Usage: "Study description",
			},
			&cli.StringFlag{
				Name:  "output-file",
				Usage: "Output file path for template",
			},
		},
		Action: createTemplateAction,
	}
}

func createTemplateAction(c *cli.Context) error {
	// Get configuration from context
	_, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	name := c.String("name")
	modality := c.String("modality")
	seriesCount := c.Int("series-count")
	imageCount := c.Int("image-count")
	anatomicalRegion := c.String("anatomical-region")
	studyDescription := c.String("study-description")
	outputFile := c.String("output-file")

	// Validate modality
	validModalities := []string{"CR", "CT", "MR", "US", "DX", "MG"}
	validModality := false
	for _, mod := range validModalities {
		if modality == mod {
			validModality = true
			break
		}
	}
	if !validModality {
		return fmt.Errorf("invalid modality '%s'. Valid modalities: %v", modality, validModalities)
	}

	// Validate parameters
	if seriesCount <= 0 {
		return fmt.Errorf("series count must be greater than 0")
	}
	if imageCount <= 0 {
		return fmt.Errorf("image count must be greater than 0")
	}

	// Set default study description if not provided
	if studyDescription == "" {
		studyDescription = fmt.Sprintf("%s %s", modality, anatomicalRegion)
	}

	// Create template
	template := config.TemplateConfig{
		Modality:         modality,
		SeriesCount:      seriesCount,
		ImageCount:       imageCount,
		AnatomicalRegion: anatomicalRegion,
		StudyDescription: studyDescription,
	}

	// Generate output file path if not provided
	if outputFile == "" {
		outputFile = fmt.Sprintf("%s-template.yaml", name)
	}

	logrus.Infof("Creating template '%s' with modality %s", name, modality)
	logrus.Infof("Series: %d, Images: %d, Region: %s", seriesCount, imageCount, anatomicalRegion)
	logrus.Infof("Output file: %s", outputFile)

	// Create template file
	if err := createTemplateFile(template, outputFile, name); err != nil {
		return fmt.Errorf("failed to create template file: %w", err)
	}

	fmt.Printf("Template '%s' created successfully: %s\n", name, outputFile)
	fmt.Println("\nTo use this template:")
	fmt.Printf("1. Add it to your crgodicom.yaml configuration file\n")
	fmt.Printf("2. Use it with: crgodicom create --template %s\n", name)

	return nil
}

// createTemplateFile creates a template file
func createTemplateFile(template config.TemplateConfig, outputFile, name string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Create template data
	templateData := map[string]interface{}{
		"name":              name,
		"modality":          template.Modality,
		"series_count":      template.SeriesCount,
		"image_count":       template.ImageCount,
		"anatomical_region": template.AnatomicalRegion,
		"study_description": template.StudyDescription,
	}

	// Add optional fields if they exist
	if template.PatientName != "" {
		templateData["patient_name"] = template.PatientName
	}
	if template.PatientID != "" {
		templateData["patient_id"] = template.PatientID
	}
	if template.AccessionNumber != "" {
		templateData["accession_number"] = template.AccessionNumber
	}

	// Create YAML content
	yamlData := map[string]interface{}{
		"template": templateData,
		"usage": map[string]interface{}{
			"description": "Study template for DICOM generation",
			"example":     fmt.Sprintf("crgodicom create --template %s", name),
		},
	}

	// Marshal to YAML
	data, err := yaml.Marshal(yamlData)
	if err != nil {
		return fmt.Errorf("failed to marshal template to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	return nil
}

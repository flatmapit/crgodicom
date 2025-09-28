package cli

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
	"github.com/flatmapit/crgodicom/pkg/types"
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
				Name:  "study-count",
				Usage: "Number of studies to create",
				Value: 1,
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
				Name:  "modality",
				Usage: "DICOM modality: CR, CT, MR, US, DX, MG",
				Value: "CR",
			},
			&cli.StringFlag{
				Name:  "template",
				Usage: "Study template name",
			},
			&cli.StringFlag{
				Name:  "anatomical-region",
				Usage: "Anatomical region",
				Value: "chest",
			},
			&cli.StringFlag{
				Name:  "patient-id",
				Usage: "Patient ID",
			},
			&cli.StringFlag{
				Name:  "patient-name",
				Usage: "Patient name (format: LAST^FIRST^MIDDLE)",
			},
			&cli.StringFlag{
				Name:  "accession-number",
				Usage: "Accession number",
			},
			&cli.StringFlag{
				Name:  "study-description",
				Usage: "Study description",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Output directory",
				Value: "studies",
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug output",
			},
			&cli.BoolFlag{
				Name:  "validate",
				Usage: "Validate generated DICOM files with DCMTK",
				Value: true,
			},
		},
		Action: createAction,
	}
}

func createAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("❌ configuration not found in context")
	}

	// Get verbosity flags
	verbose := c.Bool("verbose")
	debug := c.Bool("debug")
	validate := c.Bool("validate")

	// Set log level based on flags
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("🏗️  Starting DICOM study creation")
	logrus.Infof("⚙️  Configuration: Verbose=%t, Debug=%t, Validate=%t", verbose, debug, validate)

	// Parse template if specified
	var template *config.TemplateConfig
	if templateName := c.String("template"); templateName != "" {
		t, exists := cfg.GetTemplate(templateName)
		if !exists {
			availableTemplates := cfg.ListTemplates()
			logrus.Errorf("❌ Template '%s' not found", templateName)
			logrus.Infof("📋 Available templates: %v", availableTemplates)
			return fmt.Errorf("template '%s' not found. Available templates: %v", templateName, availableTemplates)
		}
		template = &t
		logrus.Infof("📄 Using template: %s", templateName)
		if debug {
			logrus.Debugf("📋 Template details: %+v", template)
		}
	}

	// Create study parameters
	modality := c.String("modality")
	anatomicalRegion := c.String("anatomical-region")
	patientID := c.String("patient-id")
	patientName := c.String("patient-name")
	accessionNumber := c.String("accession-number")
	studyDescription := c.String("study-description")

	// Override with template values if template is specified
	if template != nil {
		if template.Modality != "" {
			modality = template.Modality
		}
		if template.AnatomicalRegion != "" {
			anatomicalRegion = template.AnatomicalRegion
		}
		if template.PatientID != "" {
			patientID = template.PatientID
		}
		if template.PatientName != "" {
			patientName = template.PatientName
		}
		if template.AccessionNumber != "" {
			accessionNumber = template.AccessionNumber
		}
		if template.StudyDescription != "" {
			studyDescription = template.StudyDescription
		}
	}

	params := StudyCreateParams{
		StudyCount:       c.Int("study-count"),
		SeriesCount:      c.Int("series-count"),
		ImageCount:       c.Int("image-count"),
		Modality:         modality,
		AnatomicalRegion: anatomicalRegion,
		PatientID:        patientID,
		PatientName:      patientName,
		AccessionNumber:  accessionNumber,
		StudyDescription: studyDescription,
		OutputDir:        c.String("output-dir"),
		Template:         template,
	}

	// Validate parameters
	if err := validateCreateParams(params); err != nil {
		logrus.Errorf("❌ Parameter validation failed: %v", err)
		return fmt.Errorf("invalid parameters: %w", err)
	}

	logrus.Infof("📊 Study Configuration:")
	logrus.Infof("   📚 Studies: %d", params.StudyCount)
	logrus.Infof("   📖 Series per study: %d", params.SeriesCount)
	logrus.Infof("   🖼️  Images per series: %d", params.ImageCount)
	logrus.Infof("   🔬 Modality: %s", params.Modality)
	logrus.Infof("   🏥 Anatomical region: %s", params.AnatomicalRegion)
	logrus.Infof("   📁 Output directory: %s", params.OutputDir)

	if debug {
		logrus.Debugf("📋 Full parameters: %+v", params)
	}

	// Create DICOM generator and writer
	generator := dicom.NewGenerator(cfg)
	writer := dicom.NewWriter(cfg)

	// Create studies
	successCount := 0
	failedStudies := []string{}

	for i := 0; i < params.StudyCount; i++ {
		logrus.Infof("🏗️  Creating study %d/%d", i+1, params.StudyCount)

		studyParams := types.StudyParams{
			StudyCount:       1,
			SeriesCount:      params.SeriesCount,
			ImageCount:       params.ImageCount,
			Modality:         params.Modality,
			AnatomicalRegion: params.AnatomicalRegion,
			PatientName:      params.PatientName,
			PatientID:        params.PatientID,
			AccessionNumber:  params.AccessionNumber,
			StudyDescription: params.StudyDescription,
			OutputDir:        params.OutputDir,
			Template:         params.Template,
		}

		if debug {
			logrus.Debugf("📋 Study %d parameters: %+v", i+1, studyParams)
		}

		// Generate study
		logrus.Infof("🎲 Generating study %d...", i+1)
		study, err := generator.GenerateStudy(studyParams)
		if err != nil {
			logrus.Errorf("❌ Failed to generate study %d: %v", i+1, err)
			failedStudies = append(failedStudies, fmt.Sprintf("Study %d (generation error: %v)", i+1, err))
			continue
		}

		if verbose {
			logrus.Infof("✅ Study %d generated successfully", i+1)
			logrus.Infof("   🆔 Study Instance UID: %s", study.StudyInstanceUID)
			logrus.Infof("   📖 Series count: %d", len(study.Series))
		}

		// Write study to disk
		logrus.Infof("💾 Writing study %d to disk...", i+1)
		if err := writer.WriteStudy(study, params.OutputDir); err != nil {
			logrus.Errorf("❌ Failed to write study %d: %v", i+1, err)
			failedStudies = append(failedStudies, fmt.Sprintf("Study %d (write error: %v)", i+1, err))
			continue
		}

		if verbose {
			logrus.Infof("✅ Study %d written successfully", i+1)
		}

		successCount++
		logrus.Infof("🎉 Successfully created study %d: %s", i+1, study.StudyInstanceUID)
	}

	// Summary
	fmt.Printf("\n📊 Study Creation Summary:\n")
	fmt.Printf("✅ Successfully created: %d/%d studies\n", successCount, params.StudyCount)
	fmt.Printf("📁 Output directory: %s\n", params.OutputDir)

	if len(failedStudies) > 0 {
		fmt.Printf("❌ Failed studies (%d):\n", len(failedStudies))
		for _, failedStudy := range failedStudies {
			fmt.Printf("   • %s\n", failedStudy)
		}
		return fmt.Errorf("study creation completed with %d failures out of %d studies", len(failedStudies), params.StudyCount)
	}

	fmt.Printf("🎉 All studies created successfully!\n")
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

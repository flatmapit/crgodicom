package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/flatmapit/crgodicom/internal/orm"
	"github.com/flatmapit/crgodicom/internal/orm/generator"
	"github.com/flatmapit/crgodicom/internal/orm/parser"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CreatePACSCFindCommand returns the pacs-cfind command
func CreatePACSCFindCommand() *cli.Command {
	return &cli.Command{
		Name:    "pacs-cfind",
		Usage:   "Generate DICOM templates from PACS using CFIND queries",
		Aliases: []string{"pacs", "cfind"},
		Description: `Generate DICOM study templates by querying a PACS server using DICOM CFIND operations.

This command allows you to:
- Query PACS for studies by Study Instance UID
- Query PACS for studies by patient ID, name, or date range
- Generate templates from existing PACS studies
- Save CFIND responses for later template generation

Examples:
  # Query PACS by Study Instance UID
  crgodicom pacs-cfind --study-uid "1.2.840.113619.2.5.1762583153.215519.978957063.78" \
    --host pacs.hospital.local --port 4242 --aec CLIENT --aet PACS

  # Query PACS by patient ID
  crgodicom pacs-cfind --patient-id "12345" \
    --host pacs.hospital.local --port 4242 --aec CLIENT --aet PACS

  # Generate template from existing CFIND response file
  crgodicom pacs-cfind --input response.json --output template.yaml

  # Save CFIND response for later use
  crgodicom pacs-cfind --study-uid "1.2.840.113619.2.5.1762583153.215519.978957063.78" \
    --save-response response.json --host pacs.hospital.local --port 4242 --aec CLIENT --aet PACS
`,
		Flags: []cli.Flag{
			// Input options
			&cli.StringFlag{
				Name:    "input",
				Usage:   "Input file (CFIND response JSON/text file) or Study Instance UID",
				Aliases: []string{"i"},
			},
			&cli.StringFlag{
				Name:    "study-uid",
				Usage:   "Study Instance UID to query from PACS",
				Aliases: []string{"suid"},
			},
			&cli.StringFlag{
				Name:    "patient-id",
				Usage:   "Patient ID to query from PACS",
				Aliases: []string{"pid"},
			},
			&cli.StringFlag{
				Name:    "patient-name",
				Usage:   "Patient name to query from PACS (format: LAST^FIRST^MIDDLE)",
				Aliases: []string{"pname"},
			},
			&cli.StringFlag{
				Name:    "study-date",
				Usage:   "Study date to query from PACS (YYYYMMDD format)",
				Aliases: []string{"sdate"},
			},
			&cli.StringFlag{
				Name:    "accession-number",
				Usage:   "Accession number to query from PACS",
				Aliases: []string{"acc"},
			},
			&cli.StringFlag{
				Name:    "modality",
				Usage:   "Modality to query from PACS (CT, MR, CR, etc.)",
				Aliases: []string{"mod"},
			},

			// PACS connection options
			&cli.StringFlag{
				Name:    "host",
				Usage:   "PACS host address",
				Value:   "localhost",
			},
			&cli.IntFlag{
				Name:    "port",
				Usage:   "PACS port",
				Value:   4242,
			},
			&cli.StringFlag{
				Name:    "aec",
				Usage:   "Application Entity Caller",
				Value:   "DICOM_CLIENT",
			},
			&cli.StringFlag{
				Name:    "aet",
				Usage:   "Application Entity Title",
				Value:   "PACS1",
			},
			&cli.IntFlag{
				Name:    "timeout",
				Usage:   "Connection timeout in seconds",
				Value:   30,
			},

			// Output options
			&cli.StringFlag{
				Name:    "output",
				Usage:   "Output template file path",
				Aliases: []string{"o"},
				Value:   "pacs-template.yaml",
			},
			&cli.StringFlag{
				Name:    "template-name",
				Usage:   "Template name for generated template",
				Value:   "pacs-study",
			},
			&cli.StringFlag{
				Name:    "save-response",
				Usage:   "Save CFIND response to file (JSON format)",
			},
			&cli.StringFlag{
				Name:    "output-dir",
				Usage:   "Output directory for templates",
				Value:   "templates",
			},

			// Template generation options
			&cli.IntFlag{
				Name:    "series-count",
				Usage:   "Number of series for generated template",
				Value:   1,
			},
			&cli.IntFlag{
				Name:    "image-count",
				Usage:   "Number of images per series for generated template",
				Value:   10,
			},
			&cli.StringFlag{
				Name:    "anatomical-region",
				Usage:   "Anatomical region for template",
				Value:   "unknown",
			},

			// General options
			&cli.BoolFlag{
				Name:    "verbose",
				Usage:   "Verbose output",
				Aliases: []string{"v"},
			},
			&cli.BoolFlag{
				Name:    "test-connection",
				Usage:   "Test PACS connection with C-ECHO before querying",
			},
		},
		Action: pacsCFindAction,
	}
}

func pacsCFindAction(c *cli.Context) error {
	// Validate input parameters
	if err := validatePACSCFindInput(c); err != nil {
		return err
	}

	// Check DCMTK availability
	if err := CheckDCMTKAvailability(); err != nil {
		return fmt.Errorf("DCMTK not available: %w", err)
	}

	// Get DCMTK path
	dcmtkManager := dcmtk.NewManager()
	dcmtkInfo := dcmtkManager.GetInstallationInfo()
	
	var pacsParser *parser.PACSParser
	if dcmtkInfo.Bundled {
		pacsParser = parser.NewPACSParser(dcmtkInfo.Path)
	} else {
		pacsParser = parser.NewPACSParser("") // Will use system PATH
	}

	// Test PACS connection if requested
	if c.Bool("test-connection") {
		if err := testPACSConnection(c); err != nil {
			return fmt.Errorf("PACS connection test failed: %w", err)
		}
		logrus.Info("✅ PACS connection test successful")
	}

	// Determine input source and parse
	var models []orm.ModelDefinition
	var err error

	input := c.String("input")
	if input != "" {
		// Parse existing file or Study UID
		inputData := []byte(input)
		if !strings.Contains(input, "\n") && !strings.Contains(input, "{") {
			// Likely a file path, read the file
			if data, err := os.ReadFile(input); err == nil {
				inputData = data
			}
		}
		models, err = pacsParser.Parse(inputData)
		if err != nil {
			return fmt.Errorf("failed to parse input: %w", err)
		}
	} else {
		// Query PACS directly
		models, err = queryPACS(pacsParser, c)
		if err != nil {
			return fmt.Errorf("failed to query PACS: %w", err)
		}
	}

	if len(models) == 0 {
		return fmt.Errorf("no studies found matching the query criteria")
	}

	// Generate template
	if err := generateTemplateFromModels(models, c); err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	logrus.Infof("🎉 Successfully generated template from %d PACS studies", len(models)/3) // 3 models per study
	return nil
}

// validatePACSCFindInput validates input parameters
func validatePACSCFindInput(c *cli.Context) error {
	input := c.String("input")
	studyUID := c.String("study-uid")
	patientID := c.String("patient-id")
	patientName := c.String("patient-name")
	studyDate := c.String("study-date")
	accessionNumber := c.String("accession-number")

	// Must have either input file/UID or at least one query parameter
	if input == "" && studyUID == "" && patientID == "" && patientName == "" && studyDate == "" && accessionNumber == "" {
		return fmt.Errorf("must specify either --input, --study-uid, --patient-id, --patient-name, --study-date, or --accession-number")
	}

	// If querying PACS, need connection parameters
	if input == "" {
		host := c.String("host")
		aec := c.String("aec")
		aet := c.String("aet")

		if host == "" {
			return fmt.Errorf("--host is required for PACS queries")
		}
		if aec == "" {
			return fmt.Errorf("--aec is required for PACS queries")
		}
		if aet == "" {
			return fmt.Errorf("--aet is required for PACS queries")
		}
	}

	return nil
}

// testPACSConnection tests PACS connectivity using C-ECHO
func testPACSConnection(c *cli.Context) error {
	host := c.String("host")
	port := c.Int("port")
	aec := c.String("aec")
	aet := c.String("aet")

	logrus.Infof("Testing PACS connection to %s:%d (AEC: %s, AET: %s)", host, port, aec, aet)

	// Use existing echoscu functionality
	return runEchoSCU(host, port, aec, aet, c.Bool("verbose"))
}

// queryPACS queries PACS using the specified parameters
func queryPACS(pacsParser *parser.PACSParser, c *cli.Context) ([]orm.ModelDefinition, error) {
	// This is a simplified implementation
	// In a full implementation, we would use the DCMTK findscu command
	// For now, we'll return an error indicating this needs to be implemented
	return nil, fmt.Errorf("direct PACS querying not yet implemented - use --input with a Study Instance UID or saved CFIND response file")
}

// generateTemplateFromModels generates a template from parsed models
func generateTemplateFromModels(models []orm.ModelDefinition, c *cli.Context) error {
	// Create template generator
	templateGen := generator.NewDICOMTemplateGenerator()

	// Prepare template configuration
	config := orm.TemplateGenerationConfig{
		TemplateName:       c.String("template-name"),
		DefaultModality:    c.String("modality"),
		DefaultSeriesCount: c.Int("series-count"),
		DefaultImageCount:  c.Int("image-count"),
	}

	// Generate template
	template, err := templateGen.Generate(models, config)
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	// Prepare output path
	outputPath := c.String("output")
	if !strings.HasPrefix(outputPath, "/") && !strings.Contains(outputPath, ":") {
		// Relative path, prepend output directory
		outputDir := c.String("output-dir")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("failed to create output directory: %w", err)
		}
		outputPath = filepath.Join(outputDir, outputPath)
	}

	// Export template to YAML
	templateYAML, err := templateGen.ExportTemplate(template, "yaml")
	if err != nil {
		return fmt.Errorf("failed to export template: %w", err)
	}

	// Write template to file
	if err := os.WriteFile(outputPath, templateYAML, 0644); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	logrus.Infof("📋 Generated Template Summary:")
	logrus.Infof("• Template Name: %s", template.Name)
	logrus.Infof("• Modality: %s", template.Modality)
	logrus.Infof("• Series Count: %d", template.SeriesCount)
	logrus.Infof("• Image Count: %d", template.ImageCount)
	logrus.Infof("• Patient: %s (ID: %s)", template.PatientName, template.PatientID)
	logrus.Infof("• Accession: %s", template.AccessionNumber)
	logrus.Infof("• Study Description: %s", template.StudyDescription)

	logrus.Infof("🎯 Usage:")
	logrus.Infof("crgodicom create --template %s", template.Name)

	logrus.Infof("📁 Output File: %s", outputPath)

	return nil
}

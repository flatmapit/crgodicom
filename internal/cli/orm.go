package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/orm"
	"github.com/flatmapit/crgodicom/internal/orm/generator"
	"github.com/flatmapit/crgodicom/internal/orm/parser"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// CreateORMCommand creates the ORM template generation command
func CreateORMCommand() *cli.Command {
	return &cli.Command{
		Name:    "orm-generate",
		Aliases: []string{"orm"},
		Usage:   "Generate DICOM templates from ORM models or HL7 messages",
		Description: `Generate DICOM study templates from various input sources:
- HL7 ORM messages (.hl7, .txt files)
- Go struct definitions (.go files)
- SQL schema files (.sql files)
- JSON schema files (.json files)

The generated templates can be used with the 'create' command to generate
DICOM studies that match your existing data models.`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "input",
				Aliases:  []string{"i"},
				Usage:    "Input file path (HL7 message, Go structs, SQL schema, etc.)",
				Required: true,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output template file path",
				Value:   "generated-template.yaml",
			},
			&cli.StringFlag{
				Name:  "type",
				Aliases: []string{"t"},
				Usage: "Input type: hl7, go, sql, json (auto-detected if not specified)",
			},
			&cli.StringFlag{
				Name:  "template-name",
				Usage: "Name for the generated template",
				Value: "orm-generated-template",
			},
			&cli.StringFlag{
				Name:  "modality",
				Usage: "Default modality for the template",
				Value: "MR",
			},
			&cli.IntFlag{
				Name:  "series-count",
				Usage: "Default number of series",
				Value: 1,
			},
			&cli.IntFlag{
				Name:  "image-count",
				Usage: "Default number of images per series",
				Value: 10,
			},
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output",
			},
		},
		Action: generateORMTemplate,
	}
}

// generateORMTemplate handles the ORM template generation
func generateORMTemplate(c *cli.Context) error {
	inputFile := c.String("input")
	outputFile := c.String("output")
	inputType := c.String("type")
	templateName := c.String("template-name")
	modality := c.String("modality")
	seriesCount := c.Int("series-count")
	imageCount := c.Int("image-count")
	verbose := c.Bool("verbose")

	if verbose {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Infof("Generating DICOM template from ORM input: %s", inputFile)

	// Read input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read input file %s: %w", inputFile, err)
	}

	// Auto-detect input type if not specified
	if inputType == "" {
		inputType = detectInputType(inputFile)
		logrus.Debugf("Auto-detected input type: %s", inputType)
	}

	// Create ORM manager
	manager := orm.NewORMManager(nil)
	
	// Register parsers
	manager.RegisterParser("hl7", parser.NewHL7ORMParser())
	manager.RegisterParser("go", parser.NewGoStructParser())
	
	// Set generator
	manager.SetGenerator(generator.NewDICOMTemplateGenerator())

	// Create generation config
	config := orm.TemplateGenerationConfig{
		TemplateName:       templateName,
		DefaultModality:    modality,
		DefaultSeriesCount: seriesCount,
		DefaultImageCount:  imageCount,
		FieldMappings:      make(map[string]string),
		CustomTags:         make(map[string]map[string]string),
		Transformations:    make(map[string]string),
	}

	// Generate template
	template, err := manager.GenerateTemplate(inputData, inputType, config)
	if err != nil {
		return fmt.Errorf("failed to generate template: %w", err)
	}

	logrus.Debugf("Generated template with %d custom tag categories", len(template.CustomTags))

	// Export template
	templateData, err := manager.GetGenerator().ExportTemplate(template, "yaml")
	if err != nil {
		return fmt.Errorf("failed to export template: %w", err)
	}

	// Write output file
	if err := os.WriteFile(outputFile, templateData, 0644); err != nil {
		return fmt.Errorf("failed to write output file %s: %w", outputFile, err)
	}

	logrus.Infof("Successfully generated DICOM template: %s", outputFile)
	
	if verbose {
		fmt.Printf("\n📋 Generated Template Summary:\n")
		fmt.Printf("• Template Name: %s\n", template.Name)
		fmt.Printf("• Modality: %s\n", template.Modality)
		fmt.Printf("• Series Count: %d\n", template.SeriesCount)
		fmt.Printf("• Image Count: %d\n", template.ImageCount)
		fmt.Printf("• Custom Tag Categories: %d\n", len(template.CustomTags))
		
		for category, tags := range template.CustomTags {
			fmt.Printf("  - %s: %d tags\n", category, len(tags))
		}
		
		fmt.Printf("\n🎯 Usage:\n")
		fmt.Printf("crgodicom create --template %s\n", template.Name)
		fmt.Printf("\n📁 Output File: %s\n", outputFile)
	}

	return nil
}

// detectInputType attempts to detect the input type from file extension
func detectInputType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".hl7", ".txt":
		return "hl7"
	case ".go":
		return "go"
	case ".sql":
		return "sql"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	default:
		return "hl7" // Default to HL7 for unknown extensions
	}
}


package orm

import (
	"fmt"

	"github.com/flatmapit/crgodicom/internal/config"
)

// FieldType represents the data type of a model field
type FieldType string

const (
	FieldTypeString   FieldType = "string"
	FieldTypeInt      FieldType = "int"
	FieldTypeUint     FieldType = "uint"
	FieldTypeFloat    FieldType = "float"
	FieldTypeBool     FieldType = "bool"
	FieldTypeTime     FieldType = "time"
	FieldTypeBytes    FieldType = "bytes"
	FieldTypeStruct   FieldType = "struct"
	FieldTypeSlice    FieldType = "slice"
	FieldTypePointer  FieldType = "pointer"
	FieldTypeUnknown  FieldType = "unknown"
)

// DICOMTag represents a DICOM tag with its group and element
type DICOMTag struct {
	Group   uint16 `json:"group"`
	Element uint16 `json:"element"`
	VR      string `json:"vr,omitempty"`      // Value Representation
	Name    string `json:"name,omitempty"`    // Human-readable name
}

// String returns the DICOM tag in standard format (GGGG,EEEE)
func (d DICOMTag) String() string {
	return fmt.Sprintf("(%04X,%04X)", d.Group, d.Element)
}

// FieldMapping represents the mapping between an ORM field and DICOM tag
type FieldMapping struct {
	FieldName    string    `json:"field_name"`
	FieldType    FieldType `json:"field_type"`
	DICOMTag     DICOMTag  `json:"dicom_tag"`
	Transform    string    `json:"transform,omitempty"`    // Transformation function
	DefaultValue string    `json:"default_value,omitempty"` // Default value if field is empty
	Required     bool      `json:"required"`               // Whether field is required
	Validation   string    `json:"validation,omitempty"`   // Validation rules
}

// ModelDefinition represents a parsed ORM model
type ModelDefinition struct {
	Name        string         `json:"name"`
	Package     string         `json:"package,omitempty"`
	Fields      []FieldMapping `json:"fields"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Relationships []Relationship `json:"relationships,omitempty"`
}

// Relationship represents relationships between models
type Relationship struct {
	Type        string `json:"type"`         // "belongs_to", "has_one", "has_many"
	TargetModel string `json:"target_model"`
	ForeignKey  string `json:"foreign_key"`
	LocalKey    string `json:"local_key"`
}

// TemplateGenerationConfig holds configuration for template generation
type TemplateGenerationConfig struct {
	TemplateName     string            `json:"template_name"`
	DefaultModality  string            `json:"default_modality"`
	DefaultSeriesCount int             `json:"default_series_count"`
	DefaultImageCount  int             `json:"default_image_count"`
	FieldMappings    map[string]string `json:"field_mappings"`
	CustomTags       map[string]map[string]string `json:"custom_tags"`
	Transformations  map[string]string `json:"transformations"`
}

// GeneratedTemplate represents a generated DICOM template
type GeneratedTemplate struct {
	Name          string                 `yaml:"name"`
	Modality      string                 `yaml:"modality"`
	SeriesCount   int                    `yaml:"series_count"`
	ImageCount    int                    `yaml:"image_count"`
	AnatomicalRegion string              `yaml:"anatomical_region,omitempty"`
	StudyDescription string              `yaml:"study_description,omitempty"`
	PatientName   string                 `yaml:"patient_name,omitempty"`
	PatientID     string                 `yaml:"patient_id,omitempty"`
	AccessionNumber string               `yaml:"accession_number,omitempty"`
	CustomTags    map[string]map[string]string `yaml:"custom_tags,omitempty"`
	Metadata      map[string]interface{} `yaml:"metadata,omitempty"`
}

// Parser interface for different ORM input formats
type Parser interface {
	// Parse parses the input and returns model definitions
	Parse(input []byte) ([]ModelDefinition, error)
	
	// GetSupportedExtensions returns file extensions this parser supports
	GetSupportedExtensions() []string
	
	// GetParserType returns the type of parser (go, sql, json, etc.)
	GetParserType() string
}

// Generator interface for template generation
type Generator interface {
	// Generate creates a DICOM template from model definitions
	Generate(models []ModelDefinition, config TemplateGenerationConfig) (*GeneratedTemplate, error)
	
	// ValidateTemplate validates the generated template
	ValidateTemplate(template *GeneratedTemplate) error
	
	// ExportTemplate exports the template to various formats
	ExportTemplate(template *GeneratedTemplate, format string) ([]byte, error)
}

// ORMManager manages the ORM template generation process
type ORMManager struct {
	parsers    map[string]Parser
	generator  Generator
	config     *config.Config
}

// NewORMManager creates a new ORM manager
func NewORMManager(cfg *config.Config) *ORMManager {
	return &ORMManager{
		parsers:   make(map[string]Parser),
		config:    cfg,
	}
}

// RegisterParser registers a parser for a specific type
func (m *ORMManager) RegisterParser(parserType string, parser Parser) {
	m.parsers[parserType] = parser
}

// SetGenerator sets the template generator
func (m *ORMManager) SetGenerator(generator Generator) {
	m.generator = generator
}

// GenerateTemplate generates a DICOM template from input
func (m *ORMManager) GenerateTemplate(input []byte, inputType string, config TemplateGenerationConfig) (*GeneratedTemplate, error) {
	parser, exists := m.parsers[inputType]
	if !exists {
		return nil, fmt.Errorf("unsupported input type: %s", inputType)
	}
	
	models, err := parser.Parse(input)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}
	
	if m.generator == nil {
		return nil, fmt.Errorf("no generator configured")
	}
	
	template, err := m.generator.Generate(models, config)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template: %w", err)
	}
	
	if err := m.generator.ValidateTemplate(template); err != nil {
		return nil, fmt.Errorf("generated template validation failed: %w", err)
	}
	
	return template, nil
}

// GetSupportedTypes returns all supported input types
func (m *ORMManager) GetSupportedTypes() []string {
	types := make([]string, 0, len(m.parsers))
	for t := range m.parsers {
		types = append(types, t)
	}
	return types
}

// GetGenerator returns the configured generator
func (m *ORMManager) GetGenerator() Generator {
	return m.generator
}

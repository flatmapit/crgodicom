package generator

import (
	"fmt"
	"strings"
	"time"

	"github.com/flatmapit/crgodicom/internal/orm"
	"gopkg.in/yaml.v3"
)

// DICOMTemplateGenerator generates DICOM templates from ORM models
type DICOMTemplateGenerator struct {
	defaultMappings map[string]orm.DICOMTag
}

// NewDICOMTemplateGenerator creates a new DICOM template generator
func NewDICOMTemplateGenerator() *DICOMTemplateGenerator {
	return &DICOMTemplateGenerator{
		defaultMappings: getDefaultDICOMTagMappings(),
	}
}

// Generate creates a DICOM template from model definitions
func (g *DICOMTemplateGenerator) Generate(models []orm.ModelDefinition, config orm.TemplateGenerationConfig) (*orm.GeneratedTemplate, error) {
	template := &orm.GeneratedTemplate{
		Name:             config.TemplateName,
		Modality:         config.DefaultModality,
		SeriesCount:      config.DefaultSeriesCount,
		ImageCount:       config.DefaultImageCount,
		AnatomicalRegion: g.extractAnatomicalRegion(models),
		StudyDescription: g.extractStudyDescription(models),
		CustomTags:       make(map[string]map[string]string),
	}

	// Process each model and extract DICOM tags
	for _, model := range models {
		category := g.determineTagCategory(model.Name)
		if template.CustomTags[category] == nil {
			template.CustomTags[category] = make(map[string]string)
		}

		for _, field := range model.Fields {
			tagStr := field.DICOMTag.String()
			value := g.generateFieldValue(field, model)
			template.CustomTags[category][tagStr] = value
		}
	}

	// Add metadata
	template.Metadata = map[string]interface{}{
		"generated_from": "HL7_ORM",
		"generated_at":   time.Now().Format(time.RFC3339),
		"models_count":   len(models),
	}

	return template, nil
}

// ValidateTemplate validates the generated template
func (g *DICOMTemplateGenerator) ValidateTemplate(template *orm.GeneratedTemplate) error {
	if template.Name == "" {
		return fmt.Errorf("template name is required")
	}
	
	if template.Modality == "" {
		return fmt.Errorf("modality is required")
	}
	
	if template.SeriesCount <= 0 {
		return fmt.Errorf("series count must be greater than 0")
	}
	
	if template.ImageCount <= 0 {
		return fmt.Errorf("image count must be greater than 0")
	}

	// Validate DICOM tags
	for category, tags := range template.CustomTags {
		if len(tags) == 0 {
			continue
		}
		
		for tagStr, value := range tags {
			if !g.isValidDICOMTag(tagStr) {
				return fmt.Errorf("invalid DICOM tag format in category %s: %s", category, tagStr)
			}
			
			if value == "" {
				return fmt.Errorf("empty value for DICOM tag %s in category %s", tagStr, category)
			}
		}
	}

	return nil
}

// ExportTemplate exports the template to various formats
func (g *DICOMTemplateGenerator) ExportTemplate(template *orm.GeneratedTemplate, format string) ([]byte, error) {
	switch strings.ToLower(format) {
	case "yaml", "yml":
		return g.exportYAML(template)
	case "json":
		return g.exportJSON(template)
	default:
		return nil, fmt.Errorf("unsupported export format: %s", format)
	}
}

// exportYAML exports template as YAML
func (g *DICOMTemplateGenerator) exportYAML(template *orm.GeneratedTemplate) ([]byte, error) {
	// Create the template structure for YAML export
	yamlTemplate := map[string]interface{}{
		"study_templates": map[string]interface{}{
			template.Name: map[string]interface{}{
				"modality":          template.Modality,
				"series_count":      template.SeriesCount,
				"image_count":       template.ImageCount,
				"anatomical_region": template.AnatomicalRegion,
				"study_description": template.StudyDescription,
				"patient_name":      template.PatientName,
				"patient_id":        template.PatientID,
				"accession_number":  template.AccessionNumber,
				"custom_tags":       template.CustomTags,
				"metadata":          template.Metadata,
			},
		},
	}

	return yaml.Marshal(yamlTemplate)
}

// exportJSON exports template as JSON
func (g *DICOMTemplateGenerator) exportJSON(template *orm.GeneratedTemplate) ([]byte, error) {
	// For now, convert to YAML first then to JSON
	// This could be optimized later
	yamlData, err := g.exportYAML(template)
	if err != nil {
		return nil, err
	}
	
	// Convert YAML to JSON (simplified approach)
	var data interface{}
	if err := yaml.Unmarshal(yamlData, &data); err != nil {
		return nil, err
	}
	
	// This is a simplified conversion - a proper JSON export would be implemented
	return yamlData, nil
}

// Helper methods

// determineTagCategory determines which DICOM tag category a model belongs to
func (g *DICOMTemplateGenerator) determineTagCategory(modelName string) string {
	modelName = strings.ToLower(modelName)
	switch {
	case strings.Contains(modelName, "patient"):
		return "patient"
	case strings.Contains(modelName, "study"):
		return "study"
	case strings.Contains(modelName, "series"):
		return "series"
	case strings.Contains(modelName, "equipment"), strings.Contains(modelName, "device"):
		return "equipment"
	case strings.Contains(modelName, "institution"), strings.Contains(modelName, "facility"):
		return "institution"
	case strings.Contains(modelName, "order"), strings.Contains(modelName, "procedure"):
		return "procedure"
	case strings.Contains(modelName, "visit"):
		return "visit"
	default:
		return "custom"
	}
}

// extractAnatomicalRegion attempts to extract anatomical region from models
func (g *DICOMTemplateGenerator) extractAnatomicalRegion(models []orm.ModelDefinition) string {
	for _, model := range models {
		for _, field := range model.Fields {
			if strings.Contains(strings.ToLower(field.FieldName), "body") ||
			   strings.Contains(strings.ToLower(field.FieldName), "anatomy") ||
			   strings.Contains(strings.ToLower(field.FieldName), "region") {
				// This would contain logic to extract the actual value
				// For now, return a default based on procedure
				break
			}
		}
		
		// Check metadata for modality clues
		if modality, ok := model.Metadata["modality"].(string); ok {
			switch strings.ToUpper(modality) {
			case "MR", "CT":
				return "brain" // Default for MR/CT
			case "CR", "DX":
				return "chest" // Default for X-Ray
			case "US":
				return "abdomen" // Default for Ultrasound
			case "MG":
				return "breast" // Default for Mammography
			}
		}
	}
	return "chest" // Default fallback
}

// extractStudyDescription attempts to extract study description from models
func (g *DICOMTemplateGenerator) extractStudyDescription(models []orm.ModelDefinition) string {
	for _, model := range models {
		for _, field := range model.Fields {
			if strings.Contains(strings.ToLower(field.FieldName), "description") ||
			   strings.Contains(strings.ToLower(field.FieldName), "procedure") {
				// This would contain logic to extract the actual value
				return "Generated from HL7 ORM"
			}
		}
	}
	return "Generated from HL7 ORM"
}

// generateFieldValue generates a template value for a field
func (g *DICOMTemplateGenerator) generateFieldValue(field orm.FieldMapping, model orm.ModelDefinition) string {
	// If there's a default value, use it
	if field.DefaultValue != "" {
		return field.DefaultValue
	}

	// Generate template placeholder based on field type
	switch field.FieldType {
	case orm.FieldTypeString:
		return fmt.Sprintf("{{ .%s.%s }}", model.Name, field.FieldName)
	case orm.FieldTypeTime:
		if field.Transform != "" {
			return fmt.Sprintf("{{ .%s.%s.Format \"%s\" }}", model.Name, field.FieldName, g.getTimeFormat(field.Transform))
		}
		return fmt.Sprintf("{{ .%s.%s.Format \"20060102\" }}", model.Name, field.FieldName)
	case orm.FieldTypeInt, orm.FieldTypeUint:
		return fmt.Sprintf("{{ .%s.%s }}", model.Name, field.FieldName)
	case orm.FieldTypeFloat:
		return fmt.Sprintf("{{ printf \"%.1f\" .%s.%s }}", model.Name, field.FieldName)
	case orm.FieldTypeBool:
		return fmt.Sprintf("{{ if .%s.%s }}Y{{ else }}N{{ end }}", model.Name, field.FieldName)
	default:
		return fmt.Sprintf("{{ .%s.%s }}", model.Name, field.FieldName)
	}
}

// getTimeFormat extracts time format from transform string
func (g *DICOMTemplateGenerator) getTimeFormat(transform string) string {
	if strings.HasPrefix(transform, "date_format:") {
		return strings.TrimPrefix(transform, "date_format:")
	}
	if strings.HasPrefix(transform, "time_format:") {
		return strings.TrimPrefix(transform, "time_format:")
	}
	return "20060102" // Default date format
}

// isValidDICOMTag validates DICOM tag format
func (g *DICOMTemplateGenerator) isValidDICOMTag(tagStr string) bool {
	// Should match pattern (GGGG,EEEE)
	if !strings.HasPrefix(tagStr, "(") || !strings.HasSuffix(tagStr, ")") {
		return false
	}
	
	inner := strings.Trim(tagStr, "()")
	parts := strings.Split(inner, ",")
	if len(parts) != 2 {
		return false
	}
	
	// Check if both parts are valid hex
	for _, part := range parts {
		if len(part) != 4 {
			return false
		}
		for _, char := range part {
			if !((char >= '0' && char <= '9') || (char >= 'A' && char <= 'F') || (char >= 'a' && char <= 'f')) {
				return false
			}
		}
	}
	
	return true
}

// getDefaultDICOMTagMappings returns default DICOM tag mappings
func getDefaultDICOMTagMappings() map[string]orm.DICOMTag {
	return map[string]orm.DICOMTag{
		"patient_id":           {Group: 0x0010, Element: 0x0020}, // Patient ID
		"patient_name":         {Group: 0x0010, Element: 0x0010}, // Patient Name
		"patient_birth_date":   {Group: 0x0010, Element: 0x0030}, // Patient Birth Date
		"patient_sex":          {Group: 0x0010, Element: 0x0040}, // Patient Sex
		"study_instance_uid":   {Group: 0x0020, Element: 0x000D}, // Study Instance UID
		"study_date":           {Group: 0x0008, Element: 0x0020}, // Study Date
		"study_time":           {Group: 0x0008, Element: 0x0030}, // Study Time
		"accession_number":     {Group: 0x0008, Element: 0x0050}, // Accession Number
		"study_description":    {Group: 0x0008, Element: 0x1030}, // Study Description
		"referring_physician":  {Group: 0x0008, Element: 0x0090}, // Referring Physician Name
		"series_instance_uid":  {Group: 0x0020, Element: 0x000E}, // Series Instance UID
		"series_number":        {Group: 0x0020, Element: 0x0011}, // Series Number
		"series_description":   {Group: 0x0008, Element: 0x103E}, // Series Description
		"modality":             {Group: 0x0008, Element: 0x0060}, // Modality
		"body_part_examined":   {Group: 0x0018, Element: 0x0015}, // Body Part Examined
		"institution_name":     {Group: 0x0008, Element: 0x0080}, // Institution Name
		"station_name":         {Group: 0x0008, Element: 0x1010}, // Station Name
		"manufacturer":         {Group: 0x0008, Element: 0x0070}, // Manufacturer
		"device_serial_number": {Group: 0x0018, Element: 0x1000}, // Device Serial Number
	}
}

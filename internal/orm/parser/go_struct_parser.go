package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/flatmapit/crgodicom/internal/orm"
)

// GoStructParser parses Go struct definitions to extract DICOM mappings
type GoStructParser struct {
	fileSet *token.FileSet
}

// NewGoStructParser creates a new Go struct parser
func NewGoStructParser() *GoStructParser {
	return &GoStructParser{
		fileSet: token.NewFileSet(),
	}
}

// Parse parses Go source code and extracts model definitions
func (p *GoStructParser) Parse(input []byte) ([]orm.ModelDefinition, error) {
	// Parse the Go source code
	file, err := parser.ParseFile(p.fileSet, "", input, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Go source: %w", err)
	}

	var models []orm.ModelDefinition

	// Walk through the AST to find struct declarations
	ast.Inspect(file, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.TypeSpec:
			if structType, ok := node.Type.(*ast.StructType); ok {
				model, err := p.parseStruct(node.Name.Name, structType)
				if err != nil {
					// Log error but continue parsing other structs
					return true
				}
				models = append(models, model)
			}
		}
		return true
	})

	if len(models) == 0 {
		return nil, fmt.Errorf("no struct definitions found in input")
	}

	return models, nil
}

// parseStruct parses a single struct type and extracts field mappings
func (p *GoStructParser) parseStruct(name string, structType *ast.StructType) (orm.ModelDefinition, error) {
	model := orm.ModelDefinition{
		Name:     name,
		Package:  "", // Will be extracted from package declaration if needed
		Fields:   []orm.FieldMapping{},
		Metadata: make(map[string]interface{}),
	}

	for _, field := range structType.Fields.List {
		fieldMappings, err := p.parseField(field)
		if err != nil {
			// Skip fields that can't be parsed
			continue
		}
		model.Fields = append(model.Fields, fieldMappings...)
	}

	return model, nil
}

// parseField parses a single struct field and extracts DICOM mapping
func (p *GoStructParser) parseField(field *ast.Field) ([]orm.FieldMapping, error) {
	var mappings []orm.FieldMapping

	// Handle multiple field names (e.g., Name, Alias string)
	for _, name := range field.Names {
		mapping := orm.FieldMapping{
			FieldName: name.Name,
			FieldType: p.getFieldType(field.Type),
		}

		// Parse struct tags
		if field.Tag != nil {
			tagValue := field.Tag.Value
			// Remove backticks
			tagValue = strings.Trim(tagValue, "`")
			
			// Parse DICOM tag from struct tag
			if dicomTag := p.parseDICOMTag(tagValue); dicomTag != nil {
				mapping.DICOMTag = *dicomTag
			}

			// Parse other attributes from struct tags
			mapping.Required = p.parseRequired(tagValue)
			mapping.DefaultValue = p.parseDefaultValue(tagValue)
			mapping.Validation = p.parseValidation(tagValue)
			mapping.Transform = p.parseTransform(tagValue)
		}

		// Only include fields that have DICOM tags
		if mapping.DICOMTag.Group != 0 || mapping.DICOMTag.Element != 0 {
			mappings = append(mappings, mapping)
		}
	}

	return mappings, nil
}

// getFieldType determines the ORM field type from Go AST type
func (p *GoStructParser) getFieldType(expr ast.Expr) orm.FieldType {
	switch t := expr.(type) {
	case *ast.Ident:
		switch t.Name {
		case "string":
			return orm.FieldTypeString
		case "int", "int8", "int16", "int32", "int64":
			return orm.FieldTypeInt
		case "uint", "uint8", "uint16", "uint32", "uint64":
			return orm.FieldTypeUint
		case "float32", "float64":
			return orm.FieldTypeFloat
		case "bool":
			return orm.FieldTypeBool
		case "byte":
			return orm.FieldTypeBytes
		default:
			return orm.FieldTypeStruct
		}
	case *ast.SelectorExpr:
		// Handle qualified types like time.Time
		if ident, ok := t.X.(*ast.Ident); ok {
			if ident.Name == "time" && t.Sel.Name == "Time" {
				return orm.FieldTypeTime
			}
		}
		return orm.FieldTypeStruct
	case *ast.ArrayType:
		return orm.FieldTypeSlice
	case *ast.StarExpr:
		return orm.FieldTypePointer
	default:
		return orm.FieldTypeUnknown
	}
}

// parseDICOMTag extracts DICOM tag from struct tag
func (p *GoStructParser) parseDICOMTag(tagValue string) *orm.DICOMTag {
	// Look for dicom:"(GGGG,EEEE)" pattern
	parts := strings.Split(tagValue, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "dicom:") {
			// Extract the tag value
			tagStr := strings.TrimPrefix(part, "dicom:")
			tagStr = strings.Trim(tagStr, `"`)
			
			// Parse (GGGG,EEEE) format
			if strings.HasPrefix(tagStr, "(") && strings.HasSuffix(tagStr, ")") {
				tagStr = strings.Trim(tagStr, "()")
				parts := strings.Split(tagStr, ",")
				if len(parts) == 2 {
					group, err1 := strconv.ParseUint(parts[0], 16, 16)
					element, err2 := strconv.ParseUint(parts[1], 16, 16)
					if err1 == nil && err2 == nil {
						return &orm.DICOMTag{
							Group:   uint16(group),
							Element: uint16(element),
						}
					}
				}
			}
		}
	}
	return nil
}

// parseRequired checks if field is marked as required
func (p *GoStructParser) parseRequired(tagValue string) bool {
	return strings.Contains(tagValue, "required") || 
		   strings.Contains(tagValue, "not null") ||
		   strings.Contains(tagValue, "primaryKey")
}

// parseDefaultValue extracts default value from struct tag
func (p *GoStructParser) parseDefaultValue(tagValue string) string {
	parts := strings.Split(tagValue, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "default:") {
			return strings.TrimPrefix(part, "default:")
		}
	}
	return ""
}

// parseValidation extracts validation rules from struct tag
func (p *GoStructParser) parseValidation(tagValue string) string {
	parts := strings.Split(tagValue, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "validate:") {
			return strings.TrimPrefix(part, "validate:")
		}
	}
	return ""
}

// parseTransform extracts transformation function from struct tag
func (p *GoStructParser) parseTransform(tagValue string) string {
	parts := strings.Split(tagValue, " ")
	for _, part := range parts {
		if strings.HasPrefix(part, "transform:") {
			return strings.TrimPrefix(part, "transform:")
		}
	}
	return ""
}

// GetSupportedExtensions returns file extensions this parser supports
func (p *GoStructParser) GetSupportedExtensions() []string {
	return []string{".go"}
}

// GetParserType returns the type of parser
func (p *GoStructParser) GetParserType() string {
	return "go"
}

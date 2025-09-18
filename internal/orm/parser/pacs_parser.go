package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/flatmapit/crgodicom/internal/orm"
	"github.com/sirupsen/logrus"
)

// PACSStudy represents a study retrieved from PACS via CFIND
type PACSStudy struct {
	StudyInstanceUID               string `json:"study_instance_uid"`
	StudyDate                      string `json:"study_date"`
	StudyTime                      string `json:"study_time"`
	StudyDescription               string `json:"study_description"`
	AccessionNumber                string `json:"accession_number"`
	PatientName                    string `json:"patient_name"`
	PatientID                      string `json:"patient_id"`
	PatientBirthDate               string `json:"patient_birth_date"`
	PatientSex                     string `json:"patient_sex"`
	Modality                       string `json:"modality"`
	InstitutionName                string `json:"institution_name"`
	ReferringPhysician             string `json:"referring_physician"`
	StudyID                        string `json:"study_id"`
	NumberOfStudyRelatedInstances  string `json:"number_of_study_related_instances"`
	NumberOfSeriesRelatedInstances string `json:"number_of_series_related_instances"`
}

// PACSParser implements the Parser interface for PACS CFIND responses
type PACSParser struct {
	dcmtkPath string
}

// NewPACSParser creates a new PACS parser
func NewPACSParser(dcmtkPath string) *PACSParser {
	return &PACSParser{
		dcmtkPath: dcmtkPath,
	}
}

// Parse parses a PACS CFIND response file or performs CFIND query
func (p *PACSParser) Parse(input []byte) ([]orm.ModelDefinition, error) {
	inputStr := string(input)

	// Check if input is a Study Instance UID
	if p.isStudyUID(inputStr) {
		return p.queryPACSByStudyUID(inputStr)
	}

	// Parse existing CFIND response file
	return p.parseCFINDResponse(input)
}

// GetSupportedExtensions returns supported file extensions
func (p *PACSParser) GetSupportedExtensions() []string {
	return []string{".json", ".txt", ".hl7"}
}

// GetParserType returns the parser type
func (p *PACSParser) GetParserType() string {
	return "PACS_CFIND"
}

// isStudyUID checks if the input is a Study Instance UID
func (p *PACSParser) isStudyUID(input string) bool {
	// Study Instance UIDs typically start with "1.2" or "1.3" and contain dots
	studyUIDPattern := regexp.MustCompile(`^[0-9]+\.[0-9]+(\.[0-9]+)*$`)
	return studyUIDPattern.MatchString(input) && len(input) > 10
}

// queryPACSByStudyUID performs a CFIND query using Study Instance UID
func (p *PACSParser) queryPACSByStudyUID(studyUID string) ([]orm.ModelDefinition, error) {
	logrus.Infof("Performing CFIND query for Study Instance UID: %s", studyUID)

	// This would be called from the CLI command with PACS connection parameters
	// For now, return an error indicating PACS connection is required
	return nil, fmt.Errorf("PACS connection parameters required for Study UID queries - use the pacs-cfind command")
}

// parseCFINDResponse parses a CFIND response
func (p *PACSParser) parseCFINDResponse(data []byte) ([]orm.ModelDefinition, error) {
	logrus.Infof("Parsing CFIND response data")

	// Try to parse as JSON first
	var studies []PACSStudy
	if err := json.Unmarshal(data, &studies); err == nil {
		return p.convertPACSStudiesToModels(studies)
	}

	// Try to parse as DCMTK findscu output format
	studies, err := p.parseFindSCUOutput(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse CFIND response: %w", err)
	}

	return p.convertPACSStudiesToModels(studies)
}

// parseFindSCUOutput parses DCMTK findscu command output
func (p *PACSParser) parseFindSCUOutput(output string) ([]PACSStudy, error) {
	var studies []PACSStudy
	var currentStudy *PACSStudy

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Look for study start markers
		if strings.Contains(line, "Study Instance UID") {
			// Save previous study if exists
			if currentStudy != nil {
				studies = append(studies, *currentStudy)
			}

			// Start new study
			currentStudy = &PACSStudy{}
			p.parseStudyInstanceUID(line, currentStudy)
		} else if currentStudy != nil {
			// Parse other fields
			p.parseStudyField(line, currentStudy)
		}
	}

	// Add the last study
	if currentStudy != nil {
		studies = append(studies, *currentStudy)
	}

	return studies, nil
}

// parseStudyInstanceUID extracts Study Instance UID from line
func (p *PACSParser) parseStudyInstanceUID(line string, study *PACSStudy) {
	// Expected format: "Study Instance UID: 1.2.840.113619.2.5.1762583153.215519.978957063.78"
	parts := strings.Split(line, ":")
	if len(parts) >= 2 {
		study.StudyInstanceUID = strings.TrimSpace(parts[1])
	}
}

// parseStudyField parses individual study fields from DCMTK output
func (p *PACSParser) parseStudyField(line string, study *PACSStudy) {
	// Common DICOM tag patterns in findscu output
	fieldMappings := map[string]*string{
		"Study Date":                         &study.StudyDate,
		"Study Time":                         &study.StudyTime,
		"Study Description":                  &study.StudyDescription,
		"Accession Number":                   &study.AccessionNumber,
		"Patient's Name":                     &study.PatientName,
		"Patient ID":                         &study.PatientID,
		"Patient's Birth Date":               &study.PatientBirthDate,
		"Patient's Sex":                      &study.PatientSex,
		"Modality":                           &study.Modality,
		"Institution Name":                   &study.InstitutionName,
		"Referring Physician's Name":         &study.ReferringPhysician,
		"Study ID":                           &study.StudyID,
		"Number of Study Related Instances":  &study.NumberOfStudyRelatedInstances,
		"Number of Series Related Instances": &study.NumberOfSeriesRelatedInstances,
	}

	for fieldName, fieldPtr := range fieldMappings {
		if strings.Contains(line, fieldName) {
			parts := strings.Split(line, ":")
			if len(parts) >= 2 {
				*fieldPtr = strings.TrimSpace(parts[1])
			}
			break
		}
	}
}

// convertPACSStudiesToModels converts PACS studies to ORM models
func (p *PACSParser) convertPACSStudiesToModels(studies []PACSStudy) ([]orm.ModelDefinition, error) {
	var models []orm.ModelDefinition

	for _, study := range studies {
		// Create patient model
		patientModel := orm.ModelDefinition{
			Name:    "Patient",
			Package: "pacs",
			Fields: []orm.FieldMapping{
				{
					FieldName:    "PatientName",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0010,0010)"),
					DefaultValue: study.PatientName,
				},
				{
					FieldName:    "PatientID",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0010,0020)"),
					DefaultValue: study.PatientID,
				},
				{
					FieldName:    "PatientBirthDate",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0010,0030)"),
					DefaultValue: study.PatientBirthDate,
				},
				{
					FieldName:    "PatientSex",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0010,0040)"),
					DefaultValue: study.PatientSex,
				},
			},
			Metadata: map[string]interface{}{
				"source":    "PACS_CFIND",
				"study_uid": study.StudyInstanceUID,
			},
		}

		// Create study model
		studyModel := orm.ModelDefinition{
			Name:    "Study",
			Package: "pacs",
			Fields: []orm.FieldMapping{
				{
					FieldName:    "StudyInstanceUID",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0020,000D)"),
					DefaultValue: study.StudyInstanceUID,
				},
				{
					FieldName:    "StudyDate",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0020)"),
					DefaultValue: study.StudyDate,
				},
				{
					FieldName:    "StudyTime",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0030)"),
					DefaultValue: study.StudyTime,
				},
				{
					FieldName:    "StudyDescription",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,1030)"),
					DefaultValue: study.StudyDescription,
				},
				{
					FieldName:    "AccessionNumber",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0050)"),
					DefaultValue: study.AccessionNumber,
				},
				{
					FieldName:    "Modality",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0060)"),
					DefaultValue: study.Modality,
				},
				{
					FieldName:    "StudyID",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0020,0010)"),
					DefaultValue: study.StudyID,
				},
				{
					FieldName:    "ReferringPhysician",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0090)"),
					DefaultValue: study.ReferringPhysician,
				},
			},
			Metadata: map[string]interface{}{
				"source":         "PACS_CFIND",
				"study_uid":      study.StudyInstanceUID,
				"instance_count": study.NumberOfStudyRelatedInstances,
				"series_count":   study.NumberOfSeriesRelatedInstances,
			},
		}

		// Create institution model
		institutionModel := orm.ModelDefinition{
			Name:    "Institution",
			Package: "pacs",
			Fields: []orm.FieldMapping{
				{
					FieldName:    "InstitutionName",
					FieldType:    orm.FieldTypeString,
					DICOMTag:     p.parseDICOMTag("(0008,0080)"),
					DefaultValue: study.InstitutionName,
				},
			},
			Metadata: map[string]interface{}{
				"source":    "PACS_CFIND",
				"study_uid": study.StudyInstanceUID,
			},
		}

		models = append(models, patientModel, studyModel, institutionModel)
	}

	logrus.Infof("Converted %d PACS studies to %d ORM models", len(studies), len(models))
	return models, nil
}

// runFindSCU executes DCMTK findscu command
func (p *PACSParser) runFindSCU(host string, port int, aec, aet, studyUID string, verbose bool) (string, error) {
	findscuPath, err := p.getFindSCUPath()
	if err != nil {
		return "", fmt.Errorf("findscu not available: %w", err)
	}

	// Build findscu command
	args := []string{
		"-S", // Study Root Query/Retrieve Information Model
		"-aec", aec,
		"-aet", aet,
		"-k", fmt.Sprintf("StudyInstanceUID=%s", studyUID),
		"-k", "StudyDate",
		"-k", "StudyTime",
		"-k", "StudyDescription",
		"-k", "AccessionNumber",
		"-k", "PatientName",
		"-k", "PatientID",
		"-k", "PatientBirthDate",
		"-k", "PatientSex",
		"-k", "Modality",
		"-k", "InstitutionName",
		"-k", "ReferringPhysicianName",
		"-k", "StudyID",
		"-k", "NumberOfStudyRelatedInstances",
		"-k", "NumberOfSeriesRelatedInstances",
		fmt.Sprintf("%s", host),
		fmt.Sprintf("%d", port),
	}

	if verbose {
		args = append(args, "-v")
	}

	logrus.Debugf("Running findscu: %s %v", findscuPath, args)

	cmd := exec.Command(findscuPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("findscu command failed: %w, output: %s", err, string(output))
	}

	return string(output), nil
}

// getFindSCUPath gets the path to findscu executable
func (p *PACSParser) getFindSCUPath() (string, error) {
	if p.dcmtkPath != "" {
		findscuPath := filepath.Join(p.dcmtkPath, "findscu")
		if _, err := os.Stat(findscuPath); err == nil {
			return findscuPath, nil
		}
	}

	// Try system PATH
	if path, err := exec.LookPath("findscu"); err == nil {
		return path, nil
	}

	return "", fmt.Errorf("findscu not found in DCMTK installation or system PATH")
}

// parseDICOMTag parses a DICOM tag string like "(0010,0010)" into a DICOMTag struct
func (p *PACSParser) parseDICOMTag(tagStr string) orm.DICOMTag {
	// Remove parentheses and split by comma
	tagStr = strings.Trim(tagStr, "()")
	parts := strings.Split(tagStr, ",")

	if len(parts) != 2 {
		return orm.DICOMTag{}
	}

	group, err1 := strconv.ParseUint(parts[0], 16, 16)
	element, err2 := strconv.ParseUint(parts[1], 16, 16)

	if err1 != nil || err2 != nil {
		return orm.DICOMTag{}
	}

	return orm.DICOMTag{
		Group:   uint16(group),
		Element: uint16(element),
	}
}

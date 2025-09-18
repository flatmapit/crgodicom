package parser

import (
	"fmt"
	"strings"
	"time"

	"github.com/flatmapit/crgodicom/internal/orm"
)

// HL7ORMParser parses HL7 ORM (Order Management) messages to extract DICOM mappings
type HL7ORMParser struct {
	fieldSeparator        string
	componentSeparator    string
	repetitionSeparator   string
	escapeCharacter       string
	subComponentSeparator string
}

// NewHL7ORMParser creates a new HL7 ORM parser
func NewHL7ORMParser() *HL7ORMParser {
	return &HL7ORMParser{
		fieldSeparator:        "|",
		componentSeparator:    "^",
		repetitionSeparator:   "~",
		escapeCharacter:       "\\",
		subComponentSeparator: "&",
	}
}

// HL7Message represents a parsed HL7 message
type HL7Message struct {
	MSH HL7MSH   // Message Header
	PID HL7PID   // Patient Identification
	PV1 HL7PV1   // Patient Visit
	ORC HL7ORC   // Common Order
	OBR HL7OBR   // Observation Request
	OBX []HL7OBX // Observation/Result
}

// HL7MSH represents the Message Header segment
type HL7MSH struct {
	SendingApplication   string
	SendingFacility      string
	ReceivingApplication string
	ReceivingFacility    string
	DateTimeOfMessage    time.Time
	MessageType          string
	MessageControlID     string
	ProcessingID         string
	VersionID            string
}

// HL7PID represents the Patient Identification segment
type HL7PID struct {
	PatientID            []string
	PatientName          string
	DateOfBirth          time.Time
	Sex                  string
	PatientAddress       string
	CountryCode          string
	PhoneNumber          string
	PrimaryLanguage      string
	MaritalStatus        string
	Religion             string
	PatientAccountNumber string
}

// HL7PV1 represents the Patient Visit segment
type HL7PV1 struct {
	PatientClass            string
	AssignedPatientLocation string
	AttendingDoctor         string
	ReferringDoctor         string
	HospitalService         string
	AdmissionType           string
	FinancialClass          string
}

// HL7ORC represents the Common Order segment
type HL7ORC struct {
	OrderControl      string
	PlacerOrderNumber string
	FillerOrderNumber string
	PlacerGroupNumber string
	OrderStatus       string
	OrderingProvider  string
}

// HL7OBR represents the Observation Request segment
type HL7OBR struct {
	SetID               string
	PlacerOrderNumber   string
	FillerOrderNumber   string
	UniversalServiceID  string
	Priority            string
	RequestedDateTime   time.Time
	ObservationDateTime time.Time
	OrderingProvider    string
	ResultCopiesTo      string
	ReasonForStudy      string
	ClinicalHistory     string
	ProcedureCode       string
}

// HL7OBX represents the Observation/Result segment
type HL7OBX struct {
	SetID            string
	ValueType        string
	ObservationID    string
	ObservationSubID string
	ObservationValue string
	Units            string
	ResultStatus     string
}

// Parse parses HL7 ORM message and returns model definitions
func (p *HL7ORMParser) Parse(input []byte) ([]orm.ModelDefinition, error) {
	message, err := p.parseHL7Message(string(input))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HL7 message: %w", err)
	}

	// Generate model definitions from HL7 segments
	models := []orm.ModelDefinition{
		p.generatePatientModel(message.PID),
		p.generateStudyModel(message.OBR, message.MSH),
		p.generateOrderModel(message.ORC),
		p.generateVisitModel(message.PV1),
	}

	return models, nil
}

// parseHL7Message parses the raw HL7 message into structured data
func (p *HL7ORMParser) parseHL7Message(input string) (*HL7Message, error) {
	lines := strings.Split(input, "\n")
	message := &HL7Message{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		segments := strings.Split(line, p.fieldSeparator)
		if len(segments) < 2 {
			continue
		}

		segmentType := segments[0]
		switch segmentType {
		case "MSH":
			message.MSH = p.parseMSH(segments)
		case "PID":
			message.PID = p.parsePID(segments)
		case "PV1":
			message.PV1 = p.parsePV1(segments)
		case "ORC":
			message.ORC = p.parseORC(segments)
		case "OBR":
			message.OBR = p.parseOBR(segments)
		case "OBX":
			obx := p.parseOBX(segments)
			message.OBX = append(message.OBX, obx)
		}
	}

	return message, nil
}

// generatePatientModel creates a patient model from PID segment
func (p *HL7ORMParser) generatePatientModel(pid HL7PID) orm.ModelDefinition {
	return orm.ModelDefinition{
		Name:    "Patient",
		Package: "models",
		Fields: []orm.FieldMapping{
			{
				FieldName: "PatientID",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0010, Element: 0x0020},
				Required:  true,
			},
			{
				FieldName: "PatientName",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0010, Element: 0x0010},
				Required:  true,
			},
			{
				FieldName: "PatientBirthDate",
				FieldType: orm.FieldTypeTime,
				DICOMTag:  orm.DICOMTag{Group: 0x0010, Element: 0x0030},
				Transform: "date_format:20060102",
			},
			{
				FieldName:  "PatientSex",
				FieldType:  orm.FieldTypeString,
				DICOMTag:   orm.DICOMTag{Group: 0x0010, Element: 0x0040},
				Validation: "enum:M,F,O",
			},
			{
				FieldName: "PatientAddress",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0010, Element: 0x1040},
			},
		},
		Metadata: map[string]interface{}{
			"source_segment": "PID",
			"hl7_version":    "2.4",
		},
	}
}

// generateStudyModel creates a study model from OBR segment
func (p *HL7ORMParser) generateStudyModel(obr HL7OBR, msh HL7MSH) orm.ModelDefinition {
	return orm.ModelDefinition{
		Name:    "Study",
		Package: "models",
		Fields: []orm.FieldMapping{
			{
				FieldName: "StudyInstanceUID",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0020, Element: 0x000D},
				Required:  true,
			},
			{
				FieldName: "StudyDate",
				FieldType: orm.FieldTypeTime,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x0020},
				Transform: "date_format:20060102",
			},
			{
				FieldName: "StudyTime",
				FieldType: orm.FieldTypeTime,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x0030},
				Transform: "time_format:150405.000000",
			},
			{
				FieldName: "AccessionNumber",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x0050},
				Required:  true,
			},
			{
				FieldName: "StudyDescription",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x1030},
			},
			{
				FieldName: "ReferringPhysician",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x0090},
			},
			{
				FieldName: "ProcedureCode",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x1032},
			},
		},
		Metadata: map[string]interface{}{
			"source_segment": "OBR",
			"hl7_version":    "2.4",
			"modality":       "MR", // Extracted from procedure code
		},
	}
}

// generateOrderModel creates an order model from ORC segment
func (p *HL7ORMParser) generateOrderModel(orc HL7ORC) orm.ModelDefinition {
	return orm.ModelDefinition{
		Name:    "Order",
		Package: "models",
		Fields: []orm.FieldMapping{
			{
				FieldName: "OrderControl",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0040, Element: 0x1001}, // Requested Procedure ID
			},
			{
				FieldName: "PlacerOrderNumber",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0040, Element: 0x2016}, // Placer Order Number
			},
			{
				FieldName: "FillerOrderNumber",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0040, Element: 0x2017}, // Filler Order Number
			},
			{
				FieldName: "OrderingProvider",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x1050}, // Performing Physician Name
			},
		},
		Metadata: map[string]interface{}{
			"source_segment": "ORC",
		},
	}
}

// generateVisitModel creates a visit model from PV1 segment
func (p *HL7ORMParser) generateVisitModel(pv1 HL7PV1) orm.ModelDefinition {
	return orm.ModelDefinition{
		Name:    "Visit",
		Package: "models",
		Fields: []orm.FieldMapping{
			{
				FieldName: "PatientClass",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0038, Element: 0x0300}, // Current Patient Location
			},
			{
				FieldName: "AssignedPatientLocation",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x0080}, // Institution Name
			},
			{
				FieldName: "AttendingDoctor",
				FieldType: orm.FieldTypeString,
				DICOMTag:  orm.DICOMTag{Group: 0x0008, Element: 0x1048}, // Physician(s) of Record
			},
		},
		Metadata: map[string]interface{}{
			"source_segment": "PV1",
		},
	}
}

// Helper parsing methods for HL7 segments
func (p *HL7ORMParser) parseMSH(segments []string) HL7MSH {
	msh := HL7MSH{}
	if len(segments) > 2 {
		msh.SendingApplication = segments[2]
	}
	if len(segments) > 3 {
		msh.SendingFacility = segments[3]
	}
	if len(segments) > 4 {
		msh.ReceivingApplication = segments[4]
	}
	if len(segments) > 5 {
		msh.ReceivingFacility = segments[5]
	}
	if len(segments) > 6 {
		if dt, err := time.Parse("20060102150405-0700", segments[6]); err == nil {
			msh.DateTimeOfMessage = dt
		}
	}
	if len(segments) > 8 {
		msh.MessageType = segments[8]
	}
	if len(segments) > 9 {
		msh.MessageControlID = segments[9]
	}
	if len(segments) > 10 {
		msh.ProcessingID = segments[10]
	}
	if len(segments) > 11 {
		msh.VersionID = segments[11]
	}
	return msh
}

func (p *HL7ORMParser) parsePID(segments []string) HL7PID {
	pid := HL7PID{}
	if len(segments) > 3 {
		// Parse patient ID list
		idList := strings.Split(segments[3], p.repetitionSeparator)
		for _, id := range idList {
			if id != "" {
				pid.PatientID = append(pid.PatientID, strings.Split(id, p.componentSeparator)[0])
			}
		}
	}
	if len(segments) > 5 {
		// Parse patient name (LAST^FIRST^MIDDLE^SUFFIX^PREFIX)
		nameComponents := strings.Split(segments[5], p.componentSeparator)
		if len(nameComponents) >= 2 {
			pid.PatientName = nameComponents[0] + "^" + nameComponents[1]
			if len(nameComponents) > 2 && nameComponents[2] != "" {
				pid.PatientName += "^" + nameComponents[2]
			}
		}
	}
	if len(segments) > 7 {
		if dob, err := time.Parse("20060102", segments[7]); err == nil {
			pid.DateOfBirth = dob
		}
	}
	if len(segments) > 8 {
		pid.Sex = segments[8]
	}
	if len(segments) > 11 {
		pid.PatientAddress = segments[11]
	}
	return pid
}

func (p *HL7ORMParser) parsePV1(segments []string) HL7PV1 {
	pv1 := HL7PV1{}
	if len(segments) > 2 {
		pv1.PatientClass = segments[2]
	}
	if len(segments) > 3 {
		pv1.AssignedPatientLocation = segments[3]
	}
	if len(segments) > 7 {
		pv1.AttendingDoctor = segments[7]
	}
	if len(segments) > 8 {
		pv1.ReferringDoctor = segments[8]
	}
	if len(segments) > 10 {
		pv1.HospitalService = segments[10]
	}
	if len(segments) > 18 {
		pv1.AdmissionType = segments[18]
	}
	if len(segments) > 20 {
		pv1.FinancialClass = segments[20]
	}
	return pv1
}

func (p *HL7ORMParser) parseORC(segments []string) HL7ORC {
	orc := HL7ORC{}
	if len(segments) > 1 {
		orc.OrderControl = segments[1]
	}
	if len(segments) > 2 {
		orc.PlacerOrderNumber = segments[2]
	}
	if len(segments) > 3 {
		orc.FillerOrderNumber = segments[3]
	}
	if len(segments) > 4 {
		orc.PlacerGroupNumber = segments[4]
	}
	if len(segments) > 5 {
		orc.OrderStatus = segments[5]
	}
	if len(segments) > 12 {
		orc.OrderingProvider = segments[12]
	}
	return orc
}

func (p *HL7ORMParser) parseOBR(segments []string) HL7OBR {
	obr := HL7OBR{}
	if len(segments) > 1 {
		obr.SetID = segments[1]
	}
	if len(segments) > 2 {
		obr.PlacerOrderNumber = segments[2]
	}
	if len(segments) > 3 {
		obr.FillerOrderNumber = segments[3]
	}
	if len(segments) > 4 {
		// Parse Universal Service ID (procedure code)
		serviceComponents := strings.Split(segments[4], p.componentSeparator)
		if len(serviceComponents) > 0 {
			obr.UniversalServiceID = serviceComponents[0]
		}
		if len(serviceComponents) > 1 {
			obr.ProcedureCode = serviceComponents[1]
		}
	}
	if len(segments) > 6 {
		obr.Priority = segments[6]
	}
	if len(segments) > 16 {
		obr.OrderingProvider = segments[16]
	}
	if len(segments) > 31 {
		obr.ReasonForStudy = segments[31]
	}
	return obr
}

func (p *HL7ORMParser) parseOBX(segments []string) HL7OBX {
	obx := HL7OBX{}
	if len(segments) > 1 {
		obx.SetID = segments[1]
	}
	if len(segments) > 2 {
		obx.ValueType = segments[2]
	}
	if len(segments) > 3 {
		obx.ObservationID = segments[3]
	}
	if len(segments) > 4 {
		obx.ObservationSubID = segments[4]
	}
	if len(segments) > 5 {
		obx.ObservationValue = segments[5]
	}
	if len(segments) > 6 {
		obx.Units = segments[6]
	}
	if len(segments) > 11 {
		obx.ResultStatus = segments[11]
	}
	return obx
}

// GetSupportedExtensions returns file extensions this parser supports
func (p *HL7ORMParser) GetSupportedExtensions() []string {
	return []string{".hl7", ".txt"}
}

// GetParserType returns the type of parser
func (p *HL7ORMParser) GetParserType() string {
	return "hl7"
}

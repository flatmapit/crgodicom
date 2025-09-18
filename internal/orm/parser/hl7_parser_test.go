package parser

import (
	"testing"
	"time"

	"github.com/flatmapit/crgodicom/internal/orm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHL7ORMParser_Parse(t *testing.T) {
	parser := NewHL7ORMParser()
	
	// Test HL7 message (simplified version of the TESTRIS example)
	hl7Message := `MSH|^~\&|TESTRIS|TESTRIS|EXT_DEF||20241215143022+1100||ORM^O01^|abc123xyz-def|P|2.4
PID|||200000001^^^2006630&&GUID^005||JOHNSON^SARAH^MARIE^""^MS||19820623|F|||MAPLE STREET 789^""^MELBOURNE^VIC^3001^1201^home^""
PV1||I|CNR^^^E304^^^^^Melbourne General Hospital|||||789012CD^ATTENDINGDR||||||||||||MC
ORC|XO|4339239594|2024WS0000001|2024WS0000001|E|||||||789012CD^ATTENDINGDR^^^^Dr
OBR|1|4444444444|2024WS0000001-1|MRIBRAINCON^MRI Brain with Contrast^WS-MGH.ORDERABLES|||||||||Research Acc: Y||^^^Neurological, Brain, Head, Neck|||2024WS0000001-1||||MR|||^^^20241215140500+1100^^Routine||||^Clinical History: Progressive headaches||SRV-VICG-EXT-DEF@vichealth.net||||||||||||MRIBRAINCON^MRI Brain with Contrast^WS-MGH.PROCEDURES`

	models, err := parser.Parse([]byte(hl7Message))
	require.NoError(t, err)
	require.Len(t, models, 4, "Expected 4 models (Patient, Study, Order, Visit)")

	// Test Patient model
	patientModel := findModelByName(models, "Patient")
	require.NotNil(t, patientModel, "Patient model not found")
	assert.Equal(t, "Patient", patientModel.Name)
	assert.Equal(t, "models", patientModel.Package)
	assert.Len(t, patientModel.Fields, 5, "Expected 5 patient fields")

	// Test Study model
	studyModel := findModelByName(models, "Study")
	require.NotNil(t, studyModel, "Study model not found")
	assert.Equal(t, "Study", studyModel.Name)
	assert.Len(t, studyModel.Fields, 7, "Expected 7 study fields")

	// Test Order model
	orderModel := findModelByName(models, "Order")
	require.NotNil(t, orderModel, "Order model not found")
	assert.Equal(t, "Order", orderModel.Name)
	assert.Len(t, orderModel.Fields, 4, "Expected 4 order fields")

	// Test Visit model
	visitModel := findModelByName(models, "Visit")
	require.NotNil(t, visitModel, "Visit model not found")
	assert.Equal(t, "Visit", visitModel.Name)
	assert.Len(t, visitModel.Fields, 3, "Expected 3 visit fields")
}

func TestHL7ORMParser_ParseMSH(t *testing.T) {
	parser := NewHL7ORMParser()
	
	segments := []string{
		"MSH", "^~\\&", "TESTRIS", "TESTRIS", "EXT_DEF", "", 
		"20241215143022+1100", "", "ORM^O01^", "abc123xyz-def", "P", "2.4",
	}
	
	msh := parser.parseMSH(segments)
	
	assert.Equal(t, "TESTRIS", msh.SendingApplication)
	assert.Equal(t, "TESTRIS", msh.SendingFacility)
	assert.Equal(t, "EXT_DEF", msh.ReceivingApplication)
	assert.Equal(t, "ORM^O01^", msh.MessageType)
	assert.Equal(t, "abc123xyz-def", msh.MessageControlID)
	assert.Equal(t, "P", msh.ProcessingID)
	assert.Equal(t, "2.4", msh.VersionID)
}

func TestHL7ORMParser_ParsePID(t *testing.T) {
	parser := NewHL7ORMParser()
	
	segments := []string{
		"PID", "", "", "200000001^^^2006630&&GUID^005", "", 
		"JOHNSON^SARAH^MARIE^\"\"^MS", "", "19820623", "F", "", "",
		"MAPLE STREET 789^\"\"^MELBOURNE^VIC^3001^1201^home^\"\"",
	}
	
	pid := parser.parsePID(segments)
	
	assert.Contains(t, pid.PatientID, "200000001")
	assert.Equal(t, "JOHNSON^SARAH", pid.PatientName)
	assert.Equal(t, "F", pid.Sex)
	assert.Equal(t, "MAPLE STREET 789^\"\"^MELBOURNE^VIC^3001^1201^home^\"\"", pid.PatientAddress)
	
	// Test date parsing
	expectedDate := time.Date(1982, 6, 23, 0, 0, 0, 0, time.UTC)
	assert.Equal(t, expectedDate, pid.DateOfBirth)
}

func TestHL7ORMParser_ParseOBR(t *testing.T) {
	parser := NewHL7ORMParser()
	
	segments := []string{
		"OBR", "1", "4444444444", "2024WS0000001-1", 
		"MRIBRAINCON^MRI Brain with Contrast^WS-MGH.ORDERABLES",
		"", "", "", "", "", "", "", "", "", "", "", "",
		"789012CD^ATTENDINGDR", "", "", "", "", "", "", "", "MR",
	}
	
	obr := parser.parseOBR(segments)
	
	assert.Equal(t, "1", obr.SetID)
	assert.Equal(t, "4444444444", obr.PlacerOrderNumber)
	assert.Equal(t, "2024WS0000001-1", obr.FillerOrderNumber)
	assert.Equal(t, "MRIBRAINCON", obr.UniversalServiceID)
	assert.Equal(t, "MRI Brain with Contrast", obr.ProcedureCode)
	assert.Equal(t, "789012CD^ATTENDINGDR", obr.OrderingProvider)
}

func TestHL7ORMParser_GetSupportedExtensions(t *testing.T) {
	parser := NewHL7ORMParser()
	
	extensions := parser.GetSupportedExtensions()
	expectedExtensions := []string{".hl7", ".txt"}
	
	assert.ElementsMatch(t, expectedExtensions, extensions)
}

func TestHL7ORMParser_GetParserType(t *testing.T) {
	parser := NewHL7ORMParser()
	
	parserType := parser.GetParserType()
	assert.Equal(t, "hl7", parserType)
}

func TestHL7ORMParser_ParseInvalidMessage(t *testing.T) {
	parser := NewHL7ORMParser()
	
	// Test with invalid HL7 message
	invalidMessage := "This is not a valid HL7 message"
	
	models, err := parser.Parse([]byte(invalidMessage))
	
	// Should not error but should return models with empty/default values
	assert.NoError(t, err)
	assert.Len(t, models, 4) // Should still return 4 models with default values
}

func TestHL7ORMParser_ParseEmptyMessage(t *testing.T) {
	parser := NewHL7ORMParser()
	
	// Test with empty message
	models, err := parser.Parse([]byte(""))
	
	assert.NoError(t, err)
	assert.Len(t, models, 4) // Should return 4 models with default values
}

// Helper function to find model by name
func findModelByName(models []orm.ModelDefinition, name string) *orm.ModelDefinition {
	for _, model := range models {
		if model.Name == name {
			return &model
		}
	}
	return nil
}

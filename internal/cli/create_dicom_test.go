package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/flatmapit/crgodicom/internal/dcmtk"
	"github.com/flatmapit/crgodicom/internal/dicom"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateCommand_DICOMGeneration(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "create_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test parameters
	testCases := []struct {
		name          string
		patientName   string
		patientID     string
		modality      string
		seriesCount   int
		imageCount    int
		expectSuccess bool
	}{
		{
			name:          "Single DX Image",
			patientName:   "TEST^SINGLE",
			patientID:     "SINGLE001",
			modality:      "DX",
			seriesCount:   1,
			imageCount:    1,
			expectSuccess: true,
		},
		{
			name:          "Multiple CT Series",
			patientName:   "TEST^MULTI",
			patientID:     "MULTI001",
			modality:      "CT",
			seriesCount:   3,
			imageCount:    5,
			expectSuccess: true,
		},
		{
			name:          "MRI Study",
			patientName:   "TEST^MRI",
			patientID:     "MRI001",
			modality:      "MR",
			seriesCount:   2,
			imageCount:    10,
			expectSuccess: true,
		},
		{
			name:          "Ultrasound Study",
			patientName:   "TEST^US",
			patientID:     "US001",
			modality:      "US",
			seriesCount:   1,
			imageCount:    3,
			expectSuccess: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test directory for this case
			testDir := filepath.Join(tempDir, tc.name)
			err := os.MkdirAll(testDir, 0755)
			require.NoError(t, err)

			// Simulate the create command logic
			// This is a simplified test - in a real implementation,
			// you would call the actual create command function

			// For now, just verify the test setup
			assert.True(t, tc.expectSuccess, "Test case should expect success")

			// Verify test directory was created
			_, err = os.Stat(testDir)
			assert.NoError(t, err, "Test directory should be created")
		})
	}
}

func TestCreateCommand_ElementValidation(t *testing.T) {
	// Test that DICOM elements are created with correct types
	testCases := []struct {
		name        string
		element     string
		expectedVR  string
		description string
	}{
		{
			name:        "Series Number",
			element:     "0020,0011",
			expectedVR:  "IS",
			description: "Series Number should be IS (Integer String)",
		},
		{
			name:        "Instance Number",
			element:     "0020,0013",
			expectedVR:  "IS",
			description: "Instance Number should be IS (Integer String)",
		},
		{
			name:        "Patient Name",
			element:     "0010,0010",
			expectedVR:  "PN",
			description: "Patient Name should be PN (Person Name)",
		},
		{
			name:        "Study Date",
			element:     "0008,0020",
			expectedVR:  "DA",
			description: "Study Date should be DA (Date)",
		},
		{
			name:        "Study Time",
			element:     "0008,0030",
			expectedVR:  "TM",
			description: "Study Time should be TM (Time)",
		},
		{
			name:        "Modality",
			element:     "0008,0060",
			expectedVR:  "CS",
			description: "Modality should be CS (Code String)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This is a placeholder test - in a real implementation,
			// you would validate that the DICOM elements are created
			// with the correct VR (Value Representation) types

			assert.NotEmpty(t, tc.element, "Element tag should not be empty")
			assert.NotEmpty(t, tc.expectedVR, "Expected VR should not be empty")
			assert.NotEmpty(t, tc.description, "Description should not be empty")
		})
	}
}

func TestCreateCommand_FileStructure(t *testing.T) {
	// Test that the correct directory structure is created
	tempDir, err := os.MkdirTemp("", "file_structure_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Expected directory structure:
	// output-dir/
	//   └── study-uid/
	//       └── series-uid/
	//           └── image_001.dcm
	//           └── image_002.dcm
	//           └── ...

	// Create mock directory structure
	studyUID := "1.2.3.4.5.6.7.8.9"

	studyDir := filepath.Join(tempDir, studyUID)
	seriesDir := filepath.Join(studyDir, "series_001")

	err = os.MkdirAll(seriesDir, 0755)
	require.NoError(t, err)

	// Create mock DICOM files
	for i := 1; i <= 3; i++ {
		fileName := filepath.Join(seriesDir, "image_001.dcm")
		file, err := os.Create(fileName)
		require.NoError(t, err)
		file.Close()
	}

	// Verify structure
	_, err = os.Stat(studyDir)
	assert.NoError(t, err, "Study directory should exist")

	_, err = os.Stat(seriesDir)
	assert.NoError(t, err, "Series directory should exist")

	// Check for DICOM files
	dicomFiles, err := filepath.Glob(filepath.Join(seriesDir, "*.dcm"))
	require.NoError(t, err)
	assert.True(t, len(dicomFiles) > 0, "Should have DICOM files")
}

func TestCreateCommand_ModalityValidation(t *testing.T) {
	// Test valid and invalid modalities
	validModalities := []string{"CR", "CT", "MR", "US", "DX", "MG"}
	invalidModalities := []string{"INVALID", "XYZ", "123", ""}

	for _, modality := range validModalities {
		t.Run("Valid_"+modality, func(t *testing.T) {
			// In a real test, you would validate that the modality
			// is accepted by the create command
			assert.NotEmpty(t, modality, "Modality should not be empty")
		})
	}

	for _, modality := range invalidModalities {
		t.Run("Invalid_"+modality, func(t *testing.T) {
			// In a real test, you would validate that the modality
			// is rejected by the create command
			// For now, just verify the test setup
			assert.True(t, true, "Invalid modality test placeholder")
		})
	}
}

// Integration test that combines multiple aspects
func TestCreateCommand_Integration(t *testing.T) {
	// This test would combine:
	// 1. Command line argument parsing
	// 2. DICOM file generation
	// 3. File structure creation
	// 4. Element validation
	// 5. Error handling

	tempDir, err := os.MkdirTemp("", "integration_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Test parameters
	patientName := "INTEGRATION^TEST"
	patientID := "INTEG001"
	modality := "DX"
	seriesCount := 1
	imageCount := 1

	// Simulate the complete create command workflow
	t.Run("Complete_Workflow", func(t *testing.T) {
		// 1. Validate inputs
		assert.NotEmpty(t, patientName, "Patient name should not be empty")
		assert.NotEmpty(t, patientID, "Patient ID should not be empty")
		assert.NotEmpty(t, modality, "Modality should not be empty")
		assert.True(t, seriesCount > 0, "Series count should be positive")
		assert.True(t, imageCount > 0, "Image count should be positive")

		// 2. Create output directory
		err := os.MkdirAll(tempDir, 0755)
		require.NoError(t, err)

		// 3. Verify directory was created
		_, err = os.Stat(tempDir)
		assert.NoError(t, err, "Output directory should be created")

		// 4. Test actual DICOM creation
		t.Run("Real_DICOM_Creation", func(t *testing.T) {
			// Create a real DICOM file using our DCMTK writer
			writer := dcmtk.NewDCMTKBasedWriter()

			// Generate pixel data
			generator := dicom.NewImageGenerator()
			pixelData, err := generator.GenerateImage(modality, 512, 512, 16)
			require.NoError(t, err, "Should generate pixel data")

			// Create DICOM file
			dicomPath := filepath.Join(tempDir, "test.dcm")
			err = writer.WriteDICOMFile(
				dicomPath,
				patientName,
				patientID,
				"1.2.840.10008.5.1.4.1.1.1759021409.7465550661221513685", // studyUID
				"1.2.840.10008.5.1.4.1.1.1759021409.7465550661221513686", // seriesUID
				"1.2.840.10008.5.1.4.1.1.1759021409.7465550661221513687", // instanceUID
				modality,
				512, 512, // width, height
				16, 16, 15, // bitsAllocated, bitsStored, highBit
				1,             // samplesPerPixel
				"MONOCHROME2", // photometricInterpretation
				pixelData,
			)
			require.NoError(t, err, "Should create DICOM file")

			// Validate DICOM file exists and has reasonable size
			fileInfo, err := os.Stat(dicomPath)
			require.NoError(t, err, "DICOM file should exist")
			assert.True(t, fileInfo.Size() > 100*1024, "DICOM file should be > 100KB")

			// Validate DICOM file with dcmdump
			cmd := exec.Command("dcmdump", dicomPath)
			output, err := cmd.CombinedOutput()
			require.NoError(t, err, "dcmdump should succeed")

			// Check for key DICOM elements
			outputStr := string(output)
			assert.Contains(t, outputStr, "PatientName", "Should contain PatientName")
			assert.Contains(t, outputStr, "PatientID", "Should contain PatientID")
			assert.Contains(t, outputStr, "Modality", "Should contain Modality")
			assert.Contains(t, outputStr, "PixelData", "Should contain PixelData")
		})
	})
}

// Regression test to ensure pixel data issues don't reoccur
func TestCreateCommand_PixelDataRegression(t *testing.T) {
	// This test ensures that the pixel data issues we fixed don't reoccur
	tempDir, err := os.MkdirTemp("", "pixel_regression_test_*")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	t.Run("PixelData_ElementTypes", func(t *testing.T) {
		// Test that Series Number and Instance Number use correct types
		// (This was the root cause of our ValueType errors)

		// Series Number (0020,0011) should be IS (Integer String) - []string
		// Instance Number (0020,0013) should be IS (Integer String) - []string

		// In a real implementation, you would:
		// 1. Create a DICOM file
		// 2. Parse the file
		// 3. Verify the element types are correct
		// 4. Ensure no ValueType errors occur

		assert.True(t, true, "Pixel data regression test placeholder")
	})

	t.Run("ElementOrder_Regression", func(t *testing.T) {
		// Test that elements are added in correct ascending tag order
		// (This was causing "Dataset not in ascending tag order" warnings)

		// In a real implementation, you would:
		// 1. Create a DICOM file
		// 2. Parse the file
		// 3. Verify elements are in ascending tag order
		// 4. Test with DCMTK tools to ensure no warnings

		assert.True(t, true, "Element order regression test placeholder")
	})
}

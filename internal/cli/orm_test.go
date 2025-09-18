package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCreateORMCommand(t *testing.T) {
	cmd := CreateORMCommand()
	
	// Test command properties
	assert.Equal(t, "orm-generate", cmd.Name)
	assert.Contains(t, cmd.Aliases, "orm")
	assert.NotEmpty(t, cmd.Usage)
	assert.NotEmpty(t, cmd.Description)
}

func TestORMCommandFlags(t *testing.T) {
	cmd := CreateORMCommand()
	
	// Test that all expected flags are present
	expectedFlags := []string{
		"input", "output", "type", "template-name", "modality",
		"series-count", "image-count", "verbose",
	}
	
	for _, flagName := range expectedFlags {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == flagName {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected flag %s not found", flagName)
	}
	
	// Test required flags
	requiredFlags := []string{"input"}
	for _, flagName := range requiredFlags {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == flagName {
				switch f := flag.(type) {
				case *cli.StringFlag:
					assert.True(t, f.Required, "Flag %s should be required", flagName)
				}
				found = true
				break
			}
		}
		assert.True(t, found, "Required flag %s not found", flagName)
	}
}

func TestDetectInputType(t *testing.T) {
	tests := []struct {
		filename     string
		expectedType string
	}{
		{"message.hl7", "hl7"},
		{"data.txt", "hl7"},
		{"models.go", "go"},
		{"schema.sql", "sql"},
		{"api.json", "json"},
		{"config.yaml", "yaml"},
		{"config.yml", "yaml"},
		{"unknown.xyz", "hl7"}, // Default fallback
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := detectInputType(tt.filename)
			assert.Equal(t, tt.expectedType, result)
		})
	}
}

func TestORMCommandValidation(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(string) error
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing input file",
			args:    []string{"orm-generate"},
			wantErr: true,
			errMsg:  "Required flag \"input\" not set",
		},
		{
			name:    "non-existent input file",
			args:    []string{"orm-generate", "--input", "/non/existent/file.hl7"},
			wantErr: true,
			errMsg:  "failed to read input file",
		},
		{
			name: "valid HL7 input",
			args: []string{"orm-generate", "--input", "test.hl7", "--output", "output.yaml"},
			setup: func(tempDir string) error {
				// Create test HL7 file
				hl7Content := `MSH|^~\&|TESTRIS|TESTRIS|EXT_DEF||20241215143022+1100||ORM^O01^|abc123xyz-def|P|2.4
PID|||200000001^^^2006630&&GUID^005||JOHNSON^SARAH^MARIE||19820623|F
OBR|1|4444444444|2024WS0000001-1|MRIBRAINCON^MRI Brain with Contrast||||||||||||||||MR`
				return os.WriteFile(filepath.Join(tempDir, "test.hl7"), []byte(hl7Content), 0644)
			},
			wantErr: false,
		},
		{
			name: "valid Go input",
			args: []string{"orm-generate", "--input", "test.go", "--type", "go", "--output", "output.yaml"},
			setup: func(tempDir string) error {
				// Create test Go file
				goContent := `package models

type Patient struct {
	ID   uint   ` + "`" + `gorm:"primaryKey" dicom:"(0010,0020)"` + "`" + `
	Name string ` + "`" + `dicom:"(0010,0010)"` + "`" + `
}`
				return os.WriteFile(filepath.Join(tempDir, "test.go"), []byte(goContent), 0644)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()
			
			// Run setup if provided
			if tt.setup != nil {
			err := tt.setup(tempDir)
			assert.NoError(t, err, "Setup failed")
			}
			
			// Create CLI app with ORM command
			app := &cli.App{
				Name: "crgodicom-test",
				Commands: []*cli.Command{
					CreateORMCommand(),
				},
			}
			
			// Update args to use temp directory paths
			args := append([]string{"crgodicom-test"}, tt.args...)
			for i, arg := range args {
				if arg == "--input" && i+1 < len(args) {
					args[i+1] = filepath.Join(tempDir, filepath.Base(args[i+1]))
				}
				if arg == "--output" && i+1 < len(args) {
					args[i+1] = filepath.Join(tempDir, args[i+1])
				}
			}
			
			err := app.Run(args)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				
				// Check if output file was created
				outputFile := filepath.Join(tempDir, "output.yaml")
				if _, err := os.Stat(outputFile); err == nil {
					// File was created, verify it's not empty
					content, err := os.ReadFile(outputFile)
					assert.NoError(t, err)
					assert.NotEmpty(t, content, "Output file should not be empty")
				}
			}
		})
	}
}

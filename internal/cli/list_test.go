package cli

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestListCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		setup   func(string) error
		wantErr bool
		errMsg  string
	}{
		{
			name:    "empty directory",
			args:    []string{"list"},
			wantErr: false, // Should handle empty directory gracefully
		},
		{
			name:    "non-existent directory",
			args:    []string{"list", "--output-dir", "/non/existent/path"},
			wantErr: false, // Should handle gracefully with warning
		},
		{
			name:    "invalid format",
			args:    []string{"list", "--format", "invalid"},
			wantErr: true,
			errMsg:  "invalid format",
		},
		{
			name:    "valid json format",
			args:    []string{"list", "--format", "json"},
			wantErr: false,
		},
		{
			name:    "valid csv format",
			args:    []string{"list", "--format", "csv"},
			wantErr: false,
		},
		{
			name:    "verbose output",
			args:    []string{"list", "--verbose"},
			wantErr: false,
		},
		{
			name: "directory with studies",
			args: []string{"list"},
			setup: func(tempDir string) error {
				// Create mock study directory structure
				studyDir := filepath.Join(tempDir, "1.2.840.10008.5.1.4.1.1.123456789")
				seriesDir := filepath.Join(studyDir, "series_001")
				if err := os.MkdirAll(seriesDir, 0755); err != nil {
					return err
				}
				
				// Create mock DICOM file
				mockDicomFile := filepath.Join(seriesDir, "image_001.dcm")
				return os.WriteFile(mockDicomFile, []byte("mock dicom data"), 0644)
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
			
			// Create test config
			cfg := config.DefaultConfig()
			cfg.Storage.BaseDir = tempDir
			
			// Create CLI app with list command
			app := &cli.App{
				Name: "crgodicom-test",
				Commands: []*cli.Command{
					ListCommand(),
				},
				Before: func(c *cli.Context) error {
					c.Context = context.WithValue(c.Context, "config", cfg)
					return nil
				},
			}
			
			// Add output-dir flag to args if not already specified
			args := append([]string{"crgodicom-test"}, tt.args...)
			hasOutputDir := false
			for i, arg := range args {
				if arg == "--output-dir" && i+1 < len(args) {
					hasOutputDir = true
					break
				}
			}
			if !hasOutputDir {
				args = append(args, "--output-dir", tempDir)
			}
			
			err := app.Run(args)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestListCommandFlags(t *testing.T) {
	cmd := ListCommand()
	
	// Test that all expected flags are present
	expectedFlags := []string{
		"output-dir", "format", "verbose",
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
}

func TestListCommandDefaults(t *testing.T) {
	cmd := ListCommand()
	
	// Test default values
	defaults := map[string]interface{}{
		"output-dir": "studies",
		"format":     "table",
	}
	
	for flagName, expectedDefault := range defaults {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == flagName {
				found = true
				switch f := flag.(type) {
				case *cli.StringFlag:
					assert.Equal(t, expectedDefault, f.Value, "Default value mismatch for %s", flagName)
				}
				break
			}
		}
		assert.True(t, found, "Flag %s not found", flagName)
	}
}

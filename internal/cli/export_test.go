package cli

import (
	"context"
	"testing"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestExportCommand(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
		errMsg  string
	}{
		{
			name:    "missing study-id",
			args:    []string{"export", "--format", "png"},
			wantErr: true,
			errMsg:  "Required flag \"study-id\" not set",
		},
		{
			name:    "missing format",
			args:    []string{"export", "--study-id", "1.2.3.4.5"},
			wantErr: true,
			errMsg:  "Required flag \"format\" not set",
		},
		{
			name:    "invalid format",
			args:    []string{"export", "--study-id", "1.2.3.4.5", "--format", "invalid"},
			wantErr: true,
			errMsg:  "unsupported format",
		},
		{
			name:    "pdf without output-file",
			args:    []string{"export", "--study-id", "1.2.3.4.5", "--format", "pdf"},
			wantErr: true,
			errMsg:  "PDF format requires --output-file parameter",
		},
		{
			name:    "valid png export",
			args:    []string{"export", "--study-id", "1.2.3.4.5", "--format", "png"},
			wantErr: false, // Will fail due to missing study, but command structure is valid
		},
		{
			name:    "valid pdf export",
			args:    []string{"export", "--study-id", "1.2.3.4.5", "--format", "pdf", "--output-file", "test.pdf"},
			wantErr: false, // Will fail due to missing study, but command structure is valid
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()
			
			// Create test config
			cfg := config.DefaultConfig()
			cfg.Storage.BaseDir = tempDir
			
			// Create CLI app with export command
			app := &cli.App{
				Name: "crgodicom-test",
				Commands: []*cli.Command{
					ExportCommand(),
				},
				Before: func(c *cli.Context) error {
					c.Context = context.WithValue(c.Context, "config", cfg)
					return nil
				},
			}
			
			// Add input-dir flag to args
			args := append([]string{"crgodicom-test"}, tt.args...)
			args = append(args, "--input-dir", tempDir)
			
			err := app.Run(args)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				// For valid commands that fail due to missing study, that's expected
				// We're testing command structure, not the actual export functionality
				if err != nil {
					assert.Contains(t, err.Error(), "study not found")
				}
			}
		})
	}
}

func TestExportCommandFlags(t *testing.T) {
	cmd := ExportCommand()
	
	// Test that all expected flags are present
	expectedFlags := []string{
		"study-id", "format", "output-dir", "output-file",
		"input-dir", "include-metadata",
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
	requiredFlags := []string{"study-id", "format"}
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

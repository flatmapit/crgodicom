package cli

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestCreateCommand(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid basic create",
			args: []string{"create", "--study-count", "1", "--series-count", "1", "--image-count", "1"},
			wantErr: false,
		},
		{
			name: "valid template create",
			args: []string{"create", "--template", "chest-xray"},
			wantErr: false,
		},
		{
			name: "invalid modality",
			args: []string{"create", "--modality", "INVALID"},
			wantErr: true,
			errMsg: "unsupported modality",
		},
		{
			name: "negative study count",
			args: []string{"create", "--study-count", "-1"},
			wantErr: true,
		},
		{
			name: "zero image count",
			args: []string{"create", "--image-count", "0"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for test
			tempDir := t.TempDir()
			
			// Create test config
			cfg := config.DefaultConfig()
			cfg.Storage.BaseDir = tempDir
			
			// Create CLI app with create command
			app := &cli.App{
				Name: "crgodicom-test",
				Commands: []*cli.Command{
					CreateCommand(),
				},
				Before: func(c *cli.Context) error {
					c.Context = context.WithValue(c.Context, "config", cfg)
					return nil
				},
			}
			
			// Add output-dir flag to args
			args := append([]string{"crgodicom-test"}, tt.args...)
			args = append(args, "--output-dir", tempDir)
			
			err := app.Run(args)
			
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
				
				// Check if study directory was created
				studyDirs, err := filepath.Glob(filepath.Join(tempDir, "*"))
				assert.NoError(t, err)
				assert.NotEmpty(t, studyDirs, "Expected at least one study directory to be created")
			}
		})
	}
}

func TestCreateCommandFlags(t *testing.T) {
	cmd := CreateCommand()
	
	// Test that all expected flags are present
	expectedFlags := []string{
		"study-count", "series-count", "image-count", "modality",
		"template", "anatomical-region", "patient-id", "patient-name",
		"accession-number", "study-description", "output-dir",
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

func TestCreateCommandDefaults(t *testing.T) {
	cmd := CreateCommand()
	
	// Test default values
	defaults := map[string]interface{}{
		"study-count":  1,
		"series-count": 1,
		"image-count":  1,
		"modality":     "CR",
		"anatomical-region": "chest",
		"output-dir":   "studies",
	}
	
	for flagName, expectedDefault := range defaults {
		found := false
		for _, flag := range cmd.Flags {
			if flag.Names()[0] == flagName {
				found = true
				switch f := flag.(type) {
				case *cli.IntFlag:
					assert.Equal(t, expectedDefault, f.Value, "Default value mismatch for %s", flagName)
				case *cli.StringFlag:
					assert.Equal(t, expectedDefault, f.Value, "Default value mismatch for %s", flagName)
				}
				break
			}
		}
		assert.True(t, found, "Flag %s not found", flagName)
	}
}

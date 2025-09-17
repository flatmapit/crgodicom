package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	DICOM         DICOMConfig         `yaml:"dicom"`
	DefaultPACS   PACSConfig          `yaml:"default_pacs"`
	StudyTemplates map[string]TemplateConfig `yaml:"study_templates"`
	Logging       LoggingConfig       `yaml:"logging"`
	Storage       StorageConfig       `yaml:"storage"`
	TestPACS      map[string]PACSConfig `yaml:"test_pacs"`
}

// DICOMConfig contains DICOM-specific configuration
type DICOMConfig struct {
	OrgRoot string `yaml:"org_root"`
}

// PACSConfig represents PACS connection configuration
type PACSConfig struct {
	Host    string `yaml:"host"`
	Port    int    `yaml:"port"`
	AEC     string `yaml:"aec"`
	AET     string `yaml:"aet"`
	Timeout int    `yaml:"timeout"`
}

// TemplateConfig represents a study template configuration
type TemplateConfig struct {
	Modality          string `yaml:"modality"`
	SeriesCount       int    `yaml:"series_count"`
	ImageCount        int    `yaml:"image_count"`
	AnatomicalRegion  string `yaml:"anatomical_region"`
	StudyDescription  string `yaml:"study_description"`
	PatientName       string `yaml:"patient_name,omitempty"`
	PatientID         string `yaml:"patient_id,omitempty"`
	AccessionNumber   string `yaml:"accession_number,omitempty"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	File   string `yaml:"file"`
	Format string `yaml:"format"`
}

// StorageConfig represents storage configuration
type StorageConfig struct {
	BaseDir      string `yaml:"base_dir"`
	Compression  bool   `yaml:"compression"`
	IndexCache   bool   `yaml:"index_cache"`
}

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configPath)
	}

	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate and set defaults
	config.validateAndSetDefaults()

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		DICOM: DICOMConfig{
			OrgRoot: "1.2.840.10008.5.1.4.1.1",
		},
		DefaultPACS: PACSConfig{
			Host:    "localhost",
			Port:    11112,
			AEC:     "CRGODICOM",
			AET:     "PACS_SERVER",
			Timeout: 30,
		},
		StudyTemplates: getBuiltInTemplates(),
		Logging: LoggingConfig{
			Level:  "INFO",
			File:   "crgodicom.log",
			Format: "json",
		},
		Storage: StorageConfig{
			BaseDir:     "studies",
			Compression: false,
			IndexCache:  true,
		},
		TestPACS: map[string]PACSConfig{
			"orthanc_1": {
				Host:    "localhost",
				Port:    4900,
				AEC:     "CRGODICOM",
				AET:     "ORTHANC1",
				Timeout: 30,
			},
			"orthanc_2": {
				Host:    "localhost",
				Port:    4901,
				AEC:     "CRGODICOM",
				AET:     "ORTHANC2",
				Timeout: 30,
			},
		},
	}
}

// validateAndSetDefaults validates configuration and sets defaults
func (c *Config) validateAndSetDefaults() {
	// Set default DICOM org root if not specified
	if c.DICOM.OrgRoot == "" {
		c.DICOM.OrgRoot = "1.2.840.10008.5.1.4.1.1"
	}

	// Set default PACS config if not specified
	if c.DefaultPACS.Host == "" {
		c.DefaultPACS = PACSConfig{
			Host:    "localhost",
			Port:    11112,
			AEC:     "CRGODICOM",
			AET:     "PACS_SERVER",
			Timeout: 30,
		}
	}

	// Set default study templates if not specified
	if c.StudyTemplates == nil {
		c.StudyTemplates = getBuiltInTemplates()
	}

	// Set default logging config if not specified
	if c.Logging.Level == "" {
		c.Logging.Level = "INFO"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "json"
	}
	if c.Logging.File == "" {
		c.Logging.File = "crgodicom.log"
	}

	// Set default storage config if not specified
	if c.Storage.BaseDir == "" {
		c.Storage.BaseDir = "studies"
	}
}

// getBuiltInTemplates returns the built-in study templates
func getBuiltInTemplates() map[string]TemplateConfig {
	return map[string]TemplateConfig{
		"chest-xray": {
			Modality:         "CR",
			SeriesCount:      1,
			ImageCount:       2,
			AnatomicalRegion: "chest",
			StudyDescription: "Chest X-Ray",
		},
		"ct-chest": {
			Modality:         "CT",
			SeriesCount:      2,
			ImageCount:       50,
			AnatomicalRegion: "chest",
			StudyDescription: "CT Chest",
		},
		"ultrasound-abdomen": {
			Modality:         "US",
			SeriesCount:      1,
			ImageCount:       10,
			AnatomicalRegion: "abdomen",
			StudyDescription: "Ultrasound Abdomen",
		},
		"mammography": {
			Modality:         "MG",
			SeriesCount:      1,
			ImageCount:       4,
			AnatomicalRegion: "breast",
			StudyDescription: "Mammography",
		},
		"digital-xray": {
			Modality:         "DX",
			SeriesCount:      1,
			ImageCount:       1,
			AnatomicalRegion: "chest",
			StudyDescription: "Digital X-Ray",
		},
		"mri-brain": {
			Modality:         "MR",
			SeriesCount:      3,
			ImageCount:       30,
			AnatomicalRegion: "brain",
			StudyDescription: "MRI Brain",
		},
	}
}

// GetTemplate returns a study template by name
func (c *Config) GetTemplate(name string) (TemplateConfig, bool) {
	template, exists := c.StudyTemplates[name]
	return template, exists
}

// ListTemplates returns all available template names
func (c *Config) ListTemplates() []string {
	var templates []string
	for name := range c.StudyTemplates {
		templates = append(templates, name)
	}
	return templates
}

package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/pacs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// SendCommand returns the send command
func SendCommand() *cli.Command {
	return &cli.Command{
		Name:  "send",
		Usage: "Send DICOM study to PACS",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "study-id",
				Usage:    "Study Instance UID (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "host",
				Usage: "PACS host address",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "PACS port",
				Value: 11112,
			},
			&cli.StringFlag{
				Name:  "aec",
				Usage: "Application Entity Caller",
			},
			&cli.StringFlag{
				Name:  "aet",
				Usage: "Application Entity Title",
			},
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Studies directory",
				Value: "studies",
			},
			&cli.IntFlag{
				Name:  "timeout",
				Usage: "Connection timeout in seconds",
				Value: 30,
			},
			&cli.IntFlag{
				Name:  "retries",
				Usage: "Retry attempts",
				Value: 3,
			},
		},
		Action: sendAction,
	}
}

func sendAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	// Build PACS connection parameters
	pacsConfig := config.PACSConfig{
		Host:    c.String("host"),
		Port:    c.Int("port"),
		AEC:     c.String("aec"),
		AET:     c.String("aet"),
		Timeout: c.Int("timeout"),
	}

	// Use default PACS config if not specified via CLI
	if pacsConfig.Host == "" {
		pacsConfig = cfg.DefaultPACS
		logrus.Info("Using default PACS configuration")
	}

	// Validate required PACS parameters
	if pacsConfig.Host == "" || pacsConfig.AEC == "" || pacsConfig.AET == "" {
		return fmt.Errorf("PACS connection requires host, aec, and aet parameters")
	}

	studyID := c.String("study-id")
	outputDir := c.String("output-dir")
	retries := c.Int("retries")

	logrus.Infof("Sending study %s to PACS %s:%d (AEC: %s, AET: %s)",
		studyID, pacsConfig.Host, pacsConfig.Port, pacsConfig.AEC, pacsConfig.AET)
	logrus.Infof("Studies directory: %s, Retries: %d, Timeout: %ds",
		outputDir, retries, pacsConfig.Timeout)

	// Create PACS client
	client := pacs.NewClient(&pacsConfig)
	
	// Connect to PACS
	if err := client.Connect(c.Context); err != nil {
		return fmt.Errorf("failed to connect to PACS: %w", err)
	}
	defer client.Disconnect()
	
	// Test connectivity with C-ECHO
	if err := client.CEcho(c.Context); err != nil {
		return fmt.Errorf("C-ECHO failed: %w", err)
	}
	
	// Find and send DICOM files for the study
	studyDir := filepath.Join(outputDir, studyID)
	dicomFiles, err := findDICOMFiles(studyDir)
	if err != nil {
		return fmt.Errorf("failed to find DICOM files: %w", err)
	}
	
	logrus.Infof("Found %d DICOM files to send", len(dicomFiles))
	
	successCount := 0
	for _, filePath := range dicomFiles {
		logrus.Infof("Sending %s", filePath)
		
		// Read DICOM file
		dicomData, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Errorf("Failed to read %s: %v", filePath, err)
			continue
		}
		
		// Extract SOP Instance UID from filename or data (simplified)
		sopInstanceUID := extractSOPInstanceUID(filePath)
		
		// Send to PACS
		if err := client.CStore(c.Context, dicomData, sopInstanceUID); err != nil {
			logrus.Errorf("Failed to send %s: %v", filePath, err)
			continue
		}
		
		successCount++
	}
	
	fmt.Printf("Successfully sent %d/%d DICOM files to PACS\n", successCount, len(dicomFiles))
	return nil
}

// findDICOMFiles recursively finds all DICOM files in a directory
func findDICOMFiles(dir string) ([]string, error) {
	var dicomFiles []string
	
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		// Check if it's a DICOM file (.dcm extension)
		if !info.IsDir() && filepath.Ext(path) == ".dcm" {
			dicomFiles = append(dicomFiles, path)
		}
		
		return nil
	})
	
	return dicomFiles, err
}

// extractSOPInstanceUID extracts SOP Instance UID from file path
// This is a simplified implementation - in a real scenario, you'd parse the DICOM file
func extractSOPInstanceUID(filePath string) string {
	// For now, use a simple approach based on filename
	// In a real implementation, you'd parse the DICOM file to extract the actual UID
	base := filepath.Base(filePath)
	if base == "image_001.dcm" {
		// This is a placeholder - in reality you'd extract from DICOM metadata
		return "1.2.840.10008.5.1.4.1.1.1.1"
	}
	return "1.2.840.10008.5.1.4.1.1.1.1" // Default placeholder
}

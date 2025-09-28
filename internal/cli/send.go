package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
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
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Enable verbose output",
			},
			&cli.BoolFlag{
				Name:  "debug",
				Usage: "Enable debug output",
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
	verbose := c.Bool("verbose")
	debug := c.Bool("debug")

	// Set log level based on flags
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	} else if verbose {
		logrus.SetLevel(logrus.InfoLevel)
	}

	logrus.Infof("🚀 Starting DICOM study transmission")
	logrus.Infof("📋 Study ID: %s", studyID)
	logrus.Infof("🎯 PACS Target: %s:%d (AEC: %s, AET: %s)",
		pacsConfig.Host, pacsConfig.Port, pacsConfig.AEC, pacsConfig.AET)
	logrus.Infof("📁 Studies directory: %s", outputDir)
	logrus.Infof("⚙️  Configuration: Retries=%d, Timeout=%ds, Verbose=%t, Debug=%t",
		retries, pacsConfig.Timeout, verbose, debug)

	// Create PACS client with options
	clientOptions := &pacs.ClientOptions{
		Verbose: verbose,
		Debug:   debug,
	}
	client := pacs.NewClientWithOptions(&pacsConfig, clientOptions)

	// Connect to PACS
	if err := client.Connect(c.Context); err != nil {
		return fmt.Errorf("❌ failed to connect to PACS: %w", err)
	}
	defer client.Disconnect()

	// Test connectivity with C-ECHO
	if err := client.CEcho(c.Context); err != nil {
		return fmt.Errorf("❌ C-ECHO failed: %w", err)
	}

	// Find and send DICOM files for the study
	studyDir := filepath.Join(outputDir, studyID)
	dicomFiles, err := findDICOMFiles(studyDir)
	if err != nil {
		return fmt.Errorf("❌ failed to find DICOM files in %s: %w", studyDir, err)
	}

	if len(dicomFiles) == 0 {
		return fmt.Errorf("❌ no DICOM files found in study directory: %s", studyDir)
	}

	logrus.Infof("📁 Found %d DICOM files to send from study directory: %s", len(dicomFiles), studyDir)

	successCount := 0
	failedFiles := []string{}

	for i, filePath := range dicomFiles {
		logrus.Infof("📤 [%d/%d] Sending %s", i+1, len(dicomFiles), filepath.Base(filePath))

		if debug {
			logrus.Debugf("📁 Full path: %s", filePath)
		}

		// Read DICOM file
		dicomData, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Errorf("❌ Failed to read %s: %v", filePath, err)
			failedFiles = append(failedFiles, fmt.Sprintf("%s (read error: %v)", filePath, err))
			continue
		}

		if debug {
			logrus.Debugf("📊 File size: %d bytes", len(dicomData))
		}

		// Extract SOP Instance UID from actual DICOM metadata
		sopInstanceUID := extractSOPInstanceUIDFromDICOM(filePath, cfg)
		if debug {
			logrus.Debugf("🆔 SOP Instance UID: %s", sopInstanceUID)
		}

		// Send to PACS with retry logic
		for attempt := 1; attempt <= retries; attempt++ {
			if attempt > 1 {
				logrus.Warnf("🔄 Retry attempt %d/%d for %s", attempt, retries, filepath.Base(filePath))
			}

			if err := client.CStore(c.Context, dicomData, sopInstanceUID); err != nil {
				if attempt < retries {
					logrus.Warnf("⚠️  Attempt %d failed for %s: %v", attempt, filepath.Base(filePath), err)
					continue
				}
				logrus.Errorf("❌ Failed to send %s after %d attempts: %v", filePath, retries, err)
				failedFiles = append(failedFiles, fmt.Sprintf("%s (send error: %v)", filePath, err))
				break
			}

			// Success
			logrus.Infof("✅ Successfully sent %s", filepath.Base(filePath))
			successCount++
			break
		}
	}

	// Summary
	fmt.Printf("\n📊 Transmission Summary:\n")
	fmt.Printf("✅ Successfully sent: %d/%d files\n", successCount, len(dicomFiles))
	if len(failedFiles) > 0 {
		fmt.Printf("❌ Failed files (%d):\n", len(failedFiles))
		for _, failedFile := range failedFiles {
			fmt.Printf("   • %s\n", failedFile)
		}
		return fmt.Errorf("transmission completed with %d failures out of %d files", len(failedFiles), len(dicomFiles))
	}

	fmt.Printf("🎉 All files transmitted successfully!\n")
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

// extractSOPInstanceUIDFromDICOM extracts SOP Instance UID from actual DICOM metadata
func extractSOPInstanceUIDFromDICOM(filePath string, cfg *config.Config) string {
	// Create DICOM reader
	reader := dicom.NewReader(cfg)

	// Try to extract real SOP Instance UID from DICOM file
	sopInstanceUID, err := reader.ExtractSOPInstanceUID(filePath)
	if err != nil {
		logrus.Warnf("Failed to extract SOP Instance UID from %s: %v", filePath, err)
		logrus.Warnf("Using fallback SOP Instance UID generation")

		// Fallback to the old placeholder approach
		return extractSOPInstanceUID(filePath)
	}

	logrus.Debugf("Extracted real SOP Instance UID: %s", sopInstanceUID)
	return sopInstanceUID
}

// extractSOPInstanceUID extracts SOP Instance UID from file path (fallback method)
// This is a simplified implementation - used as fallback when DICOM parsing fails
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

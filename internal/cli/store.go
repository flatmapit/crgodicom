package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/pacs"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// StoreCommand returns the C-STORE command for sending DICOM files
func StoreCommand() *cli.Command {
	return &cli.Command{
		Name:  "store",
		Usage: "Send DICOM files to PACS using C-STORE (bypasses C-ECHO)",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "study-id",
				Usage:    "Study Instance UID (required)",
				Required: true,
			},
			&cli.StringFlag{
				Name:  "host",
				Usage: "PACS host address",
				Value: "localhost",
			},
			&cli.IntFlag{
				Name:  "port",
				Usage: "PACS port",
				Value: 4242,
			},
			&cli.StringFlag{
				Name:  "aec",
				Usage: "Application Entity Caller",
				Value: "DICOM_CLIENT",
			},
			&cli.StringFlag{
				Name:  "aet",
				Usage: "Application Entity Title",
				Value: "PACS1",
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
		},
		Action: storeAction,
	}
}

func storeAction(c *cli.Context) error {
	studyID := c.String("study-id")
	outputDir := c.String("output-dir")

	logrus.Infof("Sending study %s to PACS %s:%d (AEC: %s, AET: %s)",
		studyID, c.String("host"), c.Int("port"), c.String("aec"), c.String("aet"))

	// Create PACS configuration
	pacsConfig := &config.PACSConfig{
		Host:    c.String("host"),
		Port:    c.Int("port"),
		AEC:     c.String("aec"),
		AET:     c.String("aet"),
		Timeout: c.Int("timeout"),
	}

	// Find DICOM files for the study
	studyDir := filepath.Join(outputDir, studyID)
	dicomFiles, err := findDICOMFilesInStudy(studyDir)
	if err != nil {
		return fmt.Errorf("failed to find DICOM files: %w", err)
	}

	if len(dicomFiles) == 0 {
		return fmt.Errorf("no DICOM files found for study %s", studyID)
	}

	logrus.Infof("Found %d DICOM files to send", len(dicomFiles))

	// Create PACS client
	client := pacs.NewClient(pacsConfig)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pacsConfig.Timeout)*time.Second)
	defer cancel()

	// Connect to PACS (association only)
	logrus.Info("Establishing DICOM association...")
	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to establish DICOM association: %w", err)
	}
	defer client.Disconnect()

	logrus.Info("DICOM association established successfully - bypassing C-ECHO")

	// Send each DICOM file directly with C-STORE
	successCount := 0
	for i, filePath := range dicomFiles {
		logrus.Infof("Sending file %d/%d: %s", i+1, len(dicomFiles), filepath.Base(filePath))

		// Read DICOM file
		dicomData, err := os.ReadFile(filePath)
		if err != nil {
			logrus.Errorf("Failed to read DICOM file %s: %v", filePath, err)
			continue
		}

		// Extract SOP Instance UID from filename (simplified)
		sopInstanceUID := extractSOPInstanceUIDFromPath(filePath)

		logrus.Debugf("Attempting C-STORE for SOP Instance UID: %s", sopInstanceUID)

		// Send C-STORE directly
		if err := client.CStore(ctx, dicomData, sopInstanceUID); err != nil {
			logrus.Errorf("C-STORE failed for %s: %v", filePath, err)
			continue
		}

		successCount++
		logrus.Infof("Successfully sent %s", filepath.Base(filePath))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to send any DICOM files")
	}

	logrus.Infof("Successfully sent %d/%d DICOM files to PACS", successCount, len(dicomFiles))
	return nil
}

// findDICOMFilesInStudy finds all DICOM files in a study directory
func findDICOMFilesInStudy(studyDir string) ([]string, error) {
	var dicomFiles []string

	err := filepath.Walk(studyDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && filepath.Ext(path) == ".dcm" {
			dicomFiles = append(dicomFiles, path)
		}

		return nil
	})

	return dicomFiles, err
}

// extractSOPInstanceUIDFromPath extracts SOP Instance UID from file path (simplified)
func extractSOPInstanceUIDFromPath(filePath string) string {
	// For now, generate a simple UID based on filename
	// In a real implementation, you'd parse the DICOM file
	return fmt.Sprintf("1.2.840.10008.1.1.%d", time.Now().UnixNano())
}

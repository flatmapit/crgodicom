package cli

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// DCMTKCommand returns the DCMTK subprocess command for sending DICOM files
func DCMTKCommand() *cli.Command {
	return &cli.Command{
		Name:  "dcmtk",
		Usage: "Send DICOM files using DCMTK storescu subprocess (100% compatibility)",
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
			&cli.BoolFlag{
				Name:  "verbose",
				Usage: "Verbose DCMTK output",
				Value: false,
			},
		},
		Action: dcmtkAction,
	}
}

func dcmtkAction(c *cli.Context) error {
	studyID := c.String("study-id")
	outputDir := c.String("output-dir")

	logrus.Infof("Sending study %s using DCMTK storescu to PACS %s:%d",
		studyID, c.String("host"), c.Int("port"))

	// Check DCMTK availability using the manager
	if err := CheckDCMTKAvailability(); err != nil {
		return fmt.Errorf("DCMTK not available: %w", err)
	}

	// Find DICOM files for the study
	studyDir := filepath.Join(outputDir, studyID)
	dicomFiles, err := findDICOMFilesForDCMTK(studyDir)
	if err != nil {
		return fmt.Errorf("failed to find DICOM files: %w", err)
	}

	if len(dicomFiles) == 0 {
		return fmt.Errorf("no DICOM files found for study %s", studyID)
	}

	logrus.Infof("Found %d DICOM files to send via DCMTK", len(dicomFiles))

	// Test connectivity first with echoscu
	logrus.Info("Testing PACS connectivity with DCMTK echoscu...")
	if err := runEchoSCU(c.String("host"), c.Int("port"), c.String("aec"), c.String("aet"), c.Bool("verbose")); err != nil {
		return fmt.Errorf("PACS connectivity test failed: %w", err)
	}

	logrus.Info("✅ PACS connectivity confirmed with DCMTK echoscu")

	// Send each DICOM file using storescu
	successCount := 0
	for i, filePath := range dicomFiles {
		logrus.Infof("Sending file %d/%d via DCMTK storescu: %s", i+1, len(dicomFiles), filepath.Base(filePath))

		if err := runStoreSCU(c.String("host"), c.Int("port"), c.String("aec"), c.String("aet"), filePath, c.Bool("verbose")); err != nil {
			logrus.Errorf("Failed to send %s via DCMTK: %v", filepath.Base(filePath), err)
			continue
		}

		successCount++
		logrus.Infof("✅ Successfully sent %s via DCMTK storescu", filepath.Base(filePath))
	}

	if successCount == 0 {
		return fmt.Errorf("failed to send any DICOM files via DCMTK")
	}

	logrus.Infof("🎉 Successfully sent %d/%d DICOM files to PACS via DCMTK", successCount, len(dicomFiles))
	return nil
}

// runEchoSCU runs DCMTK echoscu to test connectivity
func runEchoSCU(host string, port int, aec, aet string, verbose bool) error {
	// Get DCMTK tool path
	echoscuPath, err := GetDCMTKPath("echoscu")
	if err != nil {
		return fmt.Errorf("failed to get echoscu path: %w", err)
	}
	
	args := []string{
		"-aec", aec,
		"-aet", aet,
		host,
		fmt.Sprintf("%d", port),
	}

	if verbose {
		args = append([]string{"-v"}, args...)
	}

	cmd := exec.Command(echoscuPath, args...)

	logrus.Debugf("Running: %s %s", echoscuPath, strings.Join(args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("echoscu failed: %w\nOutput: %s", err, string(output))
	}

	logrus.Debugf("echoscu output: %s", string(output))
	return nil
}

// runStoreSCU runs DCMTK storescu to send a DICOM file
func runStoreSCU(host string, port int, aec, aet, filePath string, verbose bool) error {
	// Get DCMTK tool path
	storescuPath, err := GetDCMTKPath("storescu")
	if err != nil {
		return fmt.Errorf("failed to get storescu path: %w", err)
	}
	
	args := []string{
		"-aec", aec,
		"-aet", aet,
		host,
		fmt.Sprintf("%d", port),
		filePath,
	}

	if verbose {
		args = append([]string{"-v"}, args...)
	}

	logrus.Debugf("Running: %s %s", storescuPath, strings.Join(args, " "))

	// Set timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, storescuPath, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("storescu failed: %w\nOutput: %s", err, string(output))
	}

	// Check for success indicators in output
	outputStr := string(output)
	if strings.Contains(outputStr, "Store Response (Success)") ||
		strings.Contains(outputStr, "Received Store Response (Success)") ||
		strings.Contains(outputStr, "I: Received Store Response (Success)") ||
		(strings.Contains(outputStr, "Association Accepted") &&
			strings.Contains(outputStr, "Sending Store Request") &&
			!strings.Contains(outputStr, "Store SCU Failed")) {
		logrus.Debugf("storescu success detected")
		return nil
	}

	// Check for specific failure indicators
	if strings.Contains(outputStr, "Store SCU Failed") ||
		strings.Contains(outputStr, "No presentation context") ||
		strings.Contains(outputStr, "Aborting Association") {
		return fmt.Errorf("storescu failed: %s", outputStr)
	}

	return fmt.Errorf("storescu status unclear: %s", outputStr)
}

// findDICOMFilesForDCMTK finds all DICOM files in a study directory
func findDICOMFilesForDCMTK(studyDir string) ([]string, error) {
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

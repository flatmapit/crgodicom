package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// ListCommand returns the list command
func ListCommand() *cli.Command {
	return &cli.Command{
		Name:  "list",
		Usage: "List local DICOM studies",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "output-dir",
				Usage: "Studies directory",
				Value: "studies",
			},
			&cli.StringFlag{
				Name:  "format",
				Usage: "Output format: table, json, csv",
				Value: "table",
			},
			&cli.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Show detailed information",
			},
		},
		Action: listAction,
	}
}

func listAction(c *cli.Context) error {
	// Get configuration from context
	cfg, ok := c.Context.Value("config").(*config.Config)
	if !ok {
		return fmt.Errorf("configuration not found in context")
	}

	outputDir := c.String("output-dir")
	format := c.String("format")
	verbose := c.Bool("verbose")

	// Validate format
	validFormats := []string{"table", "json", "csv"}
	validFormat := false
	for _, f := range validFormats {
		if format == f {
			validFormat = true
			break
		}
	}
	if !validFormat {
		return fmt.Errorf("invalid format '%s'. Valid formats: %v", format, validFormats)
	}

	logrus.Infof("Listing studies in directory: %s", outputDir)
	logrus.Infof("Output format: %s, Verbose: %v", format, verbose)

	// Check if output directory exists
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		logrus.Warnf("Studies directory does not exist: %s", outputDir)
		fmt.Printf("No studies directory found at: %s\n", outputDir)
		fmt.Println("Use 'crgodicom create' to generate some studies first.")
		return nil
	}

	// List studies with real DICOM metadata
	studies, err := listStudies(outputDir, cfg)
	if err != nil {
		return fmt.Errorf("failed to list studies: %w", err)
	}

	if len(studies) == 0 {
		fmt.Println("No studies found.")
		return nil
	}

	// Display studies based on format
	switch format {
	case "table":
		displayStudiesTable(studies, verbose)
	case "json":
		displayStudiesJSON(studies, verbose)
	case "csv":
		displayStudiesCSV(studies, verbose)
	}

	return nil
}

// StudyInfo represents basic study information
type StudyInfo struct {
	StudyUID         string
	PatientName      string
	PatientID        string
	StudyDate        string
	StudyDescription string
	SeriesCount      int
	ImageCount       int
	Modality         string
	AccessionNumber  string
}

// listStudies lists all studies in the directory
func listStudies(outputDir string, cfg *config.Config) ([]StudyInfo, error) {
	var studies []StudyInfo

	// Create DICOM reader
	reader := dicom.NewReader(cfg)

	// Walk through the studies directory
	err := filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Look for study directories (directories with UID-like names)
		if info.IsDir() && path != outputDir {
			// Check if this looks like a study directory
			studyUID := filepath.Base(path)
			if isUIDFormat(studyUID) {
				studyInfo, err := getStudyInfo(path, reader)
				if err != nil {
					logrus.Warnf("Failed to read study info for %s: %v", studyUID, err)
					// Still include it but with minimal info
					studies = append(studies, StudyInfo{
						StudyUID: studyUID,
					})
				} else {
					studies = append(studies, studyInfo)
				}
			}
		}

		return nil
	})

	return studies, err
}

// isUIDFormat checks if a string looks like a DICOM UID
func isUIDFormat(s string) bool {
	// Basic UID format check (contains dots and digits)
	if len(s) < 10 {
		return false
	}

	// Check if it's a series directory (starts with "series_")
	if strings.HasPrefix(s, "series_") {
		return false
	}

	// Check if it contains dots (UIDs have dots)
	if !strings.Contains(s, ".") {
		return false
	}

	// More sophisticated UID validation could be added here
	return true
}

// getStudyInfo reads study information from directory using real DICOM metadata
func getStudyInfo(studyPath string, reader *dicom.Reader) (StudyInfo, error) {
	// Read actual DICOM metadata from files
	metadata, err := reader.ReadStudyMetadata(studyPath)
	if err != nil {
		return StudyInfo{}, fmt.Errorf("failed to read study metadata: %w", err)
	}

	// Convert DICOM metadata to StudyInfo
	studyInfo := StudyInfo{
		StudyUID:         metadata.StudyUID,
		PatientName:      metadata.PatientName,
		PatientID:        metadata.PatientID,
		StudyDate:        metadata.StudyDate,
		StudyDescription: metadata.StudyDescription,
		SeriesCount:      metadata.SeriesCount,
		ImageCount:       metadata.ImageCount,
		Modality:         metadata.Modality,
		AccessionNumber:  metadata.AccessionNumber,
	}

	return studyInfo, nil
}

// displayStudiesTable displays studies in table format
func displayStudiesTable(studies []StudyInfo, verbose bool) {
	if verbose {
		fmt.Printf("%-40s %-20s %-12s %-15s %-8s %-8s %-8s %-12s\n",
			"Study UID", "Patient Name", "Patient ID", "Study Date", "Series", "Images", "Modality", "Accession")
		fmt.Println(strings.Repeat("-", 120))
	} else {
		fmt.Printf("%-40s %-20s %-12s %-15s %-8s %-8s\n",
			"Study UID", "Patient Name", "Patient ID", "Study Date", "Series", "Images")
		fmt.Println(strings.Repeat("-", 95))
	}

	for _, study := range studies {
		if verbose {
			fmt.Printf("%-40s %-20s %-12s %-15s %-8d %-8d %-8s %-12s\n",
				study.StudyUID, study.PatientName, study.PatientID, study.StudyDate,
				study.SeriesCount, study.ImageCount, study.Modality, study.AccessionNumber)
		} else {
			fmt.Printf("%-40s %-20s %-12s %-15s %-8d %-8d\n",
				study.StudyUID, study.PatientName, study.PatientID, study.StudyDate,
				study.SeriesCount, study.ImageCount)
		}
	}
}

// displayStudiesJSON displays studies in JSON format
func displayStudiesJSON(studies []StudyInfo, verbose bool) {
	// Convert studies to JSON
	jsonData, err := json.MarshalIndent(studies, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Println(string(jsonData))
}

// displayStudiesCSV displays studies in CSV format
func displayStudiesCSV(studies []StudyInfo, verbose bool) {
	if len(studies) == 0 {
		return
	}

	// Print CSV header
	if verbose {
		fmt.Println("StudyUID,PatientName,PatientID,StudyDate,StudyDescription,SeriesCount,ImageCount,Modality,AccessionNumber")
	} else {
		fmt.Println("StudyUID,PatientName,PatientID,StudyDate,SeriesCount,ImageCount")
	}

	// Print CSV data rows
	for _, study := range studies {
		if verbose {
			fmt.Printf("%s,%s,%s,%s,%s,%d,%d,%s,%s\n",
				study.StudyUID,
				strings.ReplaceAll(study.PatientName, ",", " "),
				study.PatientID,
				study.StudyDate,
				strings.ReplaceAll(study.StudyDescription, ",", " "),
				study.SeriesCount,
				study.ImageCount,
				study.Modality,
				study.AccessionNumber)
		} else {
			fmt.Printf("%s,%s,%s,%s,%d,%d\n",
				study.StudyUID,
				strings.ReplaceAll(study.PatientName, ",", " "),
				study.PatientID,
				study.StudyDate,
				study.SeriesCount,
				study.ImageCount)
		}
	}
}

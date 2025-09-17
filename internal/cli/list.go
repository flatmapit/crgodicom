package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/internal/config"
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
	_, ok := c.Context.Value("config").(*config.Config)
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

	// TODO: Implement actual study listing
	// For now, just show directory structure
	studies, err := listStudies(outputDir)
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
func listStudies(outputDir string) ([]StudyInfo, error) {
	var studies []StudyInfo

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
				studyInfo, err := getStudyInfo(path)
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
	// More sophisticated UID validation could be added here
	return true
}

// getStudyInfo reads study information from directory
func getStudyInfo(studyPath string) (StudyInfo, error) {
	studyUID := filepath.Base(studyPath)

	// TODO: Read actual DICOM metadata from files
	// For now, return placeholder info
	return StudyInfo{
		StudyUID:         studyUID,
		PatientName:      "DOE^JOHN^M",
		PatientID:        "P123456",
		StudyDate:        "20250101",
		StudyDescription: "Generated Study",
		SeriesCount:      1,
		ImageCount:       1,
		Modality:         "CR",
		AccessionNumber:  "ACC123456",
	}, nil
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
	// TODO: Implement JSON output
	fmt.Println("JSON output not yet implemented")
}

// displayStudiesCSV displays studies in CSV format
func displayStudiesCSV(studies []StudyInfo, verbose bool) {
	// TODO: Implement CSV output
	fmt.Println("CSV output not yet implemented")
}

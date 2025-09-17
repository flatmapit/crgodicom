package export

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"

	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
)

// Exporter handles DICOM export to various formats
type Exporter struct {
	outputDir string
}

// NewExporter creates a new exporter
func NewExporter(outputDir string) *Exporter {
	return &Exporter{
		outputDir: outputDir,
	}
}

// ExportStudy exports a study to PNG and PDF formats
func (e *Exporter) ExportStudy(study *types.Study) error {
	studyDir := filepath.Join(e.outputDir, study.StudyInstanceUID)
	
	logrus.Infof("Exporting study %s to %s", study.StudyInstanceUID, studyDir)
	
	// Create export directory
	exportDir := filepath.Join(studyDir, "exports")
	if err := os.MkdirAll(exportDir, 0755); err != nil {
		return fmt.Errorf("failed to create export directory: %w", err)
	}
	
	// Export each series
	var allImages []string
	for i, series := range study.Series {
		seriesExportDir := filepath.Join(exportDir, fmt.Sprintf("series_%03d", i+1))
		if err := os.MkdirAll(seriesExportDir, 0755); err != nil {
			return fmt.Errorf("failed to create series export directory: %w", err)
		}
		
		// Export images in this series
		seriesImages, err := e.exportSeries(&series, seriesExportDir)
		if err != nil {
			return fmt.Errorf("failed to export series %d: %w", i+1, err)
		}
		
		allImages = append(allImages, seriesImages...)
	}
	
	// Create PDF report
	pdfPath := filepath.Join(exportDir, fmt.Sprintf("study_%s_report.pdf", study.StudyInstanceUID))
	if err := e.createPDFReport(study, allImages, pdfPath); err != nil {
		return fmt.Errorf("failed to create PDF report: %w", err)
	}
	
	logrus.Infof("Successfully exported study to %s", exportDir)
	return nil
}

// exportSeries exports all images in a series to PNG
func (e *Exporter) exportSeries(series *types.Series, exportDir string) ([]string, error) {
	var exportedImages []string
	
	logrus.Infof("Exporting series %s with %d images", series.SeriesInstanceUID, len(series.Images))
	
	for i, image := range series.Images {
		pngPath := filepath.Join(exportDir, fmt.Sprintf("image_%03d.png", i+1))
		
		if err := e.exportImageToPNG(&image, pngPath); err != nil {
			return nil, fmt.Errorf("failed to export image %d: %w", i+1, err)
		}
		
		exportedImages = append(exportedImages, pngPath)
	}
	
	return exportedImages, nil
}

// exportImageToPNG exports a DICOM image to PNG format
func (e *Exporter) exportImageToPNG(img *types.Image, outputPath string) error {
	// Create grayscale image from pixel data
	image := image.NewGray(image.Rect(0, 0, img.Width, img.Height))
	
	// Convert pixel data to grayscale image
	bytesPerPixel := img.BitsPerPixel / 8
	if img.BitsPerPixel%8 != 0 {
		bytesPerPixel++
	}
	
	for y := 0; y < img.Height; y++ {
		for x := 0; x < img.Width; x++ {
			idx := (y*img.Width + x) * bytesPerPixel
			
			if idx+bytesPerPixel <= len(img.PixelData) {
				var pixelValue uint8
				if bytesPerPixel == 2 {
					// 16-bit to 8-bit conversion
					pixelValue = uint8(img.PixelData[idx] >> 8)
				} else {
					pixelValue = img.PixelData[idx]
				}
				
				image.SetGray(x, y, color.Gray{Y: pixelValue})
			}
		}
	}
	
	// Save as PNG
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create PNG file: %w", err)
	}
	defer file.Close()
	
	if err := png.Encode(file, image); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}
	
	logrus.Debugf("Exported image to %s", outputPath)
	return nil
}

// createPDFReport creates a PDF report with study information and images
func (e *Exporter) createPDFReport(study *types.Study, imagePaths []string, outputPath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	
	// Set font
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, "DICOM Study Report")
	pdf.Ln(15)
	
	// Study information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Study Information:")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Study Instance UID: %s", study.StudyInstanceUID))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Patient Name: %s", study.PatientName))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Patient ID: %s", study.PatientID))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Study Date: %s", study.StudyDate))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Study Time: %s", study.StudyTime))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Study Description: %s", study.StudyDescription))
	pdf.Ln(6)
	pdf.Cell(40, 6, fmt.Sprintf("Accession Number: %s", study.AccessionNumber))
	pdf.Ln(15)
	
	// Series information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Series Information:")
	pdf.Ln(10)
	
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(40, 6, fmt.Sprintf("Number of Series: %d", len(study.Series)))
	pdf.Ln(6)
	
	for i, series := range study.Series {
		pdf.Cell(40, 6, fmt.Sprintf("Series %d: %s (%s) - %d images", 
			i+1, series.SeriesDescription, series.Modality, len(series.Images)))
		pdf.Ln(6)
	}
	pdf.Ln(10)
	
	// Images section
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 8, "Study Images:")
	pdf.Ln(10)
	
	// Add images to PDF
	for i, imagePath := range imagePaths {
		if i%2 == 0 {
			pdf.AddPage()
		}
		
		pdf.SetFont("Arial", "", 8)
		pdf.Cell(40, 4, fmt.Sprintf("Image %d: %s", i+1, filepath.Base(imagePath)))
		pdf.Ln(4)
		
		// Add image to PDF (simplified - just add placeholder text for now)
		pdf.SetFont("Arial", "I", 10)
		pdf.Cell(40, 6, "[DICOM Image would be displayed here]")
		pdf.Ln(10)
	}
	
	// Save PDF
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}
	
	logrus.Infof("Created PDF report: %s", outputPath)
	return nil
}

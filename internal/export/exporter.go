package export

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/flatmapit/crgodicom/pkg/types"
	"github.com/jung-kurt/gofpdf"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
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
		seriesImages, err := e.exportSeries(study, &series, seriesExportDir)
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

// exportSeries exports all images in a series to PNG and JPEG
func (e *Exporter) exportSeries(study *types.Study, series *types.Series, exportDir string) ([]string, error) {
	var exportedImages []string

	logrus.Infof("Exporting series %s with %d images", series.SeriesInstanceUID, len(series.Images))

	for i, image := range series.Images {
		// Export as PNG
		pngPath := filepath.Join(exportDir, fmt.Sprintf("image_%03d.png", i+1))
		if err := e.exportImageToPNG(study, series, &image, i+1, len(series.Images), pngPath); err != nil {
			logrus.Errorf("Failed to export PNG image %d: %v", i+1, err)
			continue
		}
		exportedImages = append(exportedImages, pngPath)

		// Export as JPEG
		jpegPath := filepath.Join(exportDir, fmt.Sprintf("image_%03d.jpg", i+1))
		logrus.Infof("Attempting to export JPEG image %d to %s", i+1, jpegPath)
		if err := e.exportImageToJPEG(study, series, &image, i+1, len(series.Images), jpegPath); err != nil {
			logrus.Errorf("Failed to export JPEG image %d: %v", i+1, err)
			continue
		}
		logrus.Infof("Successfully exported JPEG image %d to %s", i+1, jpegPath)
		exportedImages = append(exportedImages, jpegPath)
	}

	return exportedImages, nil
}

// exportImageToJPEG exports a DICOM image to JPEG format with burnt-in metadata
func (e *Exporter) exportImageToJPEG(study *types.Study, series *types.Series, img *types.Image, instanceNum, totalInstances int, outputPath string) error {
	logrus.Debugf("Starting JPEG export for image %dx%d, %d bits per pixel, %d bytes pixel data", img.Width, img.Height, img.BitsPerPixel, len(img.PixelData))
	
	// Create grayscale image from pixel data
	grayImage := image.NewGray(image.Rect(0, 0, img.Width, img.Height))

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
					// 16-bit to 8-bit conversion (little-endian)
					lowByte := img.PixelData[idx]
					highByte := img.PixelData[idx+1]
					value16 := uint16(lowByte) | (uint16(highByte) << 8)
					// Scale 16-bit value (0-65535) to 8-bit (0-255)
					pixelValue = uint8(value16 >> 8)
				} else {
					pixelValue = img.PixelData[idx]
				}

				grayImage.SetGray(x, y, color.Gray{Y: pixelValue})
			}
		}
	}

	// Add burnt-in metadata text
	if err := e.addBurntInText(grayImage, study, series, img, instanceNum, totalInstances); err != nil {
		return fmt.Errorf("failed to add burnt-in text: %w", err)
	}

	// Convert grayscale to RGB for JPEG encoding
	rgbaImg := image.NewRGBA(grayImage.Bounds())
	draw.Draw(rgbaImg, grayImage.Bounds(), grayImage, image.Point{}, draw.Src)

	// Save as JPEG
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create JPEG file: %w", err)
	}
	defer file.Close()

	// JPEG quality set to 95 for high quality medical images
	if err := jpeg.Encode(file, rgbaImg, &jpeg.Options{Quality: 95}); err != nil {
		return fmt.Errorf("failed to encode JPEG: %w", err)
	}

	logrus.Debugf("Exported image to %s", outputPath)
	return nil
}

// exportImageToPNG exports a DICOM image to PNG format with burnt-in metadata
func (e *Exporter) exportImageToPNG(study *types.Study, series *types.Series, img *types.Image, instanceNum, totalInstances int, outputPath string) error {
	// Create grayscale image from pixel data
	grayImage := image.NewGray(image.Rect(0, 0, img.Width, img.Height))

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
					// 16-bit to 8-bit conversion (little-endian)
					lowByte := img.PixelData[idx]
					highByte := img.PixelData[idx+1]
					value16 := uint16(lowByte) | (uint16(highByte) << 8)
					// Scale 16-bit value (0-65535) to 8-bit (0-255)
					pixelValue = uint8(value16 >> 8)
				} else {
					pixelValue = img.PixelData[idx]
				}

				grayImage.SetGray(x, y, color.Gray{Y: pixelValue})
			}
		}
	}

	// Add burnt-in metadata text
	if err := e.addBurntInText(grayImage, study, series, img, instanceNum, totalInstances); err != nil {
		return fmt.Errorf("failed to add burnt-in text: %w", err)
	}

	// Save as PNG
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create PNG file: %w", err)
	}
	defer file.Close()

	if err := png.Encode(file, grayImage); err != nil {
		return fmt.Errorf("failed to encode PNG: %w", err)
	}

	logrus.Debugf("Exported image to %s", outputPath)
	return nil
}

// addBurntInText adds metadata text to the top-left corner of the image
func (e *Exporter) addBurntInText(img *image.Gray, study *types.Study, series *types.Series, dicomImg *types.Image, instanceNum, totalInstances int) error {
	// Extract body part/anatomical region from study description
	bodyPart := e.extractBodyPart(study.StudyDescription)

	// Create text lines for burnt-in metadata
	textLines := []string{
		fmt.Sprintf("Patient: %s", study.PatientName),
		fmt.Sprintf("Patient ID: %s", study.PatientID),
		fmt.Sprintf("DOB: %s", e.formatDate(study.PatientBirthDate)),
		fmt.Sprintf("Accession: %s", study.AccessionNumber),
		fmt.Sprintf("Study UID: %s", study.StudyInstanceUID),
		fmt.Sprintf("Series UID: %s", series.SeriesInstanceUID),
		fmt.Sprintf("Instance: %d of %d", instanceNum, totalInstances),
		fmt.Sprintf("Modality: %s", series.Modality),
		fmt.Sprintf("Body Part: %s", bodyPart),
		fmt.Sprintf("Study Date: %s", e.formatDate(study.StudyDate)),
		"Generated by crgodicom flatmapit.com",
	}

	// Create RGBA image from grayscale for text rendering
	rgbaImg := image.NewRGBA(img.Bounds())
	draw.Draw(rgbaImg, img.Bounds(), img, image.Point{}, draw.Src)

	// Add text to image
	fontFace := basicfont.Face7x13
	textColor := color.RGBA{255, 255, 255, 255} // White text

	// Position for text (top-left corner with padding)
	x := 10
	y := 20
	lineHeight := 15
	padding := 8

	// Calculate the maximum text width
	maxTextWidth := 0
	for _, line := range textLines {
		textWidth := len(line) * 7 // Approximate width for Face7x13 (7 pixels per character)
		if textWidth > maxTextWidth {
			maxTextWidth = textWidth
		}
	}

	// Draw background rectangle for text that properly surrounds all text
	rect := image.Rect(x-padding, y-lineHeight+padding, x+maxTextWidth+padding, y+(len(textLines)*lineHeight)+padding)

	// Semi-transparent black background
	bgColor := color.RGBA{0, 0, 0, 180}
	draw.Draw(rgbaImg, rect, &image.Uniform{bgColor}, image.Point{}, draw.Over)

	// Draw each line of text
	for i, line := range textLines {
		yPos := y + i*lineHeight
		e.drawText(rgbaImg, line, x, yPos, textColor, fontFace)
	}

	// Convert back to grayscale
	draw.Draw(img, img.Bounds(), rgbaImg, image.Point{}, draw.Src)

	return nil
}

// drawText draws text on an image
func (e *Exporter) drawText(img *image.RGBA, text string, x, y int, textColor color.RGBA, face font.Face) {
	drawer := &font.Drawer{
		Dst:  img,
		Src:  &image.Uniform{textColor},
		Face: face,
		Dot:  fixed.Point26_6{X: fixed.Int26_6(x << 6), Y: fixed.Int26_6(y << 6)},
	}
	drawer.DrawString(text)
}

// formatDate formats a DICOM date (YYYYMMDD) to a more readable format
func (e *Exporter) formatDate(dicomDate string) string {
	if len(dicomDate) != 8 {
		return dicomDate
	}

	year := dicomDate[:4]
	month := dicomDate[4:6]
	day := dicomDate[6:8]

	return fmt.Sprintf("%s/%s/%s", month, day, year)
}

// extractBodyPart extracts body part/anatomical region from study description
func (e *Exporter) extractBodyPart(studyDescription string) string {
	// Simple extraction based on common keywords in study descriptions
	description := strings.ToLower(studyDescription)

	bodyParts := map[string]string{
		"chest":    "CHEST",
		"abdomen":  "ABDOMEN",
		"pelvis":   "PELVIS",
		"head":     "HEAD",
		"brain":    "BRAIN",
		"spine":    "SPINE",
		"heart":    "HEART",
		"lung":     "LUNG",
		"liver":    "LIVER",
		"kidney":   "KIDNEY",
		"knee":     "KNEE",
		"shoulder": "SHOULDER",
		"ankle":    "ANKLE",
		"wrist":    "WRIST",
		"hip":      "HIP",
		"neck":     "NECK",
	}

	for keyword, bodyPart := range bodyParts {
		if strings.Contains(description, keyword) {
			return bodyPart
		}
	}

	// Default if no specific body part found
	return "UNKNOWN"
}

// addImageToPDF adds a PNG image to the PDF
func (e *Exporter) addImageToPDF(pdf *gofpdf.Fpdf, imagePath string) error {
	// Register the PNG image in the PDF
	imageInfo := pdf.RegisterImage(imagePath, "")
	if imageInfo == nil {
		return fmt.Errorf("failed to register image: %s", imagePath)
	}

	// Calculate image dimensions to fit on page (max width 180mm, maintain aspect ratio)
	maxWidth := 180.0
	imageWidth := imageInfo.Width()
	imageHeight := imageInfo.Height()

	// Scale image to fit page width while maintaining aspect ratio
	scale := maxWidth / imageWidth
	if scale > 1.0 {
		scale = 1.0 // Don't scale up
	}

	scaledWidth := imageWidth * scale
	scaledHeight := imageHeight * scale

	// Center the image horizontally
	x := (210.0 - scaledWidth) / 2 // A4 width is 210mm

	// Add image to PDF
	pdf.Image(imagePath, x, pdf.GetY(), scaledWidth, scaledHeight, false, "", 0, "")

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
		// Add new page for each image to give them proper space
		pdf.AddPage()

		// Add image title
		pdf.SetFont("Arial", "B", 10)
		pdf.Cell(40, 6, fmt.Sprintf("Image %d: %s", i+1, filepath.Base(imagePath)))
		pdf.Ln(8)

		// Add the actual PNG image to PDF
		if err := e.addImageToPDF(pdf, imagePath); err != nil {
			logrus.Warnf("Failed to add image %s to PDF: %v", imagePath, err)
			pdf.SetFont("Arial", "I", 10)
			pdf.Cell(40, 6, "[Failed to load image]")
			pdf.Ln(10)
		}

		pdf.Ln(10)
	}

	// Save PDF
	if err := pdf.OutputFileAndClose(outputPath); err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	logrus.Infof("Created PDF report: %s", outputPath)
	return nil
}

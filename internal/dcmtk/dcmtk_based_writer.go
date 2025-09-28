package dcmtk

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// DCMTKBasedWriter creates DICOM files using DCMTK's img2dcm tool
type DCMTKBasedWriter struct {
	tempDir string
}

// NewDCMTKBasedWriter creates a new DCMTK-based writer
func NewDCMTKBasedWriter() *DCMTKBasedWriter {
	return &DCMTKBasedWriter{
		tempDir: "/tmp/crgodicom",
	}
}

// WriteDICOMFile creates a DICOM file using DCMTK's img2dcm tool
func (w *DCMTKBasedWriter) WriteDICOMFile(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality string,
	width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel int,
	photometricInterpretation string, pixelData []byte) error {

	// Create temp directory if it doesn't exist
	if err := os.MkdirAll(w.tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Step 1: Convert pixel data to image format
	imagePath, err := w.convertPixelDataToImage(pixelData, width, height, bitsAllocated)
	if err != nil {
		return fmt.Errorf("failed to convert pixel data to image: %w", err)
	}
	defer os.Remove(imagePath) // Clean up temp image file

	// Step 2: Use img2dcm to create DICOM file
	if err := w.createDICOMWithImg2dcm(imagePath, filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality, width, height); err != nil {
		return fmt.Errorf("failed to create DICOM with img2dcm: %w", err)
	}

	return nil
}

// convertPixelDataToImage converts raw pixel data to a standard image format
func (w *DCMTKBasedWriter) convertPixelDataToImage(pixelData []byte, width, height, bitsAllocated int) (string, error) {
	// Create image from pixel data
	img, err := w.createImageFromPixelData(pixelData, width, height, bitsAllocated)
	if err != nil {
		return "", fmt.Errorf("failed to create image from pixel data: %w", err)
	}

	// Save as BMP (lossless format for better image quality)
	imagePath := filepath.Join(w.tempDir, "temp_image.bmp")
	file, err := os.Create(imagePath)
	if err != nil {
		return "", fmt.Errorf("failed to create image file: %w", err)
	}
	defer file.Close()

	// Convert to RGB for BMP encoding
	rgbaImg := image.NewRGBA(img.Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			gray := img.GrayAt(x, y)
			rgbaImg.Set(x, y, color.RGBA{gray.Y, gray.Y, gray.Y, 255})
		}
	}

	// Encode as BMP (lossless)
	if err := encodeBMP(file, rgbaImg); err != nil {
		return "", fmt.Errorf("failed to encode BMP: %w", err)
	}

	return imagePath, nil
}

// createImageFromPixelData creates a grayscale image from raw pixel data
func (w *DCMTKBasedWriter) createImageFromPixelData(pixelData []byte, width, height, bitsAllocated int) (*image.Gray, error) {
	// Create grayscale image
	grayImage := image.NewGray(image.Rect(0, 0, width, height))

	// Convert pixel data to grayscale image
	bytesPerPixel := bitsAllocated / 8
	if bitsAllocated%8 != 0 {
		bytesPerPixel++
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * bytesPerPixel

			if idx+bytesPerPixel <= len(pixelData) {
				var pixelValue uint8
				if bytesPerPixel == 2 {
					// 16-bit to 8-bit conversion (little-endian)
					lowByte := pixelData[idx]
					highByte := pixelData[idx+1]
					value16 := uint16(lowByte) | (uint16(highByte) << 8)
					// Scale 16-bit value (0-65535) to 8-bit (0-255)
					pixelValue = uint8(value16 >> 8)
				} else {
					pixelValue = pixelData[idx]
				}

				grayImage.SetGray(x, y, color.Gray{Y: pixelValue})
			}
		}
	}

	return grayImage, nil
}

// getImageDimensions extracts width and height from an image file
func (w *DCMTKBasedWriter) getImageDimensions(imagePath string) (int, int, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to open image file: %w", err)
	}
	defer file.Close()

	// Decode image to get dimensions
	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	return img.Width, img.Height, nil
}

// encodeBMP encodes an RGBA image as BMP format
func encodeBMP(w *os.File, img *image.RGBA) error {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	
	// BMP header
	fileSize := 54 + width*height*3 // Header + pixel data
	header := make([]byte, 54)
	
	// BMP file header (14 bytes)
	copy(header[0:2], "BM")                    // Signature
	binary.LittleEndian.PutUint32(header[2:6], uint32(fileSize))  // File size
	binary.LittleEndian.PutUint32(header[6:10], 0)                // Reserved
	binary.LittleEndian.PutUint32(header[10:14], 54)              // Data offset
	
	// BMP info header (40 bytes)
	binary.LittleEndian.PutUint32(header[14:18], 40)              // Header size
	binary.LittleEndian.PutUint32(header[18:22], uint32(width))   // Width
	binary.LittleEndian.PutUint32(header[22:26], uint32(height))   // Height
	binary.LittleEndian.PutUint16(header[26:28], 1)               // Planes
	binary.LittleEndian.PutUint16(header[28:30], 24)              // Bits per pixel
	binary.LittleEndian.PutUint32(header[30:34], 0)               // Compression
	binary.LittleEndian.PutUint32(header[34:38], uint32(width*height*3)) // Image size
	binary.LittleEndian.PutUint32(header[38:42], 0)               // X pixels per meter
	binary.LittleEndian.PutUint32(header[42:46], 0)               // Y pixels per meter
	binary.LittleEndian.PutUint32(header[46:50], 0)               // Colors used
	binary.LittleEndian.PutUint32(header[50:54], 0)               // Important colors
	
	// Write header
	if _, err := w.Write(header); err != nil {
		return err
	}
	
	// Write pixel data (BMP is bottom-up, so we need to flip)
	rowSize := ((width * 3 + 3) / 4) * 4 // Row size must be multiple of 4
	for y := height - 1; y >= 0; y-- {
		row := make([]byte, rowSize)
		for x := 0; x < width; x++ {
			c := img.RGBAAt(x, y)
			offset := x * 3
			row[offset] = c.B     // Blue
			row[offset+1] = c.G   // Green
			row[offset+2] = c.R   // Red
		}
		if _, err := w.Write(row); err != nil {
			return err
		}
	}
	
	return nil
}

// createDICOMWithImg2dcm uses DCMTK's img2dcm to create a DICOM file
func (w *DCMTKBasedWriter) createDICOMWithImg2dcm(imagePath, outputPath, patientName, patientID, studyUID, seriesUID, instanceUID, modality string, width, height int) error {
	// Build img2dcm command with DICOM attributes
	cmd := exec.Command("img2dcm", "--input-format", "BMP", imagePath, outputPath)

	// Add DICOM attributes as key-value pairs
	now := time.Now()
	attributes := []string{
		"0010,0010=" + patientName,                     // Patient Name
		"0010,0020=" + patientID,                       // Patient ID
		"0008,0060=" + modality,                        // Modality
		"0020,000D=" + studyUID,                        // Study Instance UID
		"0020,000E=" + seriesUID,                       // Series Instance UID
		"0008,0018=" + instanceUID,                     // SOP Instance UID
		"0008,0016=1.2.840.10008.5.1.4.1.1.2",          // SOP Class UID (CT Image Storage)
		"0008,0020=" + now.Format("20060102"),          // Study Date
		"0008,0030=" + now.Format("150405"),            // Study Time
		"0008,0050=" + now.Format("20060102") + "-001", // Accession Number
		"0008,1030=Generated Study",                    // Study Description
		"0008,103E=" + modality + " Series 1",          // Series Description
		"0010,0030=20000101",                           // Patient Birth Date
		"0010,0040=O",                                  // Patient Sex
		"0020,0011=1",                                  // Series Number
		"0020,0013=1",                                  // Instance Number
		"0028,0002=1",                                  // Samples Per Pixel
		"0028,0004=MONOCHROME2",                        // Photometric Interpretation
		"0028,0100=16",                                 // Bits Allocated
		"0028,0101=16",                                 // Bits Stored
		"0028,0102=15",                                 // High Bit
		"0028,0103=0",                                  // Pixel Representation (unsigned)
		"0028,1050=214",                                // Window Center
		"0028,1051=200",                                // Window Width
		"0028,1052=0",                                  // Rescale Intercept
		"0028,1053=1",                                  // Rescale Slope
		"0028,0008=1",                                  // Number of Frames
		"0028,0010=" + fmt.Sprintf("%d", height),       // Rows
		"0028,0011=" + fmt.Sprintf("%d", width),        // Columns
	}

	// Add all attributes as --key parameters
	for _, attr := range attributes {
		cmd.Args = append(cmd.Args, "--key", attr)
	}

	// Add options to create uncompressed DICOM
	cmd.Args = append(cmd.Args, "--write-file")          // Write file format (not just dataset)
	cmd.Args = append(cmd.Args, "--group-length-create") // Create group length elements

	// Execute img2dcm command
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("img2dcm failed: %s, output: %s", err, string(output))
	}

	return nil
}

// WriteDICOMFileWithValidation creates a DICOM file and validates it with DCMTK
func (w *DCMTKBasedWriter) WriteDICOMFileWithValidation(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality string,
	width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel int,
	photometricInterpretation string, pixelData []byte) error {

	// Write the DICOM file using img2dcm
	err := w.WriteDICOMFile(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality,
		width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel, photometricInterpretation, pixelData)
	if err != nil {
		return err
	}

	// Validate the file with DCMTK
	return ValidateDicomFile(filename)
}

// ValidateDicomFile validates a DICOM file using DCMTK tools
func ValidateDicomFile(filename string) error {
	return ValidateDicomFileWithVerbosity(filename, false)
}

// ValidateDicomFileWithVerbosity validates a DICOM file with optional verbose output
func ValidateDicomFileWithVerbosity(filename string, verbose bool) error {
	if verbose {
		fmt.Printf("🔍 Validating DICOM file: %s\n", filename)
	}

	// Run dcmdump to check if the file is readable
	cmd := exec.Command("dcmdump", filename)
	output, err := cmd.CombinedOutput()

	if err != nil {
		// If dcmdump fails, provide detailed error information
		fmt.Printf("❌ DCMTK validation FAILED for %s\n", filename)
		fmt.Printf("📋 Error details: %s\n", err.Error())
		fmt.Printf("📋 Command output:\n%s\n", string(output))

		// Try to provide more specific error information
		outputStr := string(output)
		if strings.Contains(outputStr, "Parse error") {
			fmt.Printf("💡 This appears to be a DICOM parsing error - the file structure may be invalid\n")
		} else if strings.Contains(outputStr, "Premature end of stream") {
			fmt.Printf("💡 The file appears to be truncated or incomplete\n")
		} else if strings.Contains(outputStr, "Unknown Tag") {
			fmt.Printf("💡 The file contains unknown or invalid DICOM tags\n")
		} else if strings.Contains(outputStr, "Group Length") {
			fmt.Printf("💡 There may be an issue with DICOM group length calculations\n")
		}

		return fmt.Errorf("DCMTK validation failed: %w", err)
	}

	if verbose {
		fmt.Printf("✅ DCMTK validation PASSED for %s\n", filename)
		// Show some key DICOM attributes
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "PatientName") || strings.Contains(line, "StudyInstanceUID") ||
				strings.Contains(line, "SeriesInstanceUID") || strings.Contains(line, "SOPInstanceUID") {
				fmt.Printf("📋 %s\n", strings.TrimSpace(line))
			}
		}
	}

	return nil
}

// TestDCMTKBasedWriter tests the DCMTK-based writer
func TestDCMTKBasedWriter() error {
	// Create test pixel data
	width, height := 512, 512
	pixelData := make([]byte, width*height*2) // 16-bit pixels

	// Generate test pattern
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			idx := (y*width + x) * 2
			// Create a gradient pattern
			value := uint16((x + y) * 2)
			pixelData[idx] = byte(value & 0xFF)
			pixelData[idx+1] = byte((value >> 8) & 0xFF)
		}
	}

	// Test the writer
	writer := NewDCMTKBasedWriter()
	return writer.WriteDICOMFileWithValidation(
		"test_dcmtk_based.dcm",
		"Test^Patient",
		"TEST001",
		"1.2.840.10008.5.1.4.1.1.2.123456789",
		"1.2.840.10008.5.1.4.1.1.2.987654321",
		"1.2.840.10008.5.1.4.1.1.2.111111111",
		"CT",
		512, 512, 16, 16, 15, 1,
		"MONOCHROME2",
		pixelData,
	)
}

package dcmtk

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// DicomMetadata represents extracted DICOM metadata
type DicomMetadata struct {
	PatientName               string
	PatientID                 string
	StudyUID                  string
	SeriesUID                 string
	InstanceUID               string
	Modality                  string
	StudyDate                 string
	StudyTime                 string
	StudyDescription          string
	SeriesDescription         string
	SOPClassUID               string
	Width                     int
	Height                    int
	BitsPerPixel              int
	SamplesPerPixel           int
	PhotometricInterpretation string
	PixelData                 []byte
}

// ReadDicomFile reads a DICOM file and extracts metadata using DCMTK's dcmdump
func ReadDicomFile(filename string) (*DicomMetadata, error) {
	// Use dcmdump to extract metadata from DICOM file
	cmd := exec.Command("dcmdump", filename)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("dcmdump failed: %s, output: %s", err, string(output))
	}

	// Parse the dcmdump output
	metadata := &DicomMetadata{}
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if err := parseDcmdumpLine(line, metadata); err != nil {
			// Continue parsing even if individual lines fail
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to parse dcmdump output: %w", err)
	}

	// Set defaults if not found
	if metadata.Width == 0 {
		metadata.Width = 512
	}
	if metadata.Height == 0 {
		metadata.Height = 512
	}
	if metadata.BitsPerPixel == 0 {
		metadata.BitsPerPixel = 16
	}
	if metadata.Modality == "" {
		metadata.Modality = "CT"
	}

	return metadata, nil
}

// parseDcmdumpLine parses a single line from dcmdump output
func parseDcmdumpLine(line string, metadata *DicomMetadata) error {
	// Parse lines like: (0010,0010) PN =Test^Patient                    #  10, 1 PatientName
	// or: (0020,000d) UI [1.2.840.10008.5.1.4.1.1.1758686675.7097468366877516101] #  54, 1 StudyInstanceUID
	re := regexp.MustCompile(`\(([0-9a-fA-F]{4}),([0-9a-fA-F]{4})\)\s+([A-Z]{2})\s+([^#]+)`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 5 {
		return fmt.Errorf("line doesn't match expected format")
	}

	group := matches[1]
	element := matches[2]
	_ = matches[3] // VR (Value Representation) - not used in this implementation
	value := strings.TrimSpace(matches[4])

	// Remove quotes, = sign, and square brackets from value
	value = strings.Trim(value, "=\"'[]")

	// Parse different DICOM tags
	tag := group + element
	switch tag {
	case "00100010": // Patient Name
		metadata.PatientName = value
	case "00100020": // Patient ID
		metadata.PatientID = value
	case "0020000d": // Study Instance UID
		metadata.StudyUID = value
	case "0020000e": // Series Instance UID
		metadata.SeriesUID = value
	case "00080018": // SOP Instance UID
		metadata.InstanceUID = value
	case "00080060": // Modality
		metadata.Modality = value
	case "00080020": // Study Date
		metadata.StudyDate = value
	case "00080030": // Study Time
		metadata.StudyTime = value
	case "00081030": // Study Description
		metadata.StudyDescription = value
	case "0008103e": // Series Description
		metadata.SeriesDescription = value
	case "00080016": // SOP Class UID
		metadata.SOPClassUID = value
	case "00280010": // Rows
		if height, err := strconv.Atoi(value); err == nil {
			metadata.Height = height
		}
	case "00280011": // Columns
		if width, err := strconv.Atoi(value); err == nil {
			metadata.Width = width
		}
	case "00280100": // Bits Allocated
		if bits, err := strconv.Atoi(value); err == nil {
			metadata.BitsPerPixel = bits
		}
	case "00280002": // Samples Per Pixel
		if samples, err := strconv.Atoi(value); err == nil {
			metadata.SamplesPerPixel = samples
		}
	case "00280004": // Photometric Interpretation
		metadata.PhotometricInterpretation = value
	}

	return nil
}

// TestDCMTKSimple tests basic DCMTK functionality
func TestDCMTKSimple() int {
	// Test if dcmdump is available
	cmd := exec.Command("dcmdump", "--version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0
	}
	return len(output)
}

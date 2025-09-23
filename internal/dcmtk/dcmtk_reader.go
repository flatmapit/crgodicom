package dcmtk

/*
#cgo pkg-config: dcmtk
#include "dcmtk_reader_wrapper.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
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

// ReadDicomFile reads a DICOM file and extracts metadata using DCMTK
func ReadDicomFile(filename string) (*DicomMetadata, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	// Allocate buffers for C strings
	cPatientName := (*C.char)(C.malloc(256))
	defer C.free(unsafe.Pointer(cPatientName))

	cPatientID := (*C.char)(C.malloc(256))
	defer C.free(unsafe.Pointer(cPatientID))

	cStudyUID := (*C.char)(C.malloc(256))
	defer C.free(unsafe.Pointer(cStudyUID))

	cSeriesUID := (*C.char)(C.malloc(256))
	defer C.free(unsafe.Pointer(cSeriesUID))

	cInstanceUID := (*C.char)(C.malloc(256))
	defer C.free(unsafe.Pointer(cInstanceUID))

	cModality := (*C.char)(C.malloc(16))
	defer C.free(unsafe.Pointer(cModality))

	var cWidth, cHeight, cBitsPerPixel C.int
	var cPixelData *C.uchar
	var cPixelDataLength C.int

	result := C.read_dicom_file_simple(
		cFilename,
		cPatientName,
		cPatientID,
		cStudyUID,
		cSeriesUID,
		cInstanceUID,
		cModality,
		&cWidth,
		&cHeight,
		&cBitsPerPixel,
		&cPixelData,
		&cPixelDataLength,
	)

	if result != 0 {
		return nil, fmt.Errorf("failed to read DICOM file: %s", filename)
	}

	metadata := &DicomMetadata{
		PatientName:               C.GoString(cPatientName),
		PatientID:                 C.GoString(cPatientID),
		StudyUID:                  C.GoString(cStudyUID),
		SeriesUID:                 C.GoString(cSeriesUID),
		InstanceUID:               C.GoString(cInstanceUID),
		Modality:                  C.GoString(cModality),
		Width:                     int(cWidth),
		Height:                    int(cHeight),
		BitsPerPixel:              int(cBitsPerPixel),
		SamplesPerPixel:           1,
		PhotometricInterpretation: "MONOCHROME2",
	}

	// Copy pixel data if available
	if cPixelDataLength > 0 {
		metadata.PixelData = C.GoBytes(unsafe.Pointer(cPixelData), cPixelDataLength)
	}

	return metadata, nil
}

// TestDCMTKSimple tests basic DCMTK functionality
func TestDCMTKSimple() int {
	result := C.test_dcmtk_simple()
	return int(result)
}

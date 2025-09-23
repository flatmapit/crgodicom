package dcmtk

/*
#cgo pkg-config: dcmtk
#include "dcmtk_writer_wrapper.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// WriteDicomFile creates a DICOM file using DCMTK with raw pixel data
func WriteDicomFile(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality string,
	width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel int,
	photometricInterpretation string, pixelData []byte) error {

	// Convert Go strings to C strings
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	cPatientName := C.CString(patientName)
	defer C.free(unsafe.Pointer(cPatientName))

	cPatientID := C.CString(patientID)
	defer C.free(unsafe.Pointer(cPatientID))

	cStudyUID := C.CString(studyUID)
	defer C.free(unsafe.Pointer(cStudyUID))

	cSeriesUID := C.CString(seriesUID)
	defer C.free(unsafe.Pointer(cSeriesUID))

	cInstanceUID := C.CString(instanceUID)
	defer C.free(unsafe.Pointer(cInstanceUID))

	cModality := C.CString(modality)
	defer C.free(unsafe.Pointer(cModality))

	cPhotometricInterpretation := C.CString(photometricInterpretation)
	defer C.free(unsafe.Pointer(cPhotometricInterpretation))

	// Call the C function
	result := C.write_dicom_file_simple(
		cFilename,
		cPatientName,
		cPatientID,
		cStudyUID,
		cSeriesUID,
		cInstanceUID,
		cModality,
		C.int(width),
		C.int(height),
		C.int(bitsAllocated),
		(*C.uchar)(unsafe.Pointer(&pixelData[0])),
		C.int(len(pixelData)),
	)

	if result != 0 {
		return fmt.Errorf("failed to write DICOM file")
	}

	return nil
}

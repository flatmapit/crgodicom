package dcmtk

/*
#cgo pkg-config: dcmtk
#include "dcmtk_network_wrapper.c"
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// Echo performs a C-ECHO operation to test connectivity
func Echo(host string, port int, callingAE, calledAE string) error {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	cCallingAE := C.CString(callingAE)
	defer C.free(unsafe.Pointer(cCallingAE))

	cCalledAE := C.CString(calledAE)
	defer C.free(unsafe.Pointer(cCalledAE))

	result := C.echo_test(cHost, C.int(port), cCallingAE, cCalledAE)
	if result != 0 {
		return fmt.Errorf("C-ECHO failed")
	}

	return nil
}

// Store performs a C-STORE operation to send a DICOM file
func Store(host string, port int, callingAE, calledAE, filename string) error {
	cHost := C.CString(host)
	defer C.free(unsafe.Pointer(cHost))

	cCallingAE := C.CString(callingAE)
	defer C.free(unsafe.Pointer(cCallingAE))

	cCalledAE := C.CString(calledAE)
	defer C.free(unsafe.Pointer(cCalledAE))

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	result := C.store_file(cHost, C.int(port), cCallingAE, cCalledAE, cFilename)
	if result != 0 {
		return fmt.Errorf("C-STORE failed")
	}

	return nil
}

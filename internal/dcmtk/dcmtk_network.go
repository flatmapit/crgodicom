package dcmtk

import (
	"fmt"
)

// Echo performs a C-ECHO operation to test connectivity
func Echo(host string, port int, callingAE, calledAE string) error {
	// Stub implementation - DCMTK networking not implemented yet
	fmt.Printf("DCMTK Echo stub: %s:%d %s->%s\n", host, port, callingAE, calledAE)
	return nil
}

// Store performs a C-STORE operation to send a DICOM file
func Store(host string, port int, callingAE, calledAE, filename string) error {
	// Stub implementation - DCMTK networking not implemented yet
	fmt.Printf("DCMTK Store stub: %s:%d %s->%s file=%s\n", host, port, callingAE, calledAE, filename)
	return nil
}

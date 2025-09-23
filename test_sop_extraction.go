package main

import (
	"fmt"
	"log"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
)

func main() {
	// Test SOP Instance UID extraction
	cfg := &config.Config{}
	reader := dicom.NewReader(cfg)

	testFile := "studies/1.2.840.10008.5.1.4.1.1.1758155906.5064502216790425427/series_001/image_001.dcm"

	fmt.Printf("🔍 Testing SOP Instance UID extraction from: %s\n", testFile)

	sopInstanceUID, err := reader.ExtractSOPInstanceUID(testFile)
	if err != nil {
		log.Printf("❌ Error extracting SOP Instance UID: %v", err)
		log.Printf("⚠️  This is expected if DICOM parsing is not fully implemented yet")
	} else {
		fmt.Printf("✅ Successfully extracted SOP Instance UID: %s\n", sopInstanceUID)
	}

	sopClassUID, err := reader.ExtractSOPClassUID(testFile)
	if err != nil {
		log.Printf("❌ Error extracting SOP Class UID: %v", err)
	} else {
		fmt.Printf("✅ Successfully extracted SOP Class UID: %s\n", sopClassUID)
	}
}





package main

import (
	"fmt"
	"log"
	"reflect"

	"github.com/flatmapit/crgodicom/internal/config"
	"github.com/flatmapit/crgodicom/internal/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

func main() {
	// Test what types of values the DICOM library actually returns
	cfg := &config.Config{}
	reader := dicom.NewReader(cfg)

	testFile := "studies/1.2.840.10008.5.1.4.1.1.1758155906.5064502216790425427/series_001/image_001.dcm"

	fmt.Printf("🔍 Testing DICOM value types from: %s\n", testFile)

	// Read the DICOM file
	dataset, err := reader.ReadDicomFile(testFile)
	if err != nil {
		log.Printf("❌ Error reading DICOM file: %v", err)
		return
	}

	// Test various tags to see what Value types we get
	testTags := []struct {
		Tag  tag.Tag
		Name string
	}{
		{tag.PatientName, "Patient Name"},
		{tag.PatientID, "Patient ID"},
		{tag.StudyDate, "Study Date"},
		{tag.StudyTime, "Study Time"},
		{tag.Modality, "Modality"},
		{tag.SeriesNumber, "Series Number"},
		{tag.InstanceNumber, "Instance Number"},
		{tag.Rows, "Rows"},
		{tag.Columns, "Columns"},
		{tag.BitsAllocated, "Bits Allocated"},
	}

	for _, test := range testTags {
		elem, err := dataset.FindElementByTag(test.Tag)
		if err != nil {
			fmt.Printf("❌ %s: Not found (%v)\n", test.Name, err)
			continue
		}

		if elem.Value == nil {
			fmt.Printf("⚠️  %s: Found but value is nil\n", test.Name)
			continue
		}

		valueType := reflect.TypeOf(elem.Value)
		fmt.Printf("✅ %s: Type=%v, Value=%v\n", test.Name, valueType, elem.Value)
	}
}





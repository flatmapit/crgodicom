package main

import (
	"fmt"
	"os"

	"github.com/suyashkumar/dicom"
	"github.com/suyashkumar/dicom/pkg/tag"
)

func main() {
	fmt.Println("Testing minimal DICOM element creation...")

	// Create a minimal dataset
	dataset := dicom.Dataset{
		Elements: make([]*dicom.Element, 0),
	}

	// Try creating a simple element
	if elem, err := dicom.NewElement(tag.PatientName, []string{"TEST^PATIENT"}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Patient Name element created successfully")
	} else {
		fmt.Printf("❌ Patient Name element failed: %v\n", err)
	}

	// Try creating a numeric element with different types
	fmt.Println("Testing different types for Rows element:")
	
	// Try int
	if elem, err := dicom.NewElement(tag.Rows, []int{512}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Rows element (int) created successfully")
	} else {
		fmt.Printf("❌ Rows element (int) failed: %v\n", err)
	}
	
	// Try uint32
	if elem, err := dicom.NewElement(tag.Rows, []uint32{512}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Rows element (uint32) created successfully")
	} else {
		fmt.Printf("❌ Rows element (uint32) failed: %v\n", err)
	}
	
	// Try string
	if elem, err := dicom.NewElement(tag.Rows, []string{"512"}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Rows element (string) created successfully")
	} else {
		fmt.Printf("❌ Rows element (string) failed: %v\n", err)
	}

	// Test other common elements
	fmt.Println("\nTesting other common elements:")
	
	// Test Series Number with different types
	fmt.Println("\nTesting Series Number element with different types:")
	
	// Test with int
	if elem, err := dicom.NewElement(tag.SeriesNumber, []int{1}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Series Number element (int) created successfully")
	} else {
		fmt.Printf("❌ Series Number element (int) failed: %v\n", err)
	}
	
	// Test with string
	if elem, err := dicom.NewElement(tag.SeriesNumber, []string{"1"}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Series Number element (string) created successfully")
	} else {
		fmt.Printf("❌ Series Number element (string) failed: %v\n", err)
	}
	
	// Test Instance Number
	if elem, err := dicom.NewElement(tag.InstanceNumber, []int{1}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Instance Number element created successfully")
	} else {
		fmt.Printf("❌ Instance Number element failed: %v\n", err)
	}
	
	// Test File Meta Information Group Length
	if elem, err := dicom.NewElement(tag.FileMetaInformationGroupLength, []int{0}); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ File Meta Information Group Length element created successfully")
	} else {
		fmt.Printf("❌ File Meta Information Group Length element failed: %v\n", err)
	}
	
	// Test Pixel Data with different types
	fmt.Println("\nTesting Pixel Data element:")
	
	// Test with []byte
	pixelData := make([]byte, 100)
	if elem, err := dicom.NewElement(tag.PixelData, pixelData); err == nil {
		dataset.Elements = append(dataset.Elements, elem)
		fmt.Println("✅ Pixel Data element ([]byte) created successfully")
	} else {
		fmt.Printf("❌ Pixel Data element ([]byte) failed: %v\n", err)
	}

	// Try writing to file
	file, err := os.Create("test_minimal.dcm")
	if err != nil {
		fmt.Printf("❌ Failed to create file: %v\n", err)
		return
	}
	defer file.Close()

	if err := dicom.Write(file, dataset); err != nil {
		fmt.Printf("❌ Failed to write DICOM file: %v\n", err)
	} else {
		fmt.Println("✅ DICOM file written successfully")
	}
}

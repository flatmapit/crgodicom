package dcmtk

// WriteDicomFile creates a DICOM file using DCMTK's img2dcm tool
func WriteDicomFile(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality string,
	width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel int,
	photometricInterpretation string, pixelData []byte) error {

	// Use DCMTK's img2dcm tool to create DICOM files
	writer := NewDCMTKBasedWriter()
	return writer.WriteDICOMFileWithValidation(filename, patientName, patientID, studyUID, seriesUID, instanceUID, modality,
		width, height, bitsAllocated, bitsStored, highBit, samplesPerPixel, photometricInterpretation, pixelData)
}

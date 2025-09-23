#include <stdlib.h>
#include <string.h>

// Reader functions
int read_dicom_file_simple(const char* filename, char* patient_name, char* patient_id, 
                          char* study_uid, char* series_uid, char* instance_uid, 
                          char* modality, int* width, int* height, int* bits_per_pixel,
                          unsigned char** pixel_data, int* pixel_data_length) {
    // For now, just return success with dummy data
    // TODO: Implement actual DCMTK reading
    strcpy(patient_name, "Test Patient");
    strcpy(patient_id, "TEST001");
    strcpy(study_uid, "1.2.3.4.5");
    strcpy(series_uid, "1.2.3.4.5.1");
    strcpy(instance_uid, "1.2.3.4.5.1.1");
    strcpy(modality, "CT");
    *width = 512;
    *height = 512;
    *bits_per_pixel = 16;
    *pixel_data = NULL;
    *pixel_data_length = 0;
    return 0;
}

// Test function
int test_dcmtk_simple(void) {
    return 42;
}

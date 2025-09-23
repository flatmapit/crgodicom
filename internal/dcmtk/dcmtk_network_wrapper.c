#include <stdlib.h>

// Network functions
int echo_test(const char* host, int port, const char* calling_ae, const char* called_ae) {
    // For now, just return success
    // TODO: Implement actual DCMTK C-ECHO
    return 0;
}

int store_file(const char* host, int port, const char* calling_ae, const char* called_ae, const char* filename) {
    // For now, just return success
    // TODO: Implement actual DCMTK C-STORE
    return 0;
}

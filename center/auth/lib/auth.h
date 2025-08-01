#ifndef IDB_AUTH_H
#define IDB_AUTH_H

#ifdef __cplusplus
extern "C" {
#endif

#include <stdint.h>
#include <time.h>

typedef enum {
    AUTH_MODE_LOCAL = 0,
    AUTH_MODE_REMOTE = 1
} AuthMode;

//
// Error code definitions
//
#define AUTH_OK                    0    // Success
#define AUTH_ERR_INVALID_PARAMS   -1    // Invalid parameters
#define AUTH_ERR_INVALID_SERIAL   -2    // Invalid serial number
#define AUTH_ERR_SERIAL_IO        -3    // Failed to read or write serial file
#define AUTH_ERR_MISMATCH         -4    // IP or serial number verification failed
#define AUTH_ERR_BIND_FAIL        -5    // Binding failed
#define AUTH_ERR_NOT_BOUND        -6    // Not bound
#define AUTH_ERR_EXPIRED          -7    // Authorization expired
#define AUTH_ERR_SIGN_MISMATCH    -8    // Signature mismatch

// Initialize authorization system (set authorization mode and working directory)
//   mode      : AUTH_MODE_LOCAL or AUTH_MODE_REMOTE
// Return value: AUTH_OK (0) or negative error code
int init_auth(AuthMode mode);

// Issue authorization serial number (based on IP)
//   ip        : Target server public IP
//   out       : Output buffer for serial number
//   out_size  : Size of output buffer, must be at least 32
// Return value: AUTH_OK (0) or negative error code
int issue_license(const char* ip, char* out, size_t out_size);

// Reissue authorization serial number (based on IP)
//   old_ip    : Old target server public IP
//   new_ip    : New target server public IP
//   old_serial: Old serial number
//   out   : Output buffer for new serial number
//   out_size  : Size of output buffer, must be at least 32
// Return value: AUTH_OK (0) or negative error code
int reissue_license(const char* old_ip, const char* new_ip,const char* old_serial, char* out, size_t out_size);


// Bind authorization (verify serial number and IP, generate .linked file)
//   ip        : Target server public IP
//   serial    : User-provided serial number
// Return value: AUTH_OK (0) or negative error code
int bind_license(const char* ip, const char* serial);

// Verify authorization (check .linked file exists and not expired)
//   ip        : Target server public IP
//   serial    : User-provided serial number
// Return value: AUTH_OK (0) or negative error code
int verify_license(const char* ip, const char* serial);

#ifdef __cplusplus
}
#endif

#endif // IDB_AUTH_H

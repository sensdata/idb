//go:build linux || darwin
// +build linux darwin

package auth

/*
#cgo CFLAGS: -I${SRCDIR}/lib -Wall
#cgo linux LDFLAGS: ${SRCDIR}/lib/libauth.a -lm -lcrypto
#cgo darwin LDFLAGS: ${SRCDIR}/lib/libauth_darwin.a -lm -lcrypto
#include <stdlib.h>
#include <stdio.h>
#include <stdint.h>
#include "auth.h"
*/
import "C"

import (
	"fmt"
)

// 定义与C枚举对应的Go类型
type AuthMode C.AuthMode

// 定义与C枚举对应的Go常量
const (
	AuthModeLocal  AuthMode = C.AUTH_MODE_LOCAL
	AuthModeRemote AuthMode = C.AUTH_MODE_REMOTE
)

// 定义错误码常量
const (
	AuthOK               = C.AUTH_OK
	AuthErrInvalidParams = C.AUTH_ERR_INVALID_PARAMS
	AuthErrInvalidSerial = C.AUTH_ERR_INVALID_SERIAL
	AuthErrSerialIO      = C.AUTH_ERR_SERIAL_IO
	AuthErrMismatch      = C.AUTH_ERR_MISMATCH
	AuthErrBindFail      = C.AUTH_ERR_BIND_FAIL
	AuthErrNotBound      = C.AUTH_ERR_NOT_BOUND
	AuthErrExpired       = C.AUTH_ERR_EXPIRED
	AuthErrSignMismatch  = C.AUTH_ERR_SIGN_MISMATCH
)

// InitAuth 初始化授权系统
func InitAuth(mode AuthMode) error {
	code := C.init_auth(C.AuthMode(mode))
	if code != AuthOK {
		return fmt.Errorf("auth initialization failed with code: %d", code)
	}
	return nil
}

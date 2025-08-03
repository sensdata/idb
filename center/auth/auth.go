//go:build linux || darwin

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
	"unsafe"
)

// IAuthService 定义授权服务接口
type IAuthService interface {
	// InitAuth 初始化授权系统
	InitAuth(mode AuthMode) error

	// IssueLicense 颁发授权序列号(基于IP)
	// ip: 目标服务器公网IP
	// 返回: 序列号字符串和可能的错误
	IssueLicense(ip string) (string, error)

	// ReissueLicense 重新颁发授权序列号
	// oldIp: 旧服务器公网IP
	// newIp: 新服务器公网IP
	// oldSerial: 旧序列号
	// 返回: 新序列号字符串和可能的错误
	ReissueLicense(oldIp, newIp, oldSerial string) (string, error)

	// BindLicense 绑定授权(验证序列号和IP，生成.linked文件)
	// ip: 目标服务器公网IP
	// serial: 用户提供的序列号
	// 返回: 可能的错误
	BindLicense(ip, serial string) error

	// VerifyLicense 验证授权(检查.linked文件存在且未过期)
	// ip: 目标服务器公网IP
	// serial: 用户提供的序列号
	// 返回: 可能的错误
	VerifyLicense(ip, serial string) error
}

// authService 实现IAuthService接口
type authService struct{}

// NewAuthService 创建一个新的授权服务实例
func NewAuthService() IAuthService {
	return &authService{}
}

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
func (s *authService) InitAuth(mode AuthMode) error {
	code := C.init_auth(C.AuthMode(mode))
	if code != AuthOK {
		return fmt.Errorf("auth initialization failed with code: %d", code)
	}
	return nil
}

// IssueLicense 颁发授权序列号(基于IP)
// ip: 目标服务器公网IP
// 返回: 序列号字符串和可能的错误
func (s *authService) IssueLicense(ip string) (string, error) {
	// 分配足够大的缓冲区，至少32字节
	buf := make([]byte, 64)
	cIp := C.CString(ip)
	defer C.free(unsafe.Pointer(cIp))

	code := C.issue_license(cIp, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
	if code != AuthOK {
		return "", fmt.Errorf("failed to issue license with code: %d", code)
	}

	// 转换C字符串为Go字符串
	return C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), nil
}

// ReissueLicense 重新颁发授权序列号
// oldIp: 旧服务器公网IP
// newIp: 新服务器公网IP
// oldSerial: 旧序列号
// 返回: 新序列号字符串和可能的错误
func (s *authService) ReissueLicense(oldIp, newIp, oldSerial string) (string, error) {
	buf := make([]byte, 64)
	cOldIp := C.CString(oldIp)
	cNewIp := C.CString(newIp)
	cOldSerial := C.CString(oldSerial)
	defer func() {
		C.free(unsafe.Pointer(cOldIp))
		C.free(unsafe.Pointer(cNewIp))
		C.free(unsafe.Pointer(cOldSerial))
	}()

	code := C.reissue_license(cOldIp, cNewIp, cOldSerial, (*C.char)(unsafe.Pointer(&buf[0])), C.size_t(len(buf)))
	if code != AuthOK {
		return "", fmt.Errorf("failed to reissue license with code: %d", code)
	}

	return C.GoString((*C.char)(unsafe.Pointer(&buf[0]))), nil
}

// BindLicense 绑定授权(验证序列号和IP，生成.linked文件)
// ip: 目标服务器公网IP
// serial: 用户提供的序列号
// 返回: 可能的错误
func (s *authService) BindLicense(ip, serial string) error {
	cIp := C.CString(ip)
	cSerial := C.CString(serial)
	defer func() {
		C.free(unsafe.Pointer(cIp))
		C.free(unsafe.Pointer(cSerial))
	}()

	code := C.bind_license(cIp, cSerial)
	if code != AuthOK {
		return fmt.Errorf("failed to bind license with code: %d", code)
	}

	return nil
}

// VerifyLicense 验证授权(检查.linked文件存在且未过期)
// ip: 目标服务器公网IP
// serial: 用户提供的序列号
// 返回: 可能的错误
func (s *authService) VerifyLicense(ip, serial string) error {
	cIp := C.CString(ip)
	cSerial := C.CString(serial)
	defer func() {
		C.free(unsafe.Pointer(cIp))
		C.free(unsafe.Pointer(cSerial))
	}()

	code := C.verify_license(cIp, cSerial)
	if code != AuthOK {
		return fmt.Errorf("license verification failed with code: %d", code)
	}

	return nil
}

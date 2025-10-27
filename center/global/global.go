package global

import (
	_ "embed"
	"sync"

	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/logstream"
	"github.com/sensdata/idb/core/model"

	"gorm.io/gorm"
)

var (
	Version    string = "0.0.1"
	Host       string = "127.0.0.1"
	DefaultKey string = ""

	LOG       *log.Log
	LogStream *logstream.LogStream
	DB        *gorm.DB
	VALID     *validator.Validate

	//go:embed certs/cert.pem
	CaCertPem []byte
	//go:embed certs/key.pem
	CaKeyPem []byte
	// self-sign cert & key
	CertPem []byte
	KeyPem  []byte

	// HostStatusCache 用于缓存host的状态信息
	HostStatusCache sync.Map
	// InstalledCache 用于缓存host上的agent安装状态
	InstalledCache sync.Map
)

func GetHostStatus(hostID uint) *model.HostStatusInfo {
	if val, ok := HostStatusCache.Load(hostID); ok {
		return val.(*model.HostStatusInfo)
	}
	return nil
}

func SetHostStatus(hostID uint, status *model.HostStatusInfo) {
	HostStatusCache.Store(hostID, status)
}

func DeleteHostStatus(hostID uint) {
	HostStatusCache.Delete(hostID)
}

func GetInstalledStatus(hostID uint) *string {
	if val, ok := InstalledCache.Load(hostID); ok {
		return val.(*string)
	}
	return nil
}

func SetInstalledStatus(hostID uint, status *string) {
	InstalledCache.Store(hostID, status)
}

func DeleteInstalledStatus(hostID uint) {
	InstalledCache.Delete(hostID)
}

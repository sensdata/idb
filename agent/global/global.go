package global

import (
	_ "embed"
	"sync/atomic"

	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"gorm.io/gorm"
)

var (
	Version string = "0.0.1"
	LOG     *log.Log
	DB      *gorm.DB
	//go:embed certs/key.pem
	KeyPem []byte
	//go:embed certs/cert.pem
	CertPem []byte
)

var License atomic.Value // 存储 *model.LicensePayload

func InitLicense() {
	License.Store((*model.LicensePayload)(nil)) // 初始化为 nil
}

func GetLicense() *model.LicensePayload {
	if v := License.Load(); v != nil {
		return v.(*model.LicensePayload)
	}
	return nil
}

func SetLicense(l *model.LicensePayload) {
	License.Store(l)
}

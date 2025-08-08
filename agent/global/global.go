package global

import (
	_ "embed"

	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/model"
	"gorm.io/gorm"
)

var (
	Version string = "0.0.1"
	LOG     *log.Log
	DB      *gorm.DB
	License *model.LicensePayload = &model.LicensePayload{}
	//go:embed certs/key.pem
	KeyPem []byte
	//go:embed certs/cert.pem
	CertPem []byte
)

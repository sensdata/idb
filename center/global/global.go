package global

import (
	_ "embed"

	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/core/log"
	"gorm.io/gorm"
)

var (
	Version string = "0.0.1"
	Host    string = "127.0.0.1"

	LOG   *log.Log
	DB    *gorm.DB
	VALID *validator.Validate

	//go:embed certs/key.pem
	KeyPem []byte
	//go:embed certs/cert.pem
	CertPem []byte
)

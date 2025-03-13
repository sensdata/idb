package global

import (
	_ "embed"

	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/logstream"

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

	//go:embed certs/key.pem
	KeyPem []byte
	//go:embed certs/cert.pem
	CertPem []byte
)

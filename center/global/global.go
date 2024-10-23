package global

import (
	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/core/log"
	"gorm.io/gorm"
)

var Version string = "0.0.1"
var Host string = "127.0.0.1"

var (
	LOG   *log.Log
	DB    *gorm.DB
	VALID *validator.Validate
)

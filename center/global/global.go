package global

import (
	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/core/log"
	"gorm.io/gorm"
)

var (
	CONF  config.CenterConfig
	LOG   *log.Log
	DB    *gorm.DB
	VALID *validator.Validate
)

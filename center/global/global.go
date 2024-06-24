package global

import (
	"github.com/go-playground/validator/v10"
	"github.com/sensdata/idb/center/config"
	"gorm.io/gorm"
)

var (
	CONF  config.CenterConfig
	DB    *gorm.DB
	VALID *validator.Validate
)

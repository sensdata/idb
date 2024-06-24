package global

import (
	"github.com/sensdata/idb/center/config"
	"gorm.io/gorm"
)

var (
	DB   *gorm.DB
	CONF config.CenterConfig
)

package global

import (
	"github.com/sensdata/idb/core/log"
	"gorm.io/gorm"
)

var (
	LOG *log.Log
	DB  *gorm.DB
)

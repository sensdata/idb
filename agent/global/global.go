package global

import (
	"github.com/sensdata/idb/core/log"
	"gorm.io/gorm"
)

var Version string = "0.0.1"

var (
	LOG *log.Log
	DB  *gorm.DB
)

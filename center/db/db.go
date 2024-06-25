package db

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dataSourceName string) {
	global.LOG.Info("db init begin")
	// Ensure the directory exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			global.LOG.Error("Failed to create directory: %v", err)
		}
	}

	//db logger
	gormLogger := log.NewGormLogger(global.LOG.Logger)
	dbLogger := logger.New(
		gormLogger,
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Silent,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	// Open (or create) the database
	db, err := gorm.Open(sqlite.Open(dataSourceName), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		Logger:                                   dbLogger,
	})
	if err != nil {
		global.LOG.Error("Failed to open database: %v", err)
		panic(err)
	}

	//Write-Ahead Logging
	_ = db.Exec("PRAGMA journal_mode = WAL;")
	//Settings
	sqlDB, dbError := db.DB()
	if dbError != nil {
		global.LOG.Error("Failed to open database: %v", err)
		panic(dbError)
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	global.LOG.Info("db init end")
}

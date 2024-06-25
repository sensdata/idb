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
	log.Info("db init begin")
	// Ensure the directory exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Error("Failed to create directory: %v", err)
		}
	}

	//db logger
	dbLogger := logger.New(
		log.Writer(),
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
		log.Error("Failed to open database: %v", err)
		panic(err)
	}

	//Write-Ahead Logging
	_ = db.Exec("PRAGMA journal_mode = WAL;")
	//Settings
	sqlDB, dbError := db.DB()
	if dbError != nil {
		log.Error("Failed to open database: %v", err)
		panic(dbError)
	}
	sqlDB.SetConnMaxIdleTime(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	global.DB = db
	log.Info("db init end")
}

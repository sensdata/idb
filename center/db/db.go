package db

import (
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/db/repo"
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

func InitSettings(cfg *config.CenterConfig) error {
	settingsRepo := repo.NewSettingsRepo()

	settings, err := settingsRepo.GetList()
	if err == nil || len(settings) > 0 {
		global.LOG.Info("settings already init")
		return nil
	}

	// 初始化 监听ip
	err = settingsRepo.Create("MonitorIP", "0.0.0.0")
	if err != nil {
		global.LOG.Error("Failed to init MonitorIP: %v", err)
		return err
	}

	// 初始化 port
	err = settingsRepo.Create("ServerPort", strconv.Itoa(cfg.Port))
	if err != nil {
		global.LOG.Error("Failed to init ServerPort: %v", err)
		return err
	}

	// 初始化 绑定域名
	err = settingsRepo.Create("BindDomain", "")
	if err != nil {
		global.LOG.Error("Failed to init BindDomain: %v", err)
		return err
	}

	// 初始化 Https
	err = settingsRepo.Create("Https", "no")
	if err != nil {
		global.LOG.Error("Failed to init Https: %v", err)
		return err
	}

	// 初始化 Https证书
	err = settingsRepo.Create("HttpsCertType", "default")
	if err != nil {
		global.LOG.Error("Failed to init HttpsCertType: %v", err)
		return err
	}
	err = settingsRepo.Create("HttpsCertPath", "")
	if err != nil {
		global.LOG.Error("Failed to init HttpsCertPath: %v", err)
		return err
	}
	err = settingsRepo.Create("HttpsKeyPath", "")
	if err != nil {
		global.LOG.Error("Failed to init HttpsCertPath: %v", err)
		return err
	}

	return nil
}

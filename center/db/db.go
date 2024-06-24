package db

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func Init(dataSourceName string) error {
	// Ensure the directory exists
	dir := filepath.Dir(dataSourceName)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Fatalf("Failed to create directory: %v", err)
		}
	}

	//db logger
	dbLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
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
		log.Fatalf("Failed to open database: %v", err)
		return err
	}
	global.DB = db

	// Initialize the schema and initial data
	err = initSchema(db)
	if err != nil {
		log.Fatalf("Error initializing schema: %v", err)
		return err
	}
	err = initRoles(db)
	if err != nil {
		log.Fatalf("Error initializing roles: %v", err)
		return err
	}
	err = initAdminUser(db)
	if err != nil {
		log.Fatalf("Error initializing admin: %v", err)
		return err
	}
	return nil
}

// CloseDB closes the database connection.
func CloseDB() error {
	if global.DB != nil {
		sqlDB, err := global.DB.DB()
		if err != nil {
			return err
		}
		return sqlDB.Close()
	}
	return nil
}

func initSchema(db *gorm.DB) error {
	err := db.AutoMigrate(&Role{}, &User{})
	if err != nil {
		log.Fatalf("Failed to migrate schema: %v", err)
		return err
	}
	return nil
}

func initRoles(db *gorm.DB) error {
	roles := []Role{
		{Name: "admin", Description: "Admin role"},
		{Name: "user", Description: "User role"},
	}

	for _, role := range roles {
		var existingRole Role
		if err := db.Where("name = ?", role.Name).First(&existingRole).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					log.Fatalf("Failed to insert role %s: %v", role.Name, err)
					return err
				}
			} else {
				log.Fatalf("Failed to check for role %s: %v", role.Name, err)
				return err
			}
		}
	}
	return nil
}

func initAdminUser(db *gorm.DB) error {
	var adminRole Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		log.Fatalf("Failed to get admin role ID: %v", err)
		return err
	}

	var count int64
	if err := db.Model(&User{}).Where("username = ?", "admin").Count(&count).Error; err != nil {
		log.Fatalf("Failed to check for admin user: %v", err)
		return err
	}

	if count == 0 {
		password := "admin123"
		salt := utils.GenerateNonce(8)
		passwordHash := utils.HashPassword(password, salt)
		adminUser := User{
			Username: "admin",
			Password: passwordHash,
			Salt:     salt,
			RoleID:   adminRole.ID,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			log.Fatalf("Failed to insert admin user: %v", err)
			return err
		}
	}
	return nil
}

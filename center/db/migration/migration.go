package migration

import (
	"log"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
	"gorm.io/gorm"
)

func Init() {
	gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		AddTableRole,
		AddTableUser,
	})
}

var AddTableRole = &gormigrate.Migration{
	ID: "20240624-add-table-role",
	Migrate: func(db *gorm.DB) error {
		if err := db.AutoMigrate(&model.Role{}); err != nil {
			return err
		}
		roles := []model.Role{
			{Name: "admin", Description: "Admin role"},
			{Name: "user", Description: "User role"},
		}
		for _, role := range roles {
			if err := db.Create(&role).Error; err != nil {
				log.Fatalf("Failed to insert role %s: %v", role.Name, err)
				return err
			}
		}
		return nil
	},
}

var AddTableUser = &gormigrate.Migration{
	ID: "20240624-add-table-user",
	Migrate: func(db *gorm.DB) error {
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return err
		}

		var adminRole model.Role
		if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			log.Fatalf("Failed to get admin role ID: %v", err)
			return err
		}

		password := "admin123"
		salt := utils.GenerateNonce(8)
		passwordHash := utils.HashPassword(password, salt)
		adminUser := model.User{
			Username: "admin",
			Password: passwordHash,
			Salt:     salt,
			RoleID:   adminRole.ID,
		}
		if err := db.Create(&adminUser).Error; err != nil {
			log.Fatalf("Failed to insert admin user: %v", err)
			return err
		}
		return nil
	},
}

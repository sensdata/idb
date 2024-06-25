package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
	"gorm.io/gorm"
)

func Init() {
	global.LOG.Info("db init begin")
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		AddTableRole,
		AddTableUser,
		AddTableGroup,
		AddFieldGroupIDToUser,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error("migration error: %v", err)
		panic(err)
	}
	global.LOG.Info("db init end")
}

var AddTableRole = &gormigrate.Migration{
	ID: "20240624-add-table-role",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Adding table Role")
		if err := db.AutoMigrate(&model.Role{}); err != nil {
			return err
		}
		roles := []model.Role{
			{Name: "admin", Description: "Admin role"},
			{Name: "user", Description: "User role"},
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, role := range roles {
				if err := tx.Create(&role).Error; err != nil {
					global.LOG.Error("Failed to insert role %s: %v", role.Name, err)
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table Role added successfully")
		return nil
	},
}

var AddTableUser = &gormigrate.Migration{
	ID: "20240624-add-table-user",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Adding table User")
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return err
		}

		var adminRole model.Role
		if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
			global.LOG.Error("Failed to get admin role ID: %v", err)
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

		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&adminUser).Error; err != nil {
				global.LOG.Error("Failed to insert admin user: %v", err)
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table User added successfully")
		return nil
	},
}

var AddTableGroup = &gormigrate.Migration{
	ID: "20240625-add-table-group",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Add table Group")
		if err := db.AutoMigrate(&model.Group{}); err != nil {
			return err
		}
		group := model.Group{
			GroupName: "default",
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&group).Error; err != nil {
				global.LOG.Error("Failed to insert group %s: %v", group.GroupName, err)
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table Group added successfully")
		return nil
	},
}

var AddFieldGroupIDToUser = &gormigrate.Migration{
	ID: "20240625-add-field-groupid-to-user",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Adding field GroupID to User table")

		// 增加 GroupID 字段
		if err := db.AutoMigrate(&model.User{}); err != nil {
			return err
		}

		// 查找 default 组的 ID
		var defaultGroup model.Group
		if err := db.Where("group_name = ?", "default").First(&defaultGroup).Error; err != nil {
			global.LOG.Error("Failed to get default group ID: %v", err)
			return err
		}

		// 更新所有用户记录，设置 GroupID 为 default 组的 ID
		if err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Model(&model.User{}).Where("group_id = 0").Update("group_id", defaultGroup.ID).Error; err != nil {
				global.LOG.Error("Failed to set default group id to all users: %v", err)
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table User added field GroupID successfully")

		return nil
	},
}

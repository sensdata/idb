package migration

import (
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sensdata/idb/center/db/model"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/utils"
	"gorm.io/gorm"
)

//go:embed timezones.json
var timezonesData []byte

func Init() {
	global.LOG.Info("db init begin")
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		AddTableRole,
		AddTableUser,
		AddTableGroup,
		AddFieldGroupIDToUser,
		AddTableHost,
		AddTableSettings,
		AddTableTimezone,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error("migration error: %v", err)
		panic(err)
	}
	global.LOG.Info("db init end")
}

func getTimezones() ([]model.Timezone, error) {
	timezones := make([]model.Timezone, 0)

	// 定义临时结构用于解析JSON
	type rawTimezone struct {
		Value  string   `json:"value"`
		Abbr   string   `json:"abbr"`
		Offset int      `json:"offset"`
		IsDst  bool     `json:"isdst"`
		Text   string   `json:"text"`
		UTC    []string `json:"utc"`
	}

	// 解析JSON数据
	var rawTimezones []rawTimezone
	if err := json.Unmarshal(timezonesData, &rawTimezones); err != nil {
		return nil, fmt.Errorf("failed to parse timezones data: %v", err)
	}

	// 遍历原始时区数据，将每个UTC值创建为独立的Timezone
	for _, raw := range rawTimezones {
		// 为每个UTC值创建一个新的Timezone实例
		for _, utc := range raw.UTC {
			timezone := model.Timezone{
				Value:  raw.Value,
				Abbr:   raw.Abbr,
				Offset: raw.Offset,
				IsDst:  raw.IsDst,
				Text:   raw.Text,
				UTC:    utc,
			}
			timezones = append(timezones, timezone)
		}
	}

	return timezones, nil
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
			Name:     "admin",
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

var AddTableHost = &gormigrate.Migration{
	ID: "20240627-add-table-host",
	Migrate: func(db *gorm.DB) error {

		global.LOG.Info("Adding table Host")

		if err := db.AutoMigrate(&model.HostGroup{}); err != nil {
			return err
		}
		if err := db.AutoMigrate(&model.Host{}); err != nil {
			return err
		}

		if err := db.Transaction(func(tx *gorm.DB) error {
			group := model.HostGroup{GroupName: "default"}
			if err := tx.Create(&group).Error; err != nil {
				global.LOG.Error("Failed to insert host group %s: %v", group.GroupName, err)
				return err
			}
			host := model.Host{
				IsDefault:    true,
				GroupID:      group.ID,
				Name:         "localhost",
				Addr:         global.Host,
				Port:         22,
				User:         "root",
				AuthMode:     "password",
				Password:     "",
				PrivateKey:   "",
				PassPhrase:   "",
				AgentAddr:    global.Host,
				AgentPort:    9919,
				AgentMode:    "https",
				AgentKey:     global.DefaultKey,
				AgentVersion: "",
			}
			if err := tx.Create(&host).Error; err != nil {
				global.LOG.Error("Failed to insert host %s: %v", host.Name, err)
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table Host added successfully")
		return nil
	},
}

var AddTableSettings = &gormigrate.Migration{
	ID: "20250107-add-table-settings",
	Migrate: func(db *gorm.DB) error {

		global.LOG.Info("Adding table Settings")

		if err := db.AutoMigrate(&model.Setting{}); err != nil {
			return err
		}

		global.LOG.Info("Table Settings added successfully")
		return nil
	},
}

var AddTableTimezone = &gormigrate.Migration{
	ID: "20250410-add-table-timezone",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Adding table Timezone")
		if err := db.AutoMigrate(&model.Timezone{}); err != nil {
			return err
		}

		timezones, err := getTimezones()
		if err != nil {
			return err
		}
		if err := db.Transaction(func(tx *gorm.DB) error {
			for _, timezone := range timezones {
				if err := tx.Create(&timezone).Error; err != nil {
					global.LOG.Error("Failed to insert timezone %s: %v", timezone.Value, err)
				}
			}
			return nil
		}); err != nil {
			return err
		}
		global.LOG.Info("Table Timezone added successfully")
		return nil
	},
}

package migration

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"github.com/sensdata/idb/agent/db/model"
	"github.com/sensdata/idb/agent/global"
	"gorm.io/gorm"
)

func Init() {
	global.LOG.Info("db init begin")
	m := gormigrate.New(global.DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		AddTableFavorite,
	})
	if err := m.Migrate(); err != nil {
		global.LOG.Error("migration error: %v", err)
		panic(err)
	}
	global.LOG.Info("db init end")
}

var AddTableFavorite = &gormigrate.Migration{
	ID: "20240826-add-table-favorite",
	Migrate: func(db *gorm.DB) error {
		global.LOG.Info("Adding table Favorite")
		if err := db.AutoMigrate(&model.Favorite{}); err != nil {
			return err
		}
		global.LOG.Info("Table Favorite added successfully")
		return nil
	},
}

package main

import (
	"fmt"

	"github.com/sensdata/idb/center/api"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core"
	"github.com/sensdata/idb/center/db"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/utils"
)

func main() {
	fmt.Printf("Center Starting")

	configPath := "config.json"

	manager, err := config.NewManager(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
	}

	cfg := manager.GetConfig()
	global.CONF = *cfg
	fmt.Println("Get config:")
	fmt.Printf("%+v \n", *cfg)

	//初始化日志模块
	if err := log.InitLogger(cfg.LogPath); err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
	}

	//初始化数据库
	if err := db.Init(cfg.DbPath); err != nil {
		fmt.Printf("Failed to initialize db: %v \n", err)
	}
	defer db.CloseDB()

	//启动服务
	center := core.NewCenter(*cfg)
	if err := center.Start(); err != nil {
		fmt.Printf("Failed to start center: %v", err)
	}
	defer center.Stop()

	//启动apiServer
	apiServer := api.NewApiServer(*cfg, center)
	if err := apiServer.Start(); err != nil {
		fmt.Printf("Failed to start api: %v", err)
	}

	// 等待信号
	utils.WaitForSignal()

	fmt.Println("Center Exited")
}

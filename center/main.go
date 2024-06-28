package main

import (
	"fmt"

	"github.com/sensdata/idb/center/api"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core"
	"github.com/sensdata/idb/center/db"
	"github.com/sensdata/idb/center/db/migration"
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
	log, err := log.InitLogger(cfg.LogPath)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
		panic(err)
	}
	global.LOG = log

	//初始化数据库
	db.Init(cfg.DbPath)
	migration.Init()

	//启动服务
	center := core.NewCenter(*cfg)
	if err := center.Start(); err != nil {
		fmt.Printf("Failed to start center: %v", err)
	}
	defer center.Stop()

	//启动SSH服务
	ssh := core.NewISSHService()
	if err := ssh.Start(); err != nil {
		fmt.Printf("Failed to start ssh: %v", err)
	}
	defer ssh.Stop()

	//启动apiServer
	apiServer := api.NewApiServer(*cfg)
	if err := apiServer.Start(); err != nil {
		fmt.Printf("Failed to start api: %v", err)
	}

	//初始化其他
	global.VALID = utils.InitValidator()

	// 等待信号
	utils.WaitForSignal()

	fmt.Println("Center Exited")
}

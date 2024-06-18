package main

import (
	"fmt"

	"github.com/sensdata/idb/agent/channel"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/utils"
)

func main() {
	fmt.Println("Agent Starting")

	configPath := "config.json"

	manager, err := config.NewManager(configPath)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
	}

	cfg := manager.GetConfig()
	fmt.Println("Get config:")
	fmt.Printf("%+v \n", *cfg)

	// 初始化日志模块
	if err := log.InitLogger(cfg.LogPath); err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
	}

	//启动服务
	agent := channel.NewAgent(*cfg)
	if err := agent.Start(); err != nil {
		fmt.Printf("Failed to start agent: %v", err)
	}
	defer agent.Stop()

	// 等待信号
	utils.WaitForSignal()

	fmt.Println("Agent Exited")
}

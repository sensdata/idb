package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sensdata/idb/agent/channel"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/core/log"
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
	service := channel.NewChannelService(*cfg)
	if err := service.Start(); err != nil {
		fmt.Printf("Failed to start channel service: %v", err)
	}
	defer service.Stop()

	// 等待信号
	waitForSignal()

	fmt.Println("Agent Exited")
}

func waitForSignal() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}

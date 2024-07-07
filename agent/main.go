package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/sensdata/idb/agent/agent"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	logger "github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/utils"
	"github.com/urfave/cli"
)

var app = &cli.App{
	Name:  "idbagent",
	Usage: "idb agent command line tools",
	Commands: []cli.Command{
		*agent.StopCommand,
		*agent.RestartCommand,
		*agent.ConfigCommand,
	},
}

func main() {
	// Open the log file
	logFile, err := os.OpenFile("/var/log/idb-agent.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer logFile.Close()

	// Set log output to the log file
	log.SetOutput(logFile)

	if len(os.Args) > 1 && os.Args[1] == "start" {
		err := Run()
		if err != nil {
			log.Printf("Error: %v", err)
		}
	} else {
		err := app.Run(os.Args)
		if err != nil {
			log.Printf("Error: %v", err)
		}
	}
}

func Run() error {
	// 判断pid文件是否存在
	pidfile := filepath.Join(constant.BaseDir, constant.AgentPid)
	running, err := utils.IsRunning(pidfile)
	if err != nil {
		return fmt.Errorf("agent error %v", err)
	}
	if running {
		return fmt.Errorf("agent running %v", err)
	}
	// 创建pid文件
	utils.CreatePIDFile(pidfile)

	// 启动各项服务
	if err := StartServices(); err != nil {
		return StopServices()
	}

	// 捕捉系统信号，保持运行
	utils.WaitForSignal()

	log.Println("Agent shutting down...")
	return StopServices()
}

func StartServices() error {
	// 检查目录
	if err := utils.EnsurePaths(constant.BaseDir); err != nil {
		log.Printf("Failed to initialize directories: %v \n", err)
		return err
	}

	// 初始化配置
	cfgFilePath := filepath.Join(constant.BaseDir, constant.AgentConfig)
	manager, err := config.NewManager(cfgFilePath)
	if err != nil {
		log.Printf("Failed to initialize config manager: %v \n", err)
		return err
	}
	agent.CONFMAN = manager

	// 初始化日志模块
	logger, err := logger.InitLogger(constant.BaseDir, constant.AgentLog)
	if err != nil {
		log.Printf("Failed to initialize logger: %v \n", err)
		return err
	}
	global.LOG = logger

	// 启动agent服务
	a := agent.NewAgent()
	if err := a.Start(); err != nil {
		log.Printf("Failed to start agent: %v", err)
		return err
	}
	agent.AGENT = a

	return nil
}

func StopServices() error {
	// 停止agent服务
	agent.AGENT.Stop()

	// 删除pid文件
	pidfile := filepath.Join(constant.BaseDir, constant.AgentPid)
	utils.RemovePIDFile(pidfile)

	return nil
}

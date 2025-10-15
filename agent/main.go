package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sensdata/idb/agent/agent"
	"github.com/sensdata/idb/agent/config"
	"github.com/sensdata/idb/agent/db"
	"github.com/sensdata/idb/agent/db/migration"
	"github.com/sensdata/idb/agent/global"
	"github.com/sensdata/idb/core/constant"
	logger "github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/utils"
	"github.com/urfave/cli"
)

var app = &cli.App{
	Name:    "idb-agent",
	Usage:   "idb-agent command line tools",
	Version: global.Version,
	Authors: []cli.Author{
		{
			Name:  "iDB Dev Team",
			Email: "idb@sensdata.com",
		},
	},
	Description: `idb-agent is a proxy service running on the local server 
	that maintains communication with the idb server and accepts instructions and control 
	from the idb server, such as executing shell commands, running scripts, 
	performing file operations, etc.
	The idb-agent command line tool provides the following features:
	- Control idb agent
	- Config idb agent
	- Update idb agent
	- Remove idb agent`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "version, v",
			Usage: "print the version",
		},
	},
	HideHelp:    false,
	HideVersion: false,
	Commands: []cli.Command{
		*agent.StopCommand,
		*agent.RestartCommand,
		*agent.StatusCommand,
		*agent.ConfigCommand,
		*agent.UpdateCommand,
		*agent.RemoveCommand,
	},
}

func main() {
	// 检查目录
	paths := []string{constant.AgentConfDir, constant.AgentDataDir, constant.AgentLogDir, constant.AgentRunDir}
	if err := utils.EnsurePaths(paths); err != nil {
		fmt.Printf("Agent directories error: %v", err)
		return
	}

	// 初始化日志模块
	if global.LOG == nil {
		logger, err := logger.InitLogger(constant.AgentLogDir, constant.AgentLog)
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v \n", err)
			return
		}
		global.LOG = logger
	}

	if len(os.Args) > 1 && os.Args[1] == "start" {
		err := Run()
		if err != nil {
			global.LOG.Error("Error start agent: %v", err)
		}
	} else {
		err := app.Run(os.Args)
		if err != nil {
			global.LOG.Error("Error run agent cmd: %v", err)
		}
	}
}

func Run() error {
	global.LOG.Info("Agent ver: %s", global.Version)

	// 启动各项服务
	if err := StartServices(); err != nil {
		return StopServices()
	}

	// 捕捉系统信号，保持运行
	utils.WaitForSignal()

	global.LOG.Info("Agent shutting down...")
	return StopServices()
}

func StartServices() error {
	// 初始化配置
	cfgFilePath := filepath.Join(constant.AgentConfDir, constant.AgentConfig)
	manager, err := config.NewManager(cfgFilePath)
	if err != nil {
		global.LOG.Error("Failed to initialize config manager: %v \n", err)
		return err
	}
	agent.CONFMAN = manager

	//初始化数据库
	global.LOG.Info("Init db")
	db.Init(filepath.Join(constant.AgentDataDir, constant.AgentDb))
	migration.Init()

	// 启动agent服务
	a := agent.NewAgent()
	if err := a.Start(); err != nil {
		global.LOG.Error("Failed to start agent: %v", err)
		return err
	}
	agent.AGENT = a

	return nil
}

func StopServices() error {
	// 停止agent服务
	agent.AGENT.Stop()

	// 最后关闭日志
	global.LOG.Close()

	return nil
}

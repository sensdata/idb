// @title IDB API Documentation
// @version 1.0
// @description This is the API documentation for idb application.
// @termsOfService https://static.sensdata.com/terms/

// @contact.name API Support
// @contact.url https://static.sensdata.com/support
// @contact.email support@sensdata.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host http://39.99.155.139:9918
// @BasePath /api/v1
// @schemes http

package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/command"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db"
	"github.com/sensdata/idb/center/db/migration"
	_ "github.com/sensdata/idb/center/docs"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/constant"
	logger "github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/logstream"
	"github.com/sensdata/idb/core/utils"
	"github.com/urfave/cli"
)

var app = &cli.App{
	Name:    "idb",
	Usage:   "idb command line tools",
	Version: global.Version,
	Authors: []cli.Author{
		{
			Name:  "iDB Dev Team",
			Email: "idb@sensdata.com",
		},
	},
	Description: `idb is a centralized management platform for remote servers.
	The idb command line tool provides the following features:
	- Config idb server
	- Update idb server`,
	Flags: []cli.Flag{
		cli.BoolFlag{
			Name:  "version, v",
			Usage: "print the version",
		},
	},
	HideHelp:    false,
	HideVersion: false,
	Commands: []cli.Command{
		*command.StatusCommand,
		*command.ConfigCommand,
		*command.UpdateCommand,
		*command.ResetPasswordCommand,
	},
}

func main() {
	// 设置 gin 模式为 release
	gin.SetMode(gin.ReleaseMode)

	// 检查目录
	paths := []string{constant.CenterConfDir, constant.CenterDataDir, constant.CenterAgentDir, constant.CenterLogDir, constant.CenterRunDir}
	if err := utils.EnsurePaths(paths); err != nil {
		fmt.Printf("center directories error: %v", err)
		return
	}

	//初始化日志模块
	if global.LOG == nil {
		logger, err := logger.InitLogger(constant.CenterLogDir, constant.CenterLog)
		if err != nil {
			fmt.Printf("Failed to initialize logger: %v \n", err)
			return
		}
		global.LOG = logger
	}

	if len(os.Args) > 1 && os.Args[1] == "start" {
		err := Run()
		if err != nil {
			global.LOG.Error("Error start center: %v", err)
		}
	} else {
		err := app.Run(os.Args)
		if err != nil {
			global.LOG.Error("Error run center cmd: %v", err)
		}
	}
}

func Run() error {
	global.LOG.Info("Center ver: %s", global.Version)

	// 启动各项服务
	if err := StartServices(); err != nil {
		return StopServices()
	}

	// 捕捉系统信号，保持运行
	utils.WaitForSignal()

	global.LOG.Info("Center shutting down...")
	return StopServices()
}

func StartServices() error {
	// 初始化配置
	cfgFilePath := filepath.Join(constant.CenterConfDir, constant.CenterConfig)
	manager, err := config.NewManager(cfgFilePath)
	if err != nil {
		global.LOG.Error("Failed to initialize config manager: %v \n", err)
		return err
	}
	conn.CONFMAN = manager
	global.Host = manager.GetConfig().Host

	// 初始化数据库
	global.LOG.Info("Init db")
	db.Init(filepath.Join(constant.CenterDataDir, constant.CenterDb))
	migration.Init()
	// 初始化设置
	db.InitSettings(conn.CONFMAN.GetConfig())

	// 初始化logstream
	ls, err := logstream.New(nil)
	if err != nil {
		global.LOG.Error("Failed to initialize logstream: %v", err)
		return err
	}
	global.LogStream = ls

	// 启动SSH服务
	global.LOG.Info("Init ssh")
	ssh := conn.NewSSHService()
	if err := ssh.Start(); err != nil {
		global.LOG.Error("Failed to start ssh: %v", err)
		return err
	}
	conn.SSH = ssh

	// 启动WS服务
	global.LOG.Info("Init ws")
	conn.WEBSOCKET = conn.NewWebSocketService()

	// 启动center服务
	global.LOG.Info("Init center")
	center := conn.NewCenter()
	if err := center.Start(); err != nil {
		global.LOG.Error("Failed to start center: %v", err)
		return err
	}
	conn.CENTER = center

	// 初始化路由
	global.LOG.Info("Init api")
	api.API.InitRouter()
	// 注册插件
	// plugin.RegisterPlugins()
	// 启动apiServer
	if err := api.API.Start(); err != nil {
		global.LOG.Error("Failed to start api: %v", err)
		return err
	}
	// 启动插件
	// plugin.StartPlugins()

	return nil
}

func StopServices() error {
	// 关闭 logstream
	if global.LogStream != nil {
		global.LogStream.Close()
	}

	// 停止 API 服务器
	if err := api.API.Stop(); err != nil {
		global.LOG.Error("停止 API 服务器失败: %v", err)
	}

	// 停止Agent服务
	if err := conn.CENTER.Stop(); err != nil {
		global.LOG.Error("停止Agent服务失败: %v", err)
	}

	// 停止SSH
	conn.SSH.Stop()

	return nil
}

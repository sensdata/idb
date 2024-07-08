package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

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
	"github.com/sensdata/idb/core/utils"
	"github.com/urfave/cli"
)

var app = &cli.App{
	Name:  "idbcenter",
	Usage: "idb center command line tools",
	Commands: []cli.Command{
		*command.StopCommand,
		*command.RestartCommand,
		*command.ConfigCommand,
	},
}

func main() {
	// Open the log file
	logFile, err := os.OpenFile("/var/log/idb/center-run.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
	// 检查目录
	paths := []string{constant.ConfDir, constant.DataDir, constant.LogDir}
	if err := utils.EnsurePaths(paths); err != nil {
		return fmt.Errorf("center directories error: %v", err)
	}

	// 判断pid文件是否存在
	pidfile := filepath.Join(constant.DataDir, constant.CenterPid)
	running, err := utils.IsRunning(pidfile)
	if err != nil {
		return fmt.Errorf("center error %v", err)
	}
	if running {
		return fmt.Errorf("center running %v", err)
	}
	// 创建pid文件
	utils.CreatePIDFile(pidfile)

	// 启动各项服务
	if err := StartServices(); err != nil {
		return StopServices()
	}

	// 捕捉系统信号，保持运行
	utils.WaitForSignal()

	log.Println("Center shutting down...")
	return StopServices()
}

func StartServices() error {
	// 初始化配置
	cfgFilePath := filepath.Join(constant.ConfDir, constant.CenterConfig)
	manager, err := config.NewManager(cfgFilePath)
	if err != nil {
		log.Printf("Failed to initialize config manager: %v \n", err)
		return err
	}
	conn.CONFMAN = manager

	//初始化日志模块
	logger, err := logger.InitLogger(constant.LogDir, constant.CenterLog)
	if err != nil {
		log.Printf("Failed to initialize logger: %v \n", err)
		return err
	}
	global.LOG = logger

	//初始化数据库
	db.Init(filepath.Join(constant.DataDir, constant.DBFile))
	migration.Init()

	//启动apiServer
	apiServer := api.NewApiServer()
	if err := apiServer.Start(); err != nil {
		log.Printf("Failed to start api: %v", err)
		return err
	}

	// 启动SSH服务
	ssh := conn.NewSSHService()
	if err := ssh.Start(); err != nil {
		log.Printf("Failed to start ssh: %v", err)
		return err
	}
	conn.SSH = ssh

	// 启动center服务
	center := conn.NewCenter()
	if err := center.Start(); err != nil {
		log.Printf("Failed to start center: %v", err)
		return err
	}
	conn.CENTER = center

	//初始化其他
	global.VALID = utils.InitValidator()

	return nil
}

func StopServices() error {
	// 停止Agent服务
	conn.CENTER.Stop()

	// 停止SSH
	conn.SSH.Stop()

	// 删除pid文件
	pidfile := filepath.Join(constant.DataDir, constant.CenterPid)
	utils.RemovePIDFile(pidfile)

	return nil
}

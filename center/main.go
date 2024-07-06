package main

import (
	"fmt"
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
	"github.com/sensdata/idb/core/log"
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
	if len(os.Args) > 1 && os.Args[1] == "start" {
		err := Run()
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
	} else {
		err := app.Run(os.Args)
		if err != nil {
			fmt.Printf("Error: %v", err)
		}
	}
}

func Run() error {
	pidfile := filepath.Join(constant.BaseDir, constant.CenterPid)
	running, err := utils.IsRunning(pidfile)
	if running || err != nil {
		return fmt.Errorf("center running or error %v", err)
	}

	if err := StartServices(); err != nil {
		return StopServices()
	}

	// 捕捉系统信号，保持运行
	utils.WaitForSignal()

	fmt.Println("Center shutting down...")
	return StopServices()
}

func StartServices() error {

	// 检查目录
	if err := utils.EnsurePaths(constant.BaseDir); err != nil {
		fmt.Printf("Failed to initialize directories: %v \n", err)
		return err
	}

	// 初始化配置
	cfgFilePath := filepath.Join(constant.BaseDir, constant.CenterConfig)
	manager, err := config.NewManager(cfgFilePath)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
		return err
	}
	conn.CONFMAN = manager

	//初始化日志模块
	log, err := log.InitLogger(constant.BaseDir, constant.CenterLog)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
		return err
	}
	global.LOG = log

	//初始化数据库
	db.Init(filepath.Join(constant.BaseDir, constant.DBFile))
	migration.Init()

	//启动apiServer
	apiServer := api.NewApiServer()
	if err := apiServer.Start(); err != nil {
		fmt.Printf("Failed to start api: %v", err)
		return err
	}

	// 启动SSH服务
	ssh := conn.NewSSHService()
	if err := ssh.Start(); err != nil {
		fmt.Printf("Failed to start ssh: %v", err)
		return err
	}
	conn.SSH = ssh

	// 启动center服务
	center := conn.NewCenter()
	if err := center.Start(); err != nil {
		fmt.Printf("Failed to start center: %v", err)
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

	return nil
}

package command

import (
	"flag"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/sensdata/idb/center/config"
	"github.com/sensdata/idb/center/core/api"
	"github.com/sensdata/idb/center/core/conn"
	"github.com/sensdata/idb/center/db"
	"github.com/sensdata/idb/center/db/migration"
	"github.com/sensdata/idb/center/global"
	"github.com/sensdata/idb/core/log"
	"github.com/sensdata/idb/core/utils"
	"github.com/urfave/cli"
)

var StartCommand = &cli.Command{
	Name:  "start",
	Usage: "start center",
	Action: func(c *cli.Context) error {
		//已经启动了，退出
		if global.STARTED {
			return nil
		}

		if err := StartServices(); err != nil {
			return StopServices()
		}

		// 捕捉系统信号，保持运行
		utils.WaitForSignal()

		global.LOG.Info("Shutting down center...")
		return StopServices()
	},
}

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop center",
	Action: func(c *cli.Context) error {
		conn, err := net.Dial("unix", "/tmp/idb-center.sock")
		if err != nil {
			return fmt.Errorf("failed to connect to center: %w", err)
		}
		defer conn.Close()

		_, err = conn.Write([]byte("stop"))
		if err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		fmt.Println(string(buf[:n]))
		return nil
	},
}

var RestartCommand = &cli.Command{
	Name:  "restart",
	Usage: "restart center",
	Action: func(c *cli.Context) error {
		// 调用 StopCommand
		conn, err := net.Dial("unix", "/tmp/idb-center.sock")
		if err != nil {
			return fmt.Errorf("failed to connect to center: %w", err)
		}
		defer conn.Close()

		_, err = conn.Write([]byte("stop"))
		if err != nil {
			return fmt.Errorf("failed to send stop command: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read stop response: %w", err)
		}

		fmt.Println(string(buf[:n]))

		// 确保Agent停止后再继续
		time.Sleep(2 * time.Second)

		// 创建一个新的cli.Context
		flagSet := flag.NewFlagSet("start", flag.ContinueOnError)
		startCtx := cli.NewContext(c.App, flagSet, c)

		err = StartCommand.Run(startCtx)
		if err != nil {
			return fmt.Errorf("failed to start center: %w", err)
		}

		return nil
	},
}

func StartServices() error {

	// 获取当前可执行文件的路径
	ex, err := os.Executable()
	if err != nil {
		fmt.Printf("Failed to get executable path: %v\n", err)
		return err
	}

	// 获取安装目录
	installDir := filepath.Dir(ex)

	// 初始化配置
	manager, err := config.NewManager(installDir)
	if err != nil {
		fmt.Printf("Failed to initialize config manager: %v \n", err)
		return err
	}
	conn.CONFMAN = manager

	//初始化日志模块
	log, err := log.InitLogger(installDir, "config.json")
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v \n", err)
		return err
	}
	global.LOG = log

	//初始化数据库
	db.Init(filepath.Join(installDir, "idb.db"))
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
	global.STARTED = true

	// 等待信号
	utils.WaitForSignal()
	return nil
}

func StopServices() error {
	// 停止Agent服务
	conn.CENTER.Stop()

	// 停止SSH
	conn.SSH.Stop()

	return nil
}

var ConfigCommand = &cli.Command{
	Name:  "config",
	Usage: "configure center",
	Flags: []cli.Flag{
		&cli.StringFlag{
			Name:  "key",
			Usage: "configuration key",
		},
		&cli.StringFlag{
			Name:  "value",
			Usage: "configuration value",
		},
	},
	Action: func(c *cli.Context) error {
		args := c.Args()

		var key, value string

		if len(args) > 0 {
			key = args.Get(0)
		}
		if len(args) > 1 {
			value = args.Get(1)
		}

		conn, err := net.Dial("unix", "/tmp/idb-center.sock")
		if err != nil {
			return fmt.Errorf("failed to connect to center: %w", err)
		}
		defer conn.Close()

		var configCommand string
		if key == "" {
			configCommand = "config"
		} else if value == "" {
			configCommand = fmt.Sprintf("config %s", key)
		} else {
			configCommand = fmt.Sprintf("config %s %s", key, value)
		}

		_, err = conn.Write([]byte(configCommand))
		if err != nil {
			return fmt.Errorf("failed to send command: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			return fmt.Errorf("failed to read response: %w", err)
		}

		fmt.Println(string(buf[:n]))
		return nil
	},
}

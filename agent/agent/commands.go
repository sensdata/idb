package agent

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/sensdata/idb/agent/global"
	"github.com/urfave/cli"
)

var StartCommand = &cli.Command{
	Name:  "start",
	Usage: "start agent",
	Action: func(c *cli.Context) error {
		//已经启动了，退出
		if AGENT.Started() {
			return nil
		}

		// 启动Agent服务
		err := AGENT.Start()
		if err != nil {
			return err
		}

		// 捕捉系统信号，保持运行
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
		<-sigs

		global.LOG.Info("Shutting down agent...")
		return AGENT.Stop()
	},
}

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop agent",
	Action: func(c *cli.Context) error {
		conn, err := net.Dial("unix", "/tmp/idb-agent.sock")
		if err != nil {
			return fmt.Errorf("failed to connect to agent: %w", err)
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

var ConfigCommand = &cli.Command{
	Name:  "config",
	Usage: "configure agent",
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

		fmt.Printf("config %s %s\n", key, value)

		conn, err := net.Dial("unix", "/tmp/idb-agent.sock")
		if err != nil {
			return fmt.Errorf("failed to connect to agent: %w", err)
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

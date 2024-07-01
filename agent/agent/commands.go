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
			global.LOG.Error("failed to connect to agent: %v", err)
			return fmt.Errorf("failed to connect to agent: %w", err)
		}
		defer conn.Close()

		_, err = conn.Write([]byte("stop"))
		if err != nil {
			global.LOG.Error("failed to send command: %v", err)
			return fmt.Errorf("failed to send command: %w", err)
		}

		buf := make([]byte, 1024)
		n, err := conn.Read(buf)
		if err != nil {
			global.LOG.Error("failed to read response: %v", err)
			return fmt.Errorf("failed to read response: %w", err)
		}

		global.LOG.Info(string(buf[:n]))
		return nil
	},
}

package agent

import (
	"fmt"
	"net"
	"os/exec"
	"path/filepath"

	"github.com/sensdata/idb/core/constant"
	"github.com/urfave/cli"
)

var StopCommand = &cli.Command{
	Name:  "stop",
	Usage: "stop idb-agent",
	Action: func(c *cli.Context) error {
		sockFile := filepath.Join(constant.BaseDir, constant.AgentSock)
		conn, err := net.Dial("unix", sockFile)
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

var RestartCommand = &cli.Command{
	Name:  "restart",
	Usage: "restart idb-agent",
	Action: func(c *cli.Context) error {
		err := exec.Command("systemctl", "restart", constant.AgentService).Run()
		if err != nil {
			return fmt.Errorf("failed to restart service: %w", err)
		}

		return nil
	},
}

var ConfigCommand = &cli.Command{
	Name:  "config",
	Usage: "configure idb-agent",
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

		sockFile := filepath.Join(constant.BaseDir, constant.AgentSock)
		conn, err := net.Dial("unix", sockFile)
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

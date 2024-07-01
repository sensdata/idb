package main

import (
	"log"
	"os"

	"github.com/sensdata/idb/agent/agent"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:  "idbagent",
		Usage: "idb agent command line tools",
		Commands: []cli.Command{
			*agent.StartCommand,
			*agent.StopCommand,
			*agent.ConfigCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

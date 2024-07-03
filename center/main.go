package main

import (
	"fmt"
	"log"
	"os"

	"github.com/sensdata/idb/center/core/command"
	_ "github.com/sensdata/idb/center/docs"
	"github.com/urfave/cli"
)

func main() {
	app := &cli.App{
		Name:  "idbcenter",
		Usage: "idb center command line tools",
		Commands: []cli.Command{
			*command.StartCommand,
			*command.StopCommand,
			*command.RestartCommand,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Center Exited")
}

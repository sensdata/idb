package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	logstream "github.com/sensdata/idb/core/logstream"
	"github.com/sensdata/idb/core/logstream/internal/config"
)

func main() {
	taskID := flag.String("task", "", "Task ID to follow")
	flag.Parse()

	if *taskID == "" {
		fmt.Println("Please provide task ID")
		os.Exit(1)
	}

	cfg := config.DefaultConfig()
	ls, err := logstream.New(cfg)
	if err != nil {
		panic(err)
	}
	defer ls.Close()

	reader, err := ls.GetReader(*taskID)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	logCh, _ := reader.Follow(0, 0)
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case msg, ok := <-logCh:
			if !ok {
				return
			}
			fmt.Print(string(msg))
		case <-sigCh:
			return
		}
	}
}

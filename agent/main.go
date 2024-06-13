package main

import (
	"fmt"
	"log"

	"github.com/byteseek/idb/agent/config"
)

func main() {
	configPath := "config.json"

	manager, err := config.NewManager(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize config manager: %v", err)
	}

	cfg := manager.GetConfig()
	fmt.Println("Current Config:")
	fmt.Printf("%+v\n", *cfg)
}

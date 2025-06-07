package main

import (
	"fmt"
	"log"
	"os"
)

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Check environment variable to skip prompt
	if os.Getenv("RUN_MODE") == "server" {
		runServer(config.Host)
		return
	} else if os.Getenv("RUN_MODE") == "client" {
		runClient(config.Host)
		return
	}

	// Otherwise ask the user
	var mode string
	fmt.Print("Run as server (s) or client (c)? ")
	fmt.Scanln(&mode)

	if mode == "s" {
		runServer(config.Host)
	} else {
		runClient(config.Host)
	}
}
// main.go
package main

import (
	"fmt"
	"log"
)

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	var mode string
	fmt.Print("Run as server (s) or client (c)? ")
	fmt.Scanln(&mode)

	if mode == "s" {
		runServer(config.Host)
	} else {
		runClient(config.Host)
	}
}

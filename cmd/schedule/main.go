package main

import (
	"log"
	"os"
	"schedule/config"
	"schedule/internal/app"
)

// @title Schedule API
// @version 1.0
// @description
// @BasePath /api

func main() {
	configPath := "config/config.yaml"
	if len(os.Args) > 1 {
		configPath = os.Args[1]
	}

	cfg, err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatal("read config error: ", err)
	}

	app.Run(cfg)
}
